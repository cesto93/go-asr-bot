GO       ?= go
DOCKER   ?= docker
IMG      ?= go-asr-bot
TAG      ?= latest

.PHONY: all build build-crispasr docker-build docker-build-crispasr docker-up pull pull-model crispasr-lib clean test install

all: build

docker-build:
	$(DOCKER) build -t $(IMG):$(TAG) .

docker-up:
	$(DOCKER) compose up -d

docker-down:
	$(DOCKER) compose down

test:
	$(GO) test ./...

install:
	${GO} install ./...
