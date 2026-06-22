BIN    := go-asr-bot
BIN_C  := go-asr-bot-crispasr
MODEL  ?= qwen3-asr-0.6b-q8_0
GO     ?= go

.PHONY: all build build-crispasr pull pull-model crispasr-lib clean test install

all: build

build:
	CGO_ENABLED=0 $(GO) build -o $(BIN) .

build-crispasr:
	$(GO) generate ./internal/asr/
	CGO_ENABLED=1 $(GO) build -o $(BIN_C) .

pull:
	$(GO) run . pull

pull-model:
	$(GO) run . pull --model $(MODEL)

crispasr-lib:
	$(GO) generate ./internal/asr/

clean:
	rm -f $(BIN) $(BIN_C)

test:
	$(GO) test ./...

install:
	${GO} install ./...
