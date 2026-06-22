#!/bin/sh
set -e

model="${ASR_MODEL:-$ASR_DEFAULT_MODEL}"

case "$model" in
    qwen3-asr-*)
        go-asr-bot pull --lib-path /opt/go-asr-bot/llamacpp
        ;;
esac

if [ -n "$model" ]; then
    go-asr-bot pull --model "$model" --model-path /opt/go-asr-bot/models
fi

exec go-asr-bot "$@"
