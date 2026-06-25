#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "Usage: $0 [--podman|--docker]"
    exit 1
}

RUNTIME=
COMPOSE_CMD=
while [[ $# -gt 0 ]]; do
    case "$1" in
        --podman) RUNTIME=podman; COMPOSE_CMD="podman-compose" ;;
        --docker) RUNTIME=docker; COMPOSE_CMD="docker compose" ;;
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
$COMPOSE_CMD -f "$COMPOSE_FILE" pull

if [[ "$RUNTIME" == "podman" ]]; then
    PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
    COMPOSE_ABS="$PROJECT_DIR/$COMPOSE_FILE"
    SERVICE_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
    SERVICE_FILE="$SERVICE_DIR/podman-compose-go-asr-bot.service"
    SRC_SERVICE="$PROJECT_DIR/scripts/podman-compose-go-asr-bot.service"

    mkdir -p "$SERVICE_DIR"
    sed -e "s|__WORKDIR__|$PROJECT_DIR|g" \
        -e "s|__COMPOSE_FILE__|$COMPOSE_ABS|g" \
        "$SRC_SERVICE" > "$SERVICE_FILE"

    systemctl --user daemon-reload
    systemctl --user enable --now podman-compose-go-asr-bot.service
    loginctl enable-linger "$USER" 2>/dev/null || true

    echo "Systemd user service installed and started."
else
    $COMPOSE_CMD -f "$COMPOSE_FILE" up -d
fi
