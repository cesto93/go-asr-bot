FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake build-essential git \
    && rm -rf /var/lib/apt/lists/*

COPY . .

RUN if [ ! -f lib/crispasr/CMakeLists.txt ]; then \
      rm -rf lib/crispasr && \
      git clone --depth 1 https://github.com/CrispStrobe/CrispASR lib/crispasr; \
    fi

RUN rm -rf lib/crispasr/build && GIT_ASKPASS=echo go generate ./internal/asr/

RUN go clean -cache && CGO_ENABLED=1 go build -a -o /go-asr-bot .

FROM debian:trixie-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libffi8 \
    libgomp1 \
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
