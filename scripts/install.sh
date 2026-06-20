#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="go-asr-bot"
INSTALL_DIR="/opt/${BIN_NAME}"
BIN_PATH="/usr/local/bin/${BIN_NAME}"
SERVICE_NAME="${BIN_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
USER_NAME="${BIN_NAME}"

# Resolve repo root from script location
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if [ "$(id -u)" -ne 0 ]; then
	echo "This script must be run as root" >&2
	exit 1
fi

command -v go >/dev/null 2>&1 || { echo "Go is not installed. Install Go first: https://go.dev/dl/" >&2; exit 1; }

# Build CrispASR C library if the submodule and build tools are available
CGO_ENABLED=0
if [ -f "${REPO_DIR}/lib/crispasr/CMakeLists.txt" ]; then
	if command -v cmake >/dev/null 2>&1 && command -v gcc >/dev/null 2>&1; then
		echo "Building CrispASR C library..."
		cmake -S "${REPO_DIR}/lib/crispasr" -B "${REPO_DIR}/lib/crispasr/build" \
			-DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release
		cmake --build "${REPO_DIR}/lib/crispasr/build" --target crispasr -j"$(nproc)"
		CGO_ENABLED=1
	else
		echo "CrispASR submodule found but cmake/gcc missing — building without CrispASR" >&2
	fi
fi

echo "Building ${BIN_NAME} binary..."
CGO_ENABLED=${CGO_ENABLED} go build -o "${BIN_PATH}" "${REPO_DIR}"
chmod 755 "${BIN_PATH}"

echo "Creating user ${USER_NAME}..."
id -u "${USER_NAME}" &>/dev/null || useradd -r -s /usr/sbin/nologin -d "${INSTALL_DIR}" "${USER_NAME}"

mkdir -p "${INSTALL_DIR}"
chown -R "${USER_NAME}:" "${INSTALL_DIR}"

if [ ! -f "${INSTALL_DIR}/.env" ]; then
	echo "Creating ${INSTALL_DIR}/.env..."
	cat > "${INSTALL_DIR}/.env" <<EOF
TELEGRAM_BOT_TOKEN=your_bot_token_here
DEBUG=false
USER_ID=0

# Backend: "yzma" (default) or "crispasr"
ASR_BACKEND=yzma

# Language hint (ISO 639-1, e.g. en, fr, de)
ASR_LANGUAGE=

# yzma backend (Qwen3-ASR)
MODEL_PATH=${INSTALL_DIR}/models/Qwen3-ASR-0.6B-Q8_0.gguf/Qwen3-ASR-0.6B-Q8_0.gguf
MMPROJ_PATH=${INSTALL_DIR}/models/Qwen3-ASR-0.6B-Q8_0.gguf/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf
YZMA_LIB=${INSTALL_DIR}/llamacpp

# CrispASR backend (requires CGO and libcrispasr.so)
# CRISPASR_MODEL_PATH=${INSTALL_DIR}/models/parakeet-tdt-0.6b-v3-q4_k.gguf/parakeet-tdt-0.6b-v3-q4_k.gguf
# CRISPASR_THREADS=4
EOF
	echo ">>> Please edit ${INSTALL_DIR}/.env with your configuration"
fi

echo "Pulling llama.cpp libraries..."
sudo -u "${USER_NAME}" "${BIN_PATH}" pull --upgrade 2>/dev/null || echo "  (skipped — run '${BIN_PATH} pull' manually)"

echo "Pulling default model (qwen3-asr-0.6b-q8_0)..."
sudo -u "${USER_NAME}" "${BIN_PATH}" pull --model qwen3-asr-0.6b-q8_0 --upgrade 2>/dev/null || echo "  (skipped — run '${BIN_PATH} pull --model qwen3-asr-0.6b-q8_0' manually)"

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
echo "Set your TELEGRAM_BOT_TOKEN in ${INSTALL_DIR}/.env, then:"
echo "  systemctl restart ${SERVICE_NAME}"
echo ""
echo "CLI subcommands:"
echo "  ${BIN_NAME}                        Run bot (default)"
echo "  ${BIN_NAME} pull                   Download llama.cpp libraries"
echo "  ${BIN_NAME} pull --model <variant> Download an ASR model"
echo "  ${BIN_NAME} list                   List downloaded models"
echo "  ${BIN_NAME} list --available       List models available for download"
echo "  ${BIN_NAME} run <audio-file>       Transcribe a single file"
echo ""
echo "Service commands:"
echo "  systemctl status ${SERVICE_NAME}"
echo "  journalctl -u ${SERVICE_NAME} -f"
echo ""
echo "To use CrispASR backend, build libcrispasr.so in lib/crispasr/ and set ASR_BACKEND=crispasr in .env"
