GO       ?= go
DOCKER   ?= docker
IMG      ?= go-asr-bot
TAG      ?= latest

.PHONY: all build build-crispasr docker-build docker-build-arm64 pull pull-model crispasr-lib clean test

all: build

docker-build:
	$(DOCKER) build -t $(IMG):$(TAG) .

docker-build-arm64:
	$(DOCKER) build --build-arg TARGETARCH=arm64 -t $(IMG):$(TAG)-arm64 .

test:
	$(GO) test ./...
