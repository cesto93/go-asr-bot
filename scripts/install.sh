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

cd "$(dirname "$0")"

echo "Building ${BIN_NAME}..."
go build -o "${BIN_NAME}" .

echo "Creating user ${USER_NAME}..."
id -u "${USER_NAME}" &>/dev/null || useradd -r -s /usr/sbin/nologin -d "${INSTALL_DIR}" "${USER_NAME}"

echo "Creating directory ${INSTALL_DIR}..."
mkdir -p "${INSTALL_DIR}"/{models,llamacpp}

echo "Installing binary to ${BIN_PATH}..."
cp "${BIN_NAME}" "${BIN_PATH}"
chmod 755 "${BIN_PATH}"

echo "Copying models..."
cp -r models/* "${INSTALL_DIR}/models/"

echo "Copying llama.cpp libraries..."
cp -r llamacpp/* "${INSTALL_DIR}/llamacpp/"

if [ ! -f "${INSTALL_DIR}/.env" ]; then
	echo "Copying .env.example to ${INSTALL_DIR}/.env..."
	cp .env.example "${INSTALL_DIR}/.env"
	echo ">>> Please edit ${INSTALL_DIR}/.env with your configuration"
fi

chown -R "${USER_NAME}:" "${INSTALL_DIR}"

cat > "${SERVICE_PATH}" <<UNIT
[Unit]
Description=Go ASR Bot - Telegram bot for ASR transcription
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${USER_NAME}
Group=${USER_NAME}
WorkingDirectory=${INSTALL_DIR}
ExecStart=${BIN_PATH}
Restart=always
RestartSec=5
EnvironmentFile=${INSTALL_DIR}/.env

[Install]
WantedBy=multi-user.target
UNIT

echo "Reloading systemd..."
systemctl daemon-reload

echo "Enabling and starting ${SERVICE_NAME}..."
systemctl enable --now "${SERVICE_NAME}"

echo ""
echo "Installation complete."
echo "  - Binary: ${BIN_PATH}"
echo "  - Data:   ${INSTALL_DIR}"
echo "  - Config: ${INSTALL_DIR}/.env"
echo ""
echo "Commands:"
echo "  systemctl status ${SERVICE_NAME}"
echo "  journalctl -u ${SERVICE_NAME} -f"
