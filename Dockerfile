ARG CRISPASR_VERSION=v0.8.49
ARG TARGETARCH=amd64

# Stage 1: Download pre-built CrispASR libraries and package into the
# tarball format that scripts/build-crispasr.sh expects.
FROM debian:trixie-slim AS crispasr-download
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

ARG CRISPASR_VERSION
ARG TARGETARCH
RUN set -eu; \
    mkdir -p lib-imported; \
    if [ "$TARGETARCH" = "arm64" ]; then \
        url="https://github.com/CrispStrobe/CrispASR/releases/download/${CRISPASR_VERSION}/libcrispasr-linux-arm64.tar.gz"; \
        curl -sL "$url" -o lib-imported/libcrispasr-linux-arm64.tar.gz; \
    else \
        url="https://github.com/CrispStrobe/CrispASR/releases/download/${CRISPASR_VERSION}/libcrispasr-linux-x86_64.tar.gz"; \
        curl -sL "$url" -o lib-imported/libcrispasr-linux-x86_64.tar.gz; \
    fi

# Stage 2: Rebuild ggml from source without CPU-specific optimizations
# (avoids AVX-512/SVE SIGILL). This stage depends only on the build scripts and
# CRISPASR_VERSION, so Go source changes never invalidate the cmake build.
FROM golang:1.26 AS ggml-build
WORKDIR /src

ARG CRISPASR_VERSION

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    cmake \
    curl \
    g++ \
    make \
    && rm -rf /var/lib/apt/lists/*

COPY scripts/crispasr-version.sh scripts/build-ggml.sh scripts/

RUN CRISPASR_VERSION=$CRISPASR_VERSION sh scripts/build-ggml.sh /ggml

# Stage 3: Build Go binary via go generate + go build.
FROM golang:1.26 AS build
WORKDIR /src

ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    libogg0 \
    libomp5 \
    libopenblas0-pthread \
    libopus0 \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY config/ config/
COPY main.go .
COPY scripts/ scripts/

COPY --from=crispasr-download /src/lib-imported/ lib-imported/
COPY --from=ggml-build /ggml/ /ggml/

# Extract CrispASR libraries and link the sanitized ggml build from stage 2.
# Run from internal/asr/ so ../../ relative paths resolve to repo root.
RUN cd internal/asr && TARGETARCH=$TARGETARCH GGML_PREBUILT_DIR=/ggml sh ../../scripts/build-crispasr.sh

RUN CGO_ENABLED=1 go build -a -ldflags="-s -w" -o /go-asr-bot .

# Stage 4: Runtime image
FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libffi8 \
    libgomp1 \
    libogg0 \
    libomp5 \
    libopenblas0-pthread \
    libopus0 \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/go-asr-bot/models

COPY --from=build /go-asr-bot /usr/local/bin/go-asr-bot
COPY --from=build \
    /src/lib/crispasr/build/ggml/src/libggml*.so* \
    /src/lib/crispasr/build/src/libcrispasr*.so* \
    /src/lib/crispasr/build/src/libwhisper.so* \
    /usr/local/lib/
COPY scripts/docker-entrypoint.sh /entrypoint.sh

RUN strip --strip-unneeded /usr/local/lib/*.so* 2>/dev/null || true

ENV LD_LIBRARY_PATH=/usr/local/lib

WORKDIR /opt/go-asr-bot

RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
