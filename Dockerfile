# Stage 1: Download pre-built CrispASR libraries and package into the
# tarball format that scripts/build-crispasr.sh expects.
FROM golang:1.26 AS crispasr-download
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

ARG CRISPASR_TAG=hf-space-bin
RUN set -eu; \
    url="https://github.com/CrispStrobe/CrispASR/releases/download/${CRISPASR_TAG}/crispasr-bin-linux-x64.tar.gz"; \
    mkdir -p /tmp/hf /tmp/pkg/libcrispasr-linux-x86_64/src /tmp/pkg/libcrispasr-linux-x86_64/ggml/src; \
    curl -sL "$url" -o /tmp/hf.tar.gz; \
    tar xzf /tmp/hf.tar.gz -C /tmp/hf; \
    cp -a /tmp/hf/libcrispasr*.so* /tmp/pkg/libcrispasr-linux-x86_64/src/; \
    cp -a /tmp/hf/libggml*.so* /tmp/pkg/libcrispasr-linux-x86_64/ggml/src/; \
    ln -s libcrispasr.so /tmp/pkg/libcrispasr-linux-x86_64/src/libwhisper.so; \
    mkdir -p lib-imported; \
    tar czf lib-imported/libcrispasr-linux-x86_64.tar.gz -C /tmp/pkg libcrispasr-linux-x86_64; \
    rm -rf /tmp/hf /tmp/hf.tar.gz

# Stage 2: Build ggml from CrispASR's vendored source without AVX-512.
# The pre-built libggml*.so* in the CrispASR tarball contain AVX-512
# instructions that SIGILL on older CPUs (i7-1355U, no AVX-512).
# Building from the same source with GGML_AVX512=OFF (default) produces
# compatible libraries.
FROM golang:1.26 AS ggml-build
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    cmake \
    && rm -rf /var/lib/apt/lists/*

ARG CRISPASR_TAG=hf-space-bin
RUN set -eu; \
    url="https://github.com/CrispStrobe/CrispASR/archive/${CRISPASR_TAG}.tar.gz"; \
    curl -sL "$url" | tar xzf - --strip-components=1 "CrispASR-${CRISPASR_TAG}/ggml"; \
    touch ggml/ggml.pc.in; \
    cmake -B build -S ggml \
        -DBUILD_SHARED_LIBS=ON \
        -DGGML_NATIVE=OFF \
        -DGGML_OPENMP=ON \
        -DGGML_BUILD_TESTS=OFF \
        -DGGML_BUILD_EXAMPLES=OFF; \
    cmake --build build -j "$(nproc)" --target ggml ggml-base ggml-cpu

# Stage 3: Build Go binary via go generate + go build.
FROM golang:1.26 AS build
WORKDIR /src

# The pre-built libcrispasr.so links against libomp (LLVM OpenMP) and may
# link against OpenBLAS; install both so the go-generate script's ldconfig
# check passes instead of attempting a brittle remote download.
RUN apt-get update && apt-get install -y --no-install-recommends \
    libomp5 \
    libopenblas0-pthread \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY config/ config/
COPY main.go .
COPY scripts/ scripts/

COPY --from=crispasr-download /src/lib-imported/ lib-imported/

RUN go generate ./internal/asr/

# Overwrite the AVX-512-laden ggml .so files with our non-AVX-512 build.
# COPY preserves symlinks, so the .so -> .so.0 -> .so.0.10.2 chain stays intact.
COPY --from=ggml-build /src/build/src/libggml*.so* /src/lib/crispasr/build/ggml/src/
COPY --from=ggml-build /src/build/src/libggml*.so* /src/lib/crispasr/build/src/

RUN CGO_ENABLED=1 go build -a -o /go-asr-bot .

# Stage 4: Runtime image
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    ffmpeg \
    libffi8 \
    libgomp1 \
    libomp5 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /go-asr-bot /usr/local/bin/go-asr-bot
COPY --from=build \
    /src/lib/crispasr/build/ggml/src/libggml*.so* \
    /src/lib/crispasr/build/src/libcrispasr*.so* \
    /src/lib/crispasr/build/src/libwhisper.so* \
    /usr/local/lib/
COPY scripts/docker-entrypoint.sh /entrypoint.sh

ENV LD_LIBRARY_PATH=/usr/local/lib

RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
