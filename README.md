# go-asr-bot

Telegram bot built with Go.

## Prerequisites

- Go 1.26+
- A Telegram bot token from [@BotFather](https://t.me/BotFather)

## Quick start

```bash
# Set your bot token
export TELEGRAM_BOT_TOKEN=your_bot_token_here

# Run
go run .
```

## Configuration

| Variable            | Description                  | Default |
| ------------------- | ---------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather     | (required) |
| `DEBUG`             | Enable debug logging         | `false` |

## Project structure

- `main.go` — entry point, graceful shutdown
- `config/` — environment-based configuration
- `internal/bot/` — bot lifecycle and update polling
- `internal/handlers/` — message and command handlers
