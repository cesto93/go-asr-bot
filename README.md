# go-asr-bot

Go-based Telegram bot for ASR transcription via **CrispASR** — a C library wrapping 26+ ASR backends (parakeet, canary, whisper, cohere-transcribe, etc.).

Built with `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

## Prerequisites

- Go 1.26+
- CGO (`CGO_ENABLED=1`)
- `ffmpeg` at runtime
- A Telegram bot token from [@BotFather](https://t.me/BotFather)

## Quick start

```bash
TELEGRAM_BOT_TOKEN=your_token CGO_ENABLED=1 go run .
```

## Configuration

Config is loaded from YAML (`/opt/go-asr-bot/config.yaml`), `.env`, and environment variables (env takes precedence).

| Variable             | Description                   | Default |
| -------------------- | ----------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather     | (required) |
| `DEBUG`              | Enable debug logging          | `false` |
| `USER_ID`            | Restrict to single user       | `0` (all) |
| `ASR_DEFAULT_MODEL`  | Default model variant         | `parakeet-tdt-0.6b-v3-q4_k` |
| `ASR_LANGUAGE`       | Source language (ISO 639-1)   | (none) |
| `CRISPASR_THREADS`   | CPU threads for CrispASR      | `4` |
| `MODELS_DIR`         | Model storage directory       | `/opt/go-asr-bot/models` |
| `CONFIG_PATH`        | Path to YAML config file      | `/opt/go-asr-bot/config.yaml` |

Model paths are inferred from the variant name. `Language` and `UserID` hot-reload via fsnotify; other changes require restart.

## CLI commands

| Command | Description |
|---|---|
| `go run .` | Run the bot (long-polling) |
| `go run . config` | Display or modify config (`--set-*` flags) |
| `go run . pull --model <name>` | Download a model |
| `go run . list` | List downloaded models |
| `go run . list --available` | Show models available for download |
| `go run . rm <name>` | Delete a downloaded model |
| `go run . run <audio-file>` | Transcribe a single file |

## Telegram bot commands

| Command | Description |
|---|---|
| `/start` | Start the bot |
| `/help` | Show available commands |
| `/list` | List downloaded models |
| `/status` | Show current model and language |
| `/setmodel <name>` | Set ASR model (downloads if needed) |
| `/setlang <code>` | Set language (ISO 639-1, empty to clear) |

## Available models

```bash
# cohere-transcribe
go run . pull --model cohere-transcribe-q8_0
go run . pull --model cohere-transcribe-q4_k
go run . pull --model cohere-transcribe-f16

# parakeet-tdt-0.6b-v3
go run . pull --model parakeet-tdt-0.6b-v3-q4_k  # default
go run . pull --model parakeet-tdt-0.6b-v3-q8_0
go run . pull --model parakeet-tdt-0.6b-v3-q5_0
go run . pull --model parakeet-tdt-0.6b-v3
```

## Docker

```bash
docker compose up -d
```

Requires a `.env` file with `TELEGRAM_BOT_TOKEN`.

The Dockerfile rebuilds ggml from source without AVX-512 to avoid SIGILL on CPUs lacking AVX-512 support (e.g. Intel i7-1355U).

## Build

```bash
CGO_ENABLED=1 go generate ./internal/asr/
CGO_ENABLED=1 go build .
```

The pre-built CrispASR tarball is downloaded by `go generate` and cached in `lib-imported/`.

## Project structure

```
main.go              # entry point, graceful shutdown
cmd/                 # cobra commands (root, config, pull, list, rm, run)
config/              # env/file config, ModelVariant definitions, HTTP download
internal/asr/        # Engine interface + CrispASR CGO bridge
internal/bot/        # bot struct, long-polling updates, hot-reload
internal/handlers/   # message & command handlers
lib-imported/        # Pre-built CrispASR tarball
scripts/             # build-crispasr.sh, docker-entrypoint.sh
```
