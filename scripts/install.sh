#!/usr/bin/env bash
set -euo pipefail

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH_SUFFIX= ;;
    aarch64|arm64) ARCH_SUFFIX=.arm64 ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

COMPOSE_FILE="docker-compose${ARCH_SUFFIX}.yml"

if [[ ! -f "$COMPOSE_FILE" ]]; then
    echo "Compose file not found: $COMPOSE_FILE"
    exit 1
fi

echo "Using $COMPOSE_FILE"
docker compose -f "$COMPOSE_FILE" pull
docker compose -f "$COMPOSE_FILE" up -d
