# Stage 1: Build CrispASR C library.
# Cached independently — only invalidated when lib/crispasr/ sources change.
FROM golang:1.26 AS crispasr-build
WORKDIR /src

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake build-essential git \
    && rm -rf /var/lib/apt/lists/*

COPY lib/crispasr/ lib/crispasr/

RUN if [ ! -f lib/crispasr/CMakeLists.txt ]; then \
      rm -rf lib/crispasr && \
      git clone --depth 1 https://github.com/CrispStrobe/CrispASR lib/crispasr; \
    fi

RUN cmake -S lib/crispasr -B lib/crispasr/build \
    -DBUILD_SHARED_LIBS=ON \
    -DCMAKE_BUILD_TYPE=Release \
    -DCRISPASR_BUILD_TESTS=OFF \
    -DCRISPASR_BUILD_EXAMPLES=OFF \
    -DCRISPASR_BUILD_SERVER=OFF \
    -DCRISPASR_ALL_WARNINGS=OFF \
    -DGGML_ALL_WARNINGS=OFF

RUN cmake --build lib/crispasr/build --target crispasr-lib -j"$(nproc)"

# Stage 2: Build Go binary using the prebuilt CrispASR from stage 1.
FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY config/ config/
COPY main.go .
COPY scripts/ scripts/

COPY --from=crispasr-build /src/lib/crispasr/ lib/crispasr/

RUN CGO_ENABLED=1 go build -a -o /go-asr-bot .

# Stage 3: Runtime image
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
