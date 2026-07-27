ARG CRISPASR_VERSION=v0.8.4
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

# Stage 2: Build Go binary via go generate + go build.
FROM golang:1.26 AS build
WORKDIR /src

ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
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

# Extract CrispASR libraries and rebuild ggml from source (avoids AVX-512/SVE SIGILL).
# Run from internal/asr/ so ../../ relative paths resolve to repo root.
RUN cd internal/asr && TARGETARCH=$TARGETARCH sh ../../scripts/build-crispasr.sh

RUN CGO_ENABLED=1 go build -a -ldflags="-s -w" -o /go-asr-bot .

# Stage 3: Runtime image
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
