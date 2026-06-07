# go-asr-bot

Go-based Telegram bot with ASR transcription using `github.com/hybridgroup/yzma`, llama.cpp, and Qwen3-ASR models.

Built with `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

## Prerequisites

- Go 1.26+
- A Telegram bot token from [@BotFather](https://t.me/BotFather)
- llama.cpp shared libraries in `llamacpp/` (or configured path)

## Install

```bash
go install github.com/pier/go-asr-bot@latest

# Pull llama.cpp libraries
go-asr-bot pull

# Copy and edit the environment file
cp .env.example .env

# Run
go-asr-bot
```

## Quick start

```bash
# Copy and edit the environment file
cp .env.example .env

# Pull llama.cpp libraries
go run . pull

# Run
go run .
```

## Configuration

| Variable            | Description                  | Default |
| ------------------- | ---------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather    | (required) |
| `DEBUG`             | Enable debug logging         | `false` |
| `USER_ID`           | Restrict to specific user    | (unrestricted) |
| `MODEL_PATH`        | Path to Qwen3-ASR GGUF model | `models/Qwen3-ASR-0.6B-Q8_0.gguf` |
| `MMPROJ_PATH`       | Path to multimodal projector  | `models/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf` |
| `YZMA_LIB`          | Directory with llama.cpp .so | `llamacpp` |

Configuration is loaded via environment variables with optional `.env` file support.

## Project structure

```
main.go              # entry point
cmd/root.go          # root command (runs the bot)
cmd/pull.go          # pull command (downloads llama.cpp)
config/config.go     # env-based config
internal/asr/        # ASR engine wrapping yzma/llama.cpp
internal/bot/bot.go  # bot struct, long-polling updates loop
internal/handlers/   # message & command handlers
```

## Conventions

- **No comments** in code unless necessary for non-obvious logic
- **No emojis** in code or commit messages
- Imports grouped: stdlib first, then third-party, then internal
- Error handling: return errors up, log at call site
- Handlers go in `internal/handlers/`, one file per concern if they grow
- Bot lifecycle managed in `internal/bot/bot.go`
