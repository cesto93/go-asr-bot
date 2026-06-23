GO       ?= go
DOCKER   ?= docker
IMG      ?= go-asr-bot
TAG      ?= latest

.PHONY: all build build-crispasr docker-build docker-build-arm64 docker-build-crispasr docker-up docker-up-arm64 docker-down docker-down-arm64 pull pull-model crispasr-lib clean test install

all: build

docker-build:
	$(DOCKER) build -t $(IMG):$(TAG) .

docker-build-arm64:
	$(DOCKER) build --build-arg TARGETARCH=arm64 -t $(IMG):$(TAG)-arm64 .

docker-up:
	$(DOCKER) compose up -d

docker-up-arm64:
	$(DOCKER) compose -f docker-compose.yml -f docker-compose.arm64.yml up -d

docker-down:
	$(DOCKER) compose down

docker-down-arm64:
	$(DOCKER) compose -f docker-compose.yml -f docker-compose.arm64.yml down

test:
	$(GO) test ./...
