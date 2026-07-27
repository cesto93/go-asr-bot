#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="docker-compose.yml"

if [[ ! -f "$COMPOSE_FILE" ]]; then
    echo "Compose file not found: $COMPOSE_FILE"
    exit 1
fi

docker compose pull
docker compose up -d
