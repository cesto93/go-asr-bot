#!/usr/bin/env sh
set -eu

TARBALL="../../lib-imported/libcrispasr-linux-x86_64.tar.gz"
BUILD_DIR="../../lib/crispasr/build"

if [ ! -f "$TARBALL" ]; then
	echo "ERROR: Pre-built CrispASR tarball not found at $TARBALL"
	echo "Download the artifact and place it in lib-imported/"
	exit 1
fi

mkdir -p "$BUILD_DIR/src" "$BUILD_DIR/ggml/src"

# Extract into a temp directory, then collide everything into src/
# so the linker only needs one -L/-rpath.
TMPDIR="$(mktemp -d)"
tar xzf "$TARBALL" -C "$TMPDIR"

cp -a "$TMPDIR/libcrispasr-linux-x86_64/src/"* "$BUILD_DIR/src/"
cp -a "$TMPDIR/libcrispasr-linux-x86_64/ggml/src/"* "$BUILD_DIR/ggml/src/"

# Flatten: copy ggml .so files alongside libcrispasr.so so a single
# -L/-rpath covers all transitive dependencies at link time.
cp -a "$BUILD_DIR/ggml/src/"*.so* "$BUILD_DIR/src/"

rm -rf "$TMPDIR"

# libcrispasr.so depends on libopenblas.so.0 at link time. If it isn't
# available on the system, pull the .deb and extract the .so locally.
if ! ldconfig -p | grep -qF 'libopenblas.so.0'; then
	OPENBLAS_DIR="$(mktemp -d)"
	# cache file downloaded from https://packages.ubuntu.com/noble/amd64/libopenblas0-pthread/download
	wget -q -O "$OPENBLAS_DIR/pkg.deb" \
		"http://archive.ubuntu.com/ubuntu/pool/universe/o/openblas/libopenblas0-pthread_0.3.26+ds-1ubuntu0.1_amd64.deb"
	dpkg-deb -x "$OPENBLAS_DIR/pkg.deb" "$OPENBLAS_DIR"
	cp -a "$OPENBLAS_DIR/usr/lib/x86_64-linux-gnu/openblas-pthread/"libopenblas*.so* "$BUILD_DIR/src/"
	rm -rf "$OPENBLAS_DIR"
fi

echo "CrispASR libraries extracted successfully"
