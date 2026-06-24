#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "Usage: $0 [--podman|--docker]"
    exit 1
}

RUNTIME=
while [[ $# -gt 0 ]]; do
    case "$1" in
        --podman) RUNTIME=podman ;;
        --docker) RUNTIME=docker ;;
        *) usage ;;
    esac
    shift
done

if [[ -z "$RUNTIME" ]]; then
    usage
fi

ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH_SUFFIX= ;;
    aarch64|arm64) ARCH_SUFFIX=.arm64 ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

COMPOSE_FILE="${RUNTIME}-compose${ARCH_SUFFIX}.yml"

if [[ ! -f "$COMPOSE_FILE" ]]; then
    echo "Compose file not found: $COMPOSE_FILE"
    exit 1
fi

echo "Using $COMPOSE_FILE"

if [[ "$RUNTIME" == "podman" ]]; then
    SERVICE_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
    SERVICE_FILE="$SERVICE_DIR/podman-compose-go-asr-bot.service"

    if [[ -f "$SERVICE_FILE" ]]; then
        systemctl --user disable --now podman-compose-go-asr-bot.service || true
        rm -f "$SERVICE_FILE"
        systemctl --user daemon-reload
        echo "Systemd user service removed."
    fi
fi

"${RUNTIME}-compose" -f "$COMPOSE_FILE" down
