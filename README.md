# go-asr-bot

Go-based Telegram bot with ASR transcription via two backends:
- **yzma** (default): `github.com/hybridgroup/yzma` with llama.cpp and Qwen3-ASR
- **crispasr**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, etc.)

Built with `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

## Prerequisites

- Go 1.26+
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- For yzma: llama.cpp shared libraries in `/opt/go-asr-bot/llamacpp` (or configured path)
- For CrispASR: `libcrispasr.so` from pre-built release (see [Build CrispASR](#build-crispasr-c-library))

## Quick start

```bash
# yzma backend (default)
TELEGRAM_BOT_TOKEN=your_token go run .

# CrispASR backend
CGO_ENABLED=1 ASR_BACKEND=crispasr go run .
```

## Install (yzma)

```bash
go install github.com/cesto93/go-asr-bot@latest

# Pull llama.cpp libraries
go-asr-bot pull

# Pull a model
go-asr-bot pull --model qwen3-asr-0.6b-q8_0

# Copy and edit the environment file
cp .env.example .env

# Run
go-asr-bot
```

## Install as a service

A systemd service with automatic restarts can be set up via the install script:

```bash
# As root
sudo ./scripts/install.sh
```

This will:
- Build the binary (optionally with CrispASR support if curl/wget are available)
- Create the `go-asr-bot` system user
- Create `/opt/go-asr-bot/` with a `.env` template
- Pull llama.cpp libraries and the default yzma model
- Install and start the systemd service

### Managing the service

```bash
sudo systemctl status go-asr-bot
sudo journalctl -u go-asr-bot -f
sudo systemctl restart go-asr-bot
```

### CrispASR backend

The install script auto-detects if curl/wget are available, downloads the pre-built CrispASR libraries, and runs `go generate` — no manual steps needed:

```bash
sudo ./scripts/install.sh
```

After installation, edit `/opt/go-asr-bot/.env` and set `ASR_BACKEND=crispasr` and the appropriate `CRISPASR_MODEL_PATH`.

### Uninstall

```bash
sudo ./scripts/uninstall.sh
```

The data directory `/opt/go-asr-bot` is preserved — remove it manually with `rm -rf /opt/go-asr-bot` if desired.

## Build CrispASR C library

The CrispASR C library is provided as a pre-built binary from the project's GitHub releases. `go generate` handles downloading and extracting it:

```bash
# The install script handles this automatically, or manually:
export CGO_ENABLED=1
go generate ./internal/asr/
go build .
```

The pre-built tarball is downloaded to `lib-imported/` on first run; no CMake or submodule needed.

## Configuration

| Variable             | Description                   | Default |
| -------------------- | ----------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather     | (required) |
| `DEBUG`              | Enable debug logging          | `false` |
| `USER_ID`            | Restrict to single user       | `0` (all) |
| `ASR_BACKEND`        | Backend: `yzma` or `crispasr` | `yzma` |

### yzma backend

| Variable       | Description                   | Default |
| -------------- | ----------------------------- | ------- |
| `MODEL_PATH`   | Path to Qwen3-ASR GGUF model  | `/opt/go-asr-bot/models/Qwen3-ASR-0.6B-Q8_0.gguf/Qwen3-ASR-0.6B-Q8_0.gguf` |
| `MMPROJ_PATH`  | Path to multimodal projector   | `/opt/go-asr-bot/models/Qwen3-ASR-0.6B-Q8_0.gguf/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf` |
| `YZMA_LIB`     | Directory with llama.cpp .so  | `/opt/go-asr-bot/llamacpp` |

### CrispASR backend

| Variable              | Description                  | Default |
| --------------------- | ---------------------------- | ------- |
| `CRISPASR_MODEL_PATH` | Path to any CrispASR GGUF    | `/opt/go-asr-bot/models/parakeet-tdt-0.6b-v3.gguf/parakeet-tdt-0.6b-v3.gguf` |
| `CRISPASR_THREADS`    | CPU threads                  | `4` |

Configuration is loaded via environment variables with optional `.env` file support.

## Pull models

Model downloads use direct HTTP with a live progress bar:

```bash
# yzma models (Qwen3-ASR)
go run . pull --model qwen3-asr-0.6b-q8_0
go run . pull --model qwen3-asr-1.7b-q8_0

# CrispASR models (use with ASR_BACKEND=crispasr)
go run . pull --model cohere-transcribe-q8_0
go run . pull --model cohere-transcribe-q4_k
go run . pull --model cohere-transcribe-f16

# List downloaded models
go run . list

# llama.cpp libraries (for yzma backend)
go run . pull
```

## Transcribe a single file

```bash
go run . run audio.wav --model qwen3-asr-0.6b-q8_0
```

## Project structure

```
main.go              # entry point, graceful shutdown
cmd/                 # cobra commands (root, pull, list, run)
config/config.go     # env-based config
internal/asr/        # Engine interface + backends (yzma, crispasr)
internal/bot/bot.go  # bot struct, long-polling updates loop
internal/handlers/   # message & command handlers
lib-imported/        # Pre-built CrispASR tarball (generated)
```

