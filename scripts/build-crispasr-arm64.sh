#!/usr/bin/env sh
set -eu

TARBALL="../../lib-imported/libcrispasr-linux-arm64.tar.gz"
BUILD_DIR="../../lib/crispasr/build"

if [ ! -f "$TARBALL" ]; then
	echo "ERROR: Pre-built CrispASR tarball not found at $TARBALL"
	echo "Download from: https://github.com/CrispStrobe/CrispASR/releases/download/v0.8.2/libcrispasr-linux-arm64.tar.gz"
	echo "and place it in lib-imported/"
	exit 1
fi

mkdir -p "$BUILD_DIR/src" "$BUILD_DIR/ggml/src"

TMPDIR="$(mktemp -d)"
tar xzf "$TARBALL" -C "$TMPDIR"

cp -a "$TMPDIR/libcrispasr-linux-arm64/src/"* "$BUILD_DIR/src/"
cp -a "$TMPDIR/libcrispasr-linux-arm64/ggml/src/"* "$BUILD_DIR/ggml/src/"

cp -a "$BUILD_DIR/ggml/src/"*.so* "$BUILD_DIR/src/"

rm -rf "$TMPDIR"

if ! ldconfig -p | grep -qF 'libopenblas.so.0'; then
	OPENBLAS_DIR="$(mktemp -d)"
	wget -q -O "$OPENBLAS_DIR/pkg.deb" \
		"http://ports.ubuntu.com/ubuntu-ports/pool/universe/o/openblas/libopenblas0-pthread_0.3.26+ds-1ubuntu0.1_arm64.deb"
	dpkg-deb -x "$OPENBLAS_DIR/pkg.deb" "$OPENBLAS_DIR"
	cp -a "$OPENBLAS_DIR/usr/lib/aarch64-linux-gnu/openblas-pthread/"libopenblas*.so* "$BUILD_DIR/src/"
	rm -rf "$OPENBLAS_DIR"
fi

# Rebuild ggml from source without CPU-specific optimizations (e.g. SVE).
# The pre-built libggml*.so* contain instructions that cause SIGILL on CPUs
# without those extensions (e.g. Raspberry Pi 5 Cortex-A76).
if command -v cmake >/dev/null 2>&1 && command -v curl >/dev/null 2>&1; then
	echo "Rebuilding ggml from source (GGML_NATIVE=OFF)..."
	GGML_SRC="$(mktemp -d)"
	curl -sL "https://github.com/CrispStrobe/CrispASR/archive/refs/tags/v0.8.2.tar.gz" \
		| tar xzf - --strip-components=1 -C "$GGML_SRC" "CrispASR-0.8.2/ggml"
	touch "$GGML_SRC/ggml.pc.in"
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

echo "CrispASR libraries extracted successfully (arm64)"
