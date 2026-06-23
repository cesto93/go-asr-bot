#!/usr/bin/env bash
set -euo pipefail

BIN_NAME="go-asr-bot"
INSTALL_DIR="/opt/${BIN_NAME}"
BIN_PATH="/usr/local/bin/${BIN_NAME}"
SERVICE_NAME="${BIN_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
USER_NAME="${BIN_NAME}"

INSTALL_SERVICE=1
ACCESS_USER="${SUDO_USER:-}"

while [ $# -gt 0 ]; do
	case "$1" in
		--no-service) INSTALL_SERVICE=0; shift ;;
		--user) ACCESS_USER="$2"; shift 2 ;;
		--user=*) ACCESS_USER="${1#*=}"; shift ;;
		*) echo "Unknown option: $1" >&2; exit 1 ;;
	esac
done

# Resolve repo root from script location
REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"

if [ "$(id -u)" -ne 0 ]; then
	echo "This script must be run as root" >&2
	exit 1
fi

command -v go >/dev/null 2>&1 || { echo "Go is not installed. Install Go first: https://go.dev/dl/" >&2; exit 1; }

# Set up CrispASR C library via pre-built tarball + go generate.
CGO_ENABLED=0
TARBALL="${REPO_DIR}/lib-imported/libcrispasr-linux-x86_64.tar.gz"
if [ ! -f "$TARBALL" ] && (command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1); then
	echo "Downloading pre-built CrispASR libraries..."
	mkdir -p "${REPO_DIR}/lib-imported" /tmp/crispasr-hf /tmp/crispasr-pkg/libcrispasr-linux-x86_64/src /tmp/crispasr-pkg/libcrispasr-linux-x86_64/ggml/src
	if command -v curl >/dev/null 2>&1; then
		curl -sL "https://github.com/CrispStrobe/CrispASR/releases/download/hf-space-bin/crispasr-bin-linux-x64.tar.gz" -o /tmp/crispasr-bin.tar.gz
	else
		wget -q "https://github.com/CrispStrobe/CrispASR/releases/download/hf-space-bin/crispasr-bin-linux-x64.tar.gz" -O /tmp/crispasr-bin.tar.gz
	fi
	tar xzf /tmp/crispasr-bin.tar.gz -C /tmp/crispasr-hf
	cp -a /tmp/crispasr-hf/libcrispasr*.so* /tmp/crispasr-pkg/libcrispasr-linux-x86_64/src/
	cp -a /tmp/crispasr-hf/libggml*.so* /tmp/crispasr-pkg/libcrispasr-linux-x86_64/ggml/src/
	ln -s libcrispasr.so /tmp/crispasr-pkg/libcrispasr-linux-x86_64/src/libwhisper.so
	tar czf "$TARBALL" -C /tmp/crispasr-pkg libcrispasr-linux-x86_64
	rm -rf /tmp/crispasr-hf /tmp/crispasr-pkg /tmp/crispasr-bin.tar.gz
fi
if [ -f "$TARBALL" ]; then
	echo "Building CrispASR C library via go generate..."
	if (cd "${REPO_DIR}" && CGO_ENABLED=1 go generate ./internal/asr/); then
		CGO_ENABLED=1
	else
		echo "go generate failed — building without CrispASR" >&2
	fi
else
	echo "CrispASR tarball not available — building without CrispASR" >&2
fi

echo "Building ${BIN_NAME} binary..."
CGO_ENABLED=${CGO_ENABLED} go build -o "${BIN_PATH}" "${REPO_DIR}"
chmod 755 "${BIN_PATH}"

echo "Creating user ${USER_NAME}..."
id -u "${USER_NAME}" &>/dev/null || useradd -r -s /usr/sbin/nologin -d "${INSTALL_DIR}" -g "${USER_NAME}" "${USER_NAME}"

if [ -n "${ACCESS_USER}" ] && [ "${ACCESS_USER}" != "root" ]; then
	echo "Adding user ${ACCESS_USER} to group ${USER_NAME}..."
	usermod -a -G "${USER_NAME}" "${ACCESS_USER}"
fi

