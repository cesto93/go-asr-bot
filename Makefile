BIN    := go-asr-bot
BIN_C  := go-asr-bot-crispasr
MODEL  ?= qwen3-asr-0.6b-q8_0
GO     ?= go

.PHONY: all build build-crispasr pull pull-model crispasr-lib clean test install

all: build

build:
	CGO_ENABLED=0 $(GO) build -o $(BIN) .

build-crispasr:
	CGO_ENABLED=1 $(GO) build -o $(BIN_C) .

pull:
	$(GO) run . pull

pull-model:
	$(GO) run . pull --model $(MODEL)

crispasr-lib:
	cmake -S lib/crispasr -B lib/crispasr/build -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release
	cmake --build lib/crispasr/build --target crispasr-lib -j$$(nproc)

clean:
	rm -f $(BIN) $(BIN_C)

test:
	$(GO) test ./...

install:
	${GO} install ./...
