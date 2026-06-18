# go-asr-bot

Go-based Telegram bot with ASR transcription via two backends:
- **yzma** (default): `github.com/hybridgroup/yzma` with llama.cpp and Qwen3-ASR
- **crispasr**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, etc.)

Built with `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

## Prerequisites

- Go 1.26+
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- For yzma: llama.cpp shared libraries in `llamacpp/` (or configured path)
- For CrispASR: `libcrispasr.so` built from source (see [Build CrispASR](#build-crispasr-c-library))

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

# Copy and edit the environment file
cp .env.example .env

# Run
go-asr-bot
```

## Build CrispASR C library

```bash
git submodule add https://github.com/CrispStrobe/CrispASR lib/crispasr
cmake -S lib/crispasr -B lib/crispasr/build -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release
cmake --build lib/crispasr/build --target crispasr -j$(nproc)
```

Then build the bot with `CGO_ENABLED=1`.

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
| `MODEL_PATH`   | Path to Qwen3-ASR GGUF model  | `models/Qwen3-ASR-0.6B-Q8_0.gguf` |
| `MMPROJ_PATH`  | Path to multimodal projector   | `models/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf` |
| `YZMA_LIB`     | Directory with llama.cpp .so  | `llamacpp` |

### CrispASR backend

| Variable              | Description                  | Default |
| --------------------- | ---------------------------- | ------- |
| `CRISPASR_MODEL_PATH` | Path to any CrispASR GGUF    | `/opt/go-asr-bot/models/parakeet-tdt-0.6b-v3.gguf` |
| `CRISPASR_THREADS`    | CPU threads                  | `4` |

Configuration is loaded via environment variables with optional `.env` file support.

## Project structure

```
main.go              # entry point
cmd/                 # cobra commands (root, pull)
config/config.go     # env-based config
internal/asr/        # Engine interface + backends (yzma, crispasr)
internal/bot/bot.go  # bot struct, long-polling updates loop
internal/handlers/   # message & command handlers
lib/crispasr/        # CrispASR git submodule
```

## Conventions

- **No comments** in code unless necessary for non-obvious logic
- **No emojis** in code or commit messages
- Imports grouped: stdlib first, then third-party, then internal
- Error handling: return errors up, log at call site
- Handlers go in `internal/handlers/`, one file per concern if they grow
- Bot lifecycle managed in `internal/bot/bot.go`
