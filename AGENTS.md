# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

## Run

```bash
TELEGRAM_BOT_TOKEN=your_token go run .
```

## Project structure

```
main.go              # entry point, graceful shutdown
config/config.go     # env-based config (TELEGRAM_BOT_TOKEN, DEBUG)
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
