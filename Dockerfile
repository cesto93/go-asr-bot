ARG BACKEND=yzma

FROM golang:1.26 AS build-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM build-base AS build-yzma
RUN CGO_ENABLED=0 go build -o /go-asr-bot .

FROM build-base AS build-crispasr
COPY lib/crispasr /src/lib/crispasr/
RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake build-essential git \
    && rm -rf /var/lib/apt/lists/*
RUN if [ ! -f lib/crispasr/CMakeLists.txt ]; then \
      rm -rf lib/crispasr && \
      git clone --depth 1 https://github.com/CrispStrobe/CrispASR lib/crispasr; \
    fi
RUN rm -rf lib/crispasr/build && go generate ./internal/asr/
RUN CGO_ENABLED=1 go build -tags cgo -o /go-asr-bot .

FROM build-${BACKEND} AS build-final

FROM debian:bookworm-slim AS runtime-base
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

FROM runtime-base AS runtime-yzma

FROM runtime-base AS runtime-crispasr
COPY --from=build-crispasr /src/lib/crispasr/build/src/libcrispasr* /usr/local/lib/
ENV LD_LIBRARY_PATH=/usr/local/lib

FROM runtime-${BACKEND}
COPY --from=build-final /go-asr-bot /usr/local/bin/go-asr-bot
COPY scripts/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