mkdir -p "${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}/models"
chown -R "${USER_NAME}:" "${INSTALL_DIR}"
chmod -R g+rw "${INSTALL_DIR}"
chmod o+w "${INSTALL_DIR}/models"
chmod g+s "${INSTALL_DIR}" "${INSTALL_DIR}/models" 2>/dev/null || true

# Grant admin user write access to the data dir so both user and service can
# write config.yaml. Uses ACL when available, falls back to world-writable.
if [ -n "${ACCESS_USER}" ] && [ "${ACCESS_USER}" != "root" ]; then
	if command -v setfacl >/dev/null 2>&1; then
		setfacl -m u:"${ACCESS_USER}":rwx "${INSTALL_DIR}" 2>/dev/null || chmod o+w "${INSTALL_DIR}"
	else
		chmod o+w "${INSTALL_DIR}"
	fi
fi

if [ ! -f "${INSTALL_DIR}/.env" ]; then
	echo "Creating ${INSTALL_DIR}/.env..."
	cat > "${INSTALL_DIR}/.env" <<EOF
TELEGRAM_BOT_TOKEN=your_bot_token_here
DEBUG=false
USER_ID=0

# Language hint (ISO 639-1, e.g. en, fr, de)
ASR_LANGUAGE=

# Default model variant (inferred from --model flag or this value)
# ASR_DEFAULT_MODEL=parakeet-tdt-0.6b-v3-q8_0

# CPU threads for CrispASR (only used with crispasr models)
# CRISPASR_THREADS=4
EOF
	echo ">>> Please edit ${INSTALL_DIR}/.env with your configuration"
fi

echo "Pulling default model (parakeet-tdt-0.6b-v3-q8_0)..."
sudo -u "${USER_NAME}" "${BIN_PATH}" pull --model parakeet-tdt-0.6b-v3-q8_0 --upgrade 2>/dev/null || echo "  (skipped — run '${BIN_PATH} pull --model <variant>' manually)"

chown -R "${USER_NAME}:" "${INSTALL_DIR}"
chmod -R g+rw "${INSTALL_DIR}"
chmod o+w "${INSTALL_DIR}/models"

if [ "${INSTALL_SERVICE}" -eq 1 ]; then
	SERVICE_FILE="${REPO_DIR}/scripts/${SERVICE_NAME}"
	if [ ! -f "${SERVICE_FILE}" ]; then
		echo "Service file not found: ${SERVICE_FILE}" >&2
		exit 1
	fi

	cp "${SERVICE_FILE}" "${SERVICE_PATH}"

	echo "Reloading systemd..."
	systemctl daemon-reload

	echo "Enabling and starting ${SERVICE_NAME}..."
	systemctl enable --now "${SERVICE_NAME}"
fi

echo ""
echo "Installation complete."
echo "  - Binary: ${BIN_PATH}"
echo "  - Data:   ${INSTALL_DIR}"
echo "  - Config: ${INSTALL_DIR}/.env"
if [ "${INSTALL_SERVICE}" -eq 1 ]; then
	echo "  - Service: ${SERVICE_PATH}"
fi
if [ -n "${ACCESS_USER}" ] && [ "${ACCESS_USER}" != "root" ]; then
	echo "  - Group:   ${USER_NAME} (${ACCESS_USER} has write access)"
fi
echo ""
echo "Set your TELEGRAM_BOT_TOKEN in ${INSTALL_DIR}/.env, then:"
if [ "${INSTALL_SERVICE}" -eq 1 ]; then
	echo "  systemctl restart ${SERVICE_NAME}"
else
	echo "  ${BIN_PATH}"
fi
echo ""
if [ "${INSTALL_SERVICE}" -eq 1 ]; then
	echo "Service commands:"
	echo "  systemctl status ${SERVICE_NAME}"
	echo "  journalctl -u ${SERVICE_NAME} -f"
	echo ""
fi
if [ -n "${ACCESS_USER}" ] && [ "${ACCESS_USER}" != "root" ]; then
	echo "NOTE: Log out and back in (or run 'newgrp ${USER_NAME}') for group membership to take effect."
	echo ""
fi
echo "To use the CrispASR backend, pass --model <crispasr-variant> or set ASR_DEFAULT_MODEL in ${INSTALL_DIR}/.env"
