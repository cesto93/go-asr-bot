#!/usr/bin/env sh
set -eu

OUT_DIR="${1:-}"

if [ -z "$OUT_DIR" ]; then
	echo "usage: $0 <output-dir>"
	exit 1
fi

if [ -z "${CRISPASR_VERSION:-}" ]; then
	. "$(dirname "$0")/crispasr-version.sh"
fi

for tool in cmake curl tar; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		echo "ERROR: $tool not found"
		exit 1
	fi
done

echo "Resolving ggml submodule commit from CrispASR ${CRISPASR_VERSION}..."
GGML_COMMIT="$(curl -sL "https://api.github.com/repos/CrispStrobe/CrispASR/git/trees/${CRISPASR_VERSION}" \
	| grep -A3 '"path": "ggml"' | grep '"sha"' | head -1 | cut -d'"' -f4)"

if [ -z "$GGML_COMMIT" ]; then
	echo "ERROR: could not resolve ggml submodule commit from GitHub API."
	exit 1
fi

echo "ggml submodule commit: ${GGML_COMMIT}"

GGML_SRC="$(mktemp -d)"
GGML_BUILD="$(mktemp -d)"
trap 'rm -rf "$GGML_SRC" "$GGML_BUILD"' EXIT

echo "Rebuilding ggml from source (GGML_NATIVE=OFF)..."
curl -sL "https://github.com/CrispStrobe/ggml/archive/${GGML_COMMIT}.tar.gz" \
	| tar xzf - --strip-components=1 -C "$GGML_SRC"

cmake -B "$GGML_BUILD" -S "$GGML_SRC" \
	-DBUILD_SHARED_LIBS=ON \
	-DGGML_NATIVE=OFF \
	-DGGML_OPENMP=ON \
	-DGGML_BUILD_TESTS=OFF \
	-DGGML_BUILD_EXAMPLES=OFF

cmake --build "$GGML_BUILD" -j "$(nproc)" --target ggml ggml-base ggml-cpu

mkdir -p "$OUT_DIR"
cp -a "$GGML_BUILD/src/"libggml*.so* "$OUT_DIR/"

echo "ggml rebuilt successfully in ${OUT_DIR}"
