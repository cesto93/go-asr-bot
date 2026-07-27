#!/usr/bin/env sh
set -eu

. "$(dirname "$0")/crispasr-version.sh"

CRISPASR_VERSION_NO_V="${CRISPASR_VERSION#v}"
TARGETARCH="${TARGETARCH:-amd64}"

case "$TARGETARCH" in
    amd64)
        ARCH_SUFFIX="x86_64"
        OPENBLAS_DEB="libopenblas0-pthread_0.3.26+ds-1ubuntu0.1_amd64.deb"
        OPENBLAS_URL="http://archive.ubuntu.com/ubuntu/pool/universe/o/openblas/$OPENBLAS_DEB"
        OPENBLAS_LIBDIR="x86_64-linux-gnu/openblas-pthread"
        ;;
    arm64)
        ARCH_SUFFIX="aarch64"
        OPENBLAS_DEB="libopenblas0-pthread_0.3.26+ds-1ubuntu0.1_arm64.deb"
        OPENBLAS_URL="http://ports.ubuntu.com/ubuntu-ports/pool/universe/o/openblas/$OPENBLAS_DEB"
        OPENBLAS_LIBDIR="aarch64-linux-gnu/openblas-pthread"
        ;;
    *)
        echo "ERROR: unsupported TARGETARCH=$TARGETARCH"
        exit 1
        ;;
esac

TARBALL="../../lib-imported/libcrispasr-linux-${ARCH_SUFFIX}.tar.gz"
BUILD_DIR="../../lib/crispasr/build"

if [ ! -f "$TARBALL" ]; then
	echo "ERROR: Pre-built CrispASR tarball not found at $TARBALL"
	echo "Download from: https://github.com/CrispStrobe/CrispASR/releases/download/${CRISPASR_VERSION}/libcrispasr-linux-${ARCH_SUFFIX}.tar.gz"
	echo "and place it in lib-imported/"
	exit 1
fi

mkdir -p "$BUILD_DIR/src" "$BUILD_DIR/ggml/src"

TMPDIR="$(mktemp -d)"
tar xzf "$TARBALL" -C "$TMPDIR"

cp -a "$TMPDIR/libcrispasr-linux-${ARCH_SUFFIX}/src/"* "$BUILD_DIR/src/"
cp -a "$TMPDIR/libcrispasr-linux-${ARCH_SUFFIX}/ggml/src/"* "$BUILD_DIR/ggml/src/"

cp -a "$BUILD_DIR/ggml/src/"*.so* "$BUILD_DIR/src/"

rm -rf "$TMPDIR"

# libcrispasr.so depends on libopenblas.so.0 at link time. If it isn't
# available on the system, pull the .deb and extract the .so locally.
if ! ldconfig -p | grep -qF 'libopenblas.so.0'; then
	OPENBLAS_DIR="$(mktemp -d)"
	wget -q -O "$OPENBLAS_DIR/pkg.deb" "$OPENBLAS_URL"
	dpkg-deb -x "$OPENBLAS_DIR/pkg.deb" "$OPENBLAS_DIR"
	cp -a "$OPENBLAS_DIR/usr/lib/${OPENBLAS_LIBDIR}/"libopenblas*.so* "$BUILD_DIR/src/"
	rm -rf "$OPENBLAS_DIR"
fi

# Rebuild ggml from source without CPU-specific optimizations (e.g. AVX-512 on
# amd64, SVE on arm64). The pre-built libggml*.so* in the CrispASR tarball
# contain instructions that cause SIGILL on CPUs without those extensions
# (e.g. Intel i7-1355U, Raspberry Pi 5).
if command -v cmake >/dev/null 2>&1 && command -v curl >/dev/null 2>&1; then
	echo "Rebuilding ggml from source (GGML_NATIVE=OFF)..."
	GGML_SRC="$(mktemp -d)"
	curl -sL "https://github.com/CrispStrobe/CrispASR/archive/refs/tags/${CRISPASR_VERSION}.tar.gz" \
		| tar xzf - --strip-components=1 -C "$GGML_SRC" "CrispASR-${CRISPASR_VERSION_NO_V}/ggml"
	touch "$GGML_SRC/ggml/ggml.pc.in"
	GGML_BUILD="$(mktemp -d)"
	cmake -B "$GGML_BUILD" -S "$GGML_SRC/ggml" \
		-DBUILD_SHARED_LIBS=ON \
		-DGGML_NATIVE=OFF \
		-DGGML_OPENMP=ON \
		-DGGML_BUILD_TESTS=OFF \
		-DGGML_BUILD_EXAMPLES=OFF
	cmake --build "$GGML_BUILD" -j "$(nproc)" --target ggml ggml-base ggml-cpu
	cp -a "$GGML_BUILD/src/"libggml*.so* "$BUILD_DIR/ggml/src/"
	cp -a "$GGML_BUILD/src/"libggml*.so* "$BUILD_DIR/src/"
	rm -rf "$GGML_SRC" "$GGML_BUILD"
	echo "ggml rebuilt successfully"
else
	echo "WARNING: cmake or curl not found; using pre-built ggml libraries."
	echo "If you encounter SIGILL, install cmake+curl and re-run, or use Docker."
fi

echo "CrispASR libraries extracted successfully (${TARGETARCH})"
