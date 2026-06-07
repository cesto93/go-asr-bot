# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

ASR transcription using `github.com/hybridgroup/yzma` with llama.cpp and Qwen3-ASR models.

## Run

```bash
TELEGRAM_BOT_TOKEN=your_token go run .
```

## Configuration

| Variable            | Description                  | Default |
| ------------------- | ---------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather    | (required) |
| `DEBUG`             | Enable debug logging         | `false` |
| `MODEL_PATH`        | Path to Qwen3-ASR GGUF model | `models/Qwen3-ASR-0.6B-Q8_0.gguf` |
| `MMPROJ_PATH`       | Path to multimodal projector  | `models/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf` |
| `YZMA_LIB`          | Directory with llama.cpp .so | `llamacpp` |

## Project structure

```
main.go              # entry point, graceful shutdown
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
