# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

ASR transcription via two backends:
- **yzma** (default): `github.com/hybridgroup/yzma` with llama.cpp and Qwen3-ASR
- **crispasr**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, cohere-transcribe, etc.)

## Run

```bash
# yzma (default)
TELEGRAM_BOT_TOKEN=your_token go run .

# CrispASR (auto-built via go generate when cmake is available)
CGO_ENABLED=1 go run . --model parakeet-tdt-0.6b-v3-q4_k
```

## Configuration

| Variable             | Description                     | Default |
| -------------------- | ------------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather (env takes precedence over config file) | (required) |
| `DEBUG`              | Enable debug logging            | `false` |
| `USER_ID`            | Restrict to single user         | `0` (all) |
| `ASR_DEFAULT_MODEL`  | Default model variant           | `qwen3-asr-0.6b-q8_0` |
| `ASR_LANGUAGE`       | Source language (ISO 639-1)     | (none) |
| `CRISPASR_THREADS`   | CPU threads for CrispASR        | `4` |

Model paths are always inferred from the selected model variant (via `--model` flag or `ASR_DEFAULT_MODEL`).

## Pull models

Model downloads use direct HTTP (not yzma's go-getter), with a live progress bar:

```bash
# yzma models (Qwen3-ASR)
go run . pull --model qwen3-asr-0.6b-q8_0
go run . pull --model qwen3-asr-1.7b-q8_0

# CrispASR models (backend auto-detected from variant name)
go run . pull --model cohere-transcribe-q8_0
go run . pull --model cohere-transcribe-q4_k
go run . pull --model cohere-transcribe-f16
go run . pull --model parakeet-tdt-0.6b-v3-q4_k
go run . pull --model parakeet-tdt-0.6b-v3-q8_0
go run . pull --model parakeet-tdt-0.6b-v3-q5_0
go run . pull --model parakeet-tdt-0.6b-v3

# List downloaded models
go run . list

# llama.cpp libraries (for yzma backend)
go run . pull
```

## Build CrispASR C library

The CrispASR C library is built automatically via `go generate` when cmake is available:

```bash
git submodule add https://github.com/CrispStrobe/CrispASR lib/crispasr
go generate ./internal/asr/
```

Then build the bot with `CGO_ENABLED=1`.

## Project structure

```
main.go              # entry point, graceful shutdown
cmd/                 # cobra commands (root, pull, list, run)
config/config.go     # env-based config
internal/asr/        # Engine interface + backends (yzma, crispasr)
internal/bot/bot.go  # bot struct, long-polling updates loop
internal/handlers/   # message & command handlers
lib/crispasr/        # CrispASR git submodule
```

## Models directory

Each model variant lives in its own subdirectory named after the model `.gguf` file:

```
/opt/go-asr-bot/models/
├── Qwen3-ASR-0.6B-Q8_0.gguf/
│   ├── Qwen3-ASR-0.6B-Q8_0.gguf
│   └── mmproj-Qwen3-ASR-0.6B-Q8_0.gguf
├── cohere-transcribe-q8_0.gguf/
│   └── cohere-transcribe-q8_0.gguf
├── cohere-transcribe-q4_k.gguf/
│   └── cohere-transcribe-q4_k.gguf
├── parakeet-tdt-0.6b-v3-q4_k.gguf/
│   └── parakeet-tdt-0.6b-v3-q4_k.gguf
├── parakeet-tdt-0.6b-v3-q8_0.gguf/
│   └── parakeet-tdt-0.6b-v3-q8_0.gguf
├── parakeet-tdt-0.6b-v3-q5_0.gguf/
│   └── parakeet-tdt-0.6b-v3-q5_0.gguf
└── parakeet-tdt-0.6b-v3.gguf/
    └── parakeet-tdt-0.6b-v3.gguf
```

## Notes

- Model download uses `net/http` directly (not yzma's `download.GetModel`/go-getter), writing files straight to the destination path without extra directory nesting. A terminal progress bar shows percentage and sizes.
- If a model directory contains a subdirectory instead of the expected file (e.g. from an older download), `resolveModelPath` in `cmd/pull.go` drills one level deeper automatically.

## Conventions

- **No comments** in code unless necessary for non-obvious logic
- **No emojis** in code or commit messages
- Imports grouped: stdlib first, then third-party, then internal
- Error handling: return errors up, log at call site
- Handlers go in `internal/handlers/`, one file per concern if they grow
- Bot lifecycle managed in `internal/bot/bot.go`
