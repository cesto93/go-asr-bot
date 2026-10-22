#!/usr/bin/env sh
set -eu

. "$(dirname "$0")/crispasr-version.sh"

TARGETARCH="${TARGETARCH:-amd64}"

case "$TARGETARCH" in
    amd64)
        ARCH_TAR="x86_64"
        ARCH_TRIPLE="x86_64-linux-gnu"
        OPENBLAS_DEB="libopenblas0-pthread_0.3.26+ds-1ubuntu0.1_amd64.deb"
        OPENBLAS_URL="http://archive.ubuntu.com/ubuntu/pool/universe/o/openblas/$OPENBLAS_DEB"
        ;;
    arm64)
        ARCH_TAR="arm64"
        ARCH_TRIPLE="aarch64-linux-gnu"
        OPENBLAS_DEB="libopenblas0-pthread_0.3.26+ds-1ubuntu0.1_arm64.deb"
        OPENBLAS_URL="http://ports.ubuntu.com/ubuntu-ports/pool/universe/o/openblas/$OPENBLAS_DEB"
        ;;
    *)
        echo "ERROR: unsupported TARGETARCH=$TARGETARCH"
        exit 1
        ;;
esac

TARBALL="../../lib-imported/libcrispasr-linux-${ARCH_TAR}.tar.gz"
BUILD_DIR="../../lib/crispasr/build"

if [ ! -f "$TARBALL" ]; then
	echo "ERROR: Pre-built CrispASR tarball not found at $TARBALL"
	echo "Download from: https://github.com/CrispStrobe/CrispASR/releases/download/${CRISPASR_VERSION}/libcrispasr-linux-${ARCH_TAR}.tar.gz"
	echo "and place it in lib-imported/"
	exit 1
fi

mkdir -p "$BUILD_DIR/src" "$BUILD_DIR/ggml/src"

TMPDIR="$(mktemp -d)"
tar xzf "$TARBALL" -C "$TMPDIR"

cp -a "$TMPDIR/libcrispasr-linux-${ARCH_TAR}/lib/"* "$BUILD_DIR/src/"
cp -a "$BUILD_DIR/src/"libggml*.so* "$BUILD_DIR/ggml/src/"

rm -rf "$TMPDIR"

# libcrispasr.so depends on libopenblas.so.0 at link time. If it isn't
# available on the system, pull the .deb and extract the .so locally.
if ! ldconfig -p | grep -qF 'libopenblas.so.0'; then
	OPENBLAS_DIR="$(mktemp -d)"
	wget -q -O "$OPENBLAS_DIR/pkg.deb" "$OPENBLAS_URL"
	dpkg-deb -x "$OPENBLAS_DIR/pkg.deb" "$OPENBLAS_DIR"
	cp -a "$OPENBLAS_DIR/usr/lib/${ARCH_TRIPLE}/openblas-pthread/"libopenblas*.so* "$BUILD_DIR/src/"
	rm -rf "$OPENBLAS_DIR"
fi

# Rebuild ggml from source without CPU-specific optimizations (e.g. AVX-512 on
# amd64, SVE on arm64). The pre-built libggml*.so* in the CrispASR tarball
# contain instructions that cause SIGILL on CPUs without those extensions
# (e.g. Intel i7-1355U, Raspberry Pi 5).
#
# GGML_PREBUILT_DIR points to a directory already containing the sanitized
# libraries (built by scripts/build-ggml.sh). The Docker build sets it so the
# expensive cmake step lives in its own cacheable stage.
if [ -n "${GGML_PREBUILT_DIR:-}" ]; then
	if ! ls "$GGML_PREBUILT_DIR/"libggml*.so* >/dev/null 2>&1; then
		echo "ERROR: GGML_PREBUILT_DIR=$GGML_PREBUILT_DIR does not contain libggml*.so*"
		exit 1
	fi
	echo "Using pre-built ggml libraries from ${GGML_PREBUILT_DIR}..."
	cp -a "$GGML_PREBUILT_DIR/"libggml*.so* "$BUILD_DIR/ggml/src/"
	cp -a "$GGML_PREBUILT_DIR/"libggml*.so* "$BUILD_DIR/src/"
elif command -v cmake >/dev/null 2>&1 && command -v curl >/dev/null 2>&1; then
	if sh "$(dirname "$0")/build-ggml.sh" "$BUILD_DIR/ggml/src"; then
		cp -a "$BUILD_DIR/ggml/src/"libggml*.so* "$BUILD_DIR/src/"
	else
		echo "WARNING: ggml rebuild failed."
		echo "Using pre-built ggml libraries from the CrispASR tarball."
	fi
else
	echo "WARNING: cmake or curl not found; using pre-built ggml libraries."
	echo "If you encounter SIGILL, install cmake+curl and re-run, or use Docker."
fi

echo "CrispASR libraries extracted successfully (${TARGETARCH})"
