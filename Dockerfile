ARG TARGETARCH=amd64

# Stage 1: Download pre-built CrispASR libraries and package into the
# tarball format that scripts/build-crispasr.sh expects.
FROM debian:trixie-slim AS crispasr-download
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

ARG TARGETARCH
RUN set -eu; \
    if [ "$TARGETARCH" = "arm64" ]; then \
        url="https://github.com/CrispStrobe/CrispASR/releases/download/v0.8.2/libcrispasr-linux-arm64.tar.gz"; \
        mkdir -p lib-imported; \
        curl -sL "$url" -o lib-imported/libcrispasr-linux-arm64.tar.gz; \
    else \
        url="https://github.com/CrispStrobe/CrispASR/releases/download/hf-space-bin/crispasr-bin-linux-x64.tar.gz"; \
        mkdir -p /tmp/hf /tmp/pkg/libcrispasr-linux-x86_64/src /tmp/pkg/libcrispasr-linux-x86_64/ggml/src; \
        curl -sL "$url" -o /tmp/hf.tar.gz; \
        tar xzf /tmp/hf.tar.gz -C /tmp/hf; \
        cp -a /tmp/hf/libcrispasr*.so* /tmp/pkg/libcrispasr-linux-x86_64/src/; \
        cp -a /tmp/hf/libggml*.so* /tmp/pkg/libcrispasr-linux-x86_64/ggml/src/; \
        ln -s libcrispasr.so /tmp/pkg/libcrispasr-linux-x86_64/src/libwhisper.so; \
        mkdir -p lib-imported; \
        tar czf lib-imported/libcrispasr-linux-x86_64.tar.gz -C /tmp/pkg libcrispasr-linux-x86_64; \
        rm -rf /tmp/hf /tmp/hf.tar.gz; \
    fi

# Stage 2: Build Go binary via go generate + go build.
FROM golang:1.26 AS build
WORKDIR /src

ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
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

# Extract CrispASR libraries (arch-specific script).
# Run from internal/asr/ so ../../ relative paths resolve to repo root.
RUN if [ "$TARGETARCH" = "arm64" ]; then \
        cd internal/asr && sh ../../scripts/build-crispasr-arm64.sh; \
    else \
        cd internal/asr && sh ../../scripts/build-crispasr.sh; \
    fi

# Rebuild ggml from source without CPU-specific optimizations (e.g. AVX-512 on
# amd64, SVE on arm64). The pre-built libggml*.so* in the CrispASR tarball
# contain instructions that cause SIGILL on CPUs without those extensions
# (e.g. Intel i7-1355U, Raspberry Pi 5).
RUN set -eu; \
    if [ "$TARGETARCH" = "arm64" ]; then \
        url="https://github.com/CrispStrobe/CrispASR/archive/refs/tags/v0.8.2.tar.gz"; \
        member="CrispASR-0.8.2/ggml"; \
    else \
        url="https://github.com/CrispStrobe/CrispASR/archive/hf-space-bin.tar.gz"; \
        member="CrispASR-hf-space-bin/ggml"; \
    fi; \
    curl -sL "$url" | tar xzf - --strip-components=1 "$member"; \
    touch ggml/ggml.pc.in; \
    cmake -B ggml-build -S ggml \
        -DBUILD_SHARED_LIBS=ON \
        -DGGML_NATIVE=OFF \
        -DGGML_OPENMP=ON \
        -DGGML_BUILD_TESTS=OFF \
        -DGGML_BUILD_EXAMPLES=OFF; \
    cmake --build ggml-build -j "$(nproc)" --target ggml ggml-base ggml-cpu; \
    cp -a ggml-build/src/libggml*.so* lib/crispasr/build/ggml/src/; \
    cp -a ggml-build/src/libggml*.so* lib/crispasr/build/src/

RUN CGO_ENABLED=1 go build -a -ldflags="-s -w" -o /go-asr-bot .

# Stage 3: Runtime image
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    ffmpeg \
    libffi8 \
    libgomp1 \
    libomp5 \
    libopenblas0-pthread \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /go-asr-bot /usr/local/bin/go-asr-bot
COPY --from=build \
    /src/lib/crispasr/build/ggml/src/libggml*.so* \
    /src/lib/crispasr/build/src/libcrispasr*.so* \
    /src/lib/crispasr/build/src/libwhisper.so* \
    /usr/local/lib/
COPY scripts/docker-entrypoint.sh /entrypoint.sh

RUN strip --strip-unneeded /usr/local/lib/*.so* 2>/dev/null || true

ENV LD_LIBRARY_PATH=/usr/local/lib

RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
