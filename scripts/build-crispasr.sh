#!/usr/bin/env sh
set -eu

command -v cmake >/dev/null 2>&1 || {
    echo 'cmake not found - skipping CrispASR build'
    exit 0
}

LAUNCHER_ARGS=""
CCACHE="$(command -v ccache 2>/dev/null || true)"
if [ -n "$CCACHE" ]; then
    LAUNCHER_ARGS="-DCMAKE_C_COMPILER_LAUNCHER=$CCACHE -DCMAKE_CXX_COMPILER_LAUNCHER=$CCACHE"
fi

cmake -S ../../lib/crispasr -B ../../lib/crispasr/build \
    -DBUILD_SHARED_LIBS=ON \
    -DCMAKE_BUILD_TYPE=Release \
    -DCRISPASR_BUILD_TESTS=OFF \
    -DCRISPASR_BUILD_EXAMPLES=OFF \
    -DCRISPASR_BUILD_SERVER=OFF \
    -DCRISPASR_ALL_WARNINGS=OFF \
    -DGGML_ALL_WARNINGS=OFF \
    $LAUNCHER_ARGS

cmake --build ../../lib/crispasr/build --target crispasr-lib -j"$(nproc)"
