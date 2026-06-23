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

echo "CrispASR libraries extracted successfully (arm64)"
