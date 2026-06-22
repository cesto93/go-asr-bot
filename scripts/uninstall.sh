#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="go-asr-bot"
INSTALL_DIR="/opt/${BIN_NAME}"
BIN_PATH="/usr/local/bin/${BIN_NAME}"
SERVICE_NAME="${BIN_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
USER_NAME="${BIN_NAME}"

if [ "$(id -u)" -ne 0 ]; then
	echo "This script must be run as root" >&2
	exit 1
fi

echo "Stopping and disabling ${SERVICE_NAME}..."
systemctl disable --now "${SERVICE_NAME}" 2>/dev/null || true

echo "Removing systemd service file..."
rm -f "${SERVICE_PATH}"

echo "Reloading systemd..."
systemctl daemon-reload

echo "Removing binary..."
rm -f "${BIN_PATH}"

if [ -d "${INSTALL_DIR}" ]; then
	echo "Removing data directory..."
	rm -rf "${INSTALL_DIR}"
fi

echo "Removing system user ${USER_NAME}..."
userdel -r "${USER_NAME}" 2>/dev/null || true

echo ""
echo "Uninstall complete."
