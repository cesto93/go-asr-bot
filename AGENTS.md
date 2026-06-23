# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

ASR transcription via **CrispASR**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, cohere-transcribe, etc.)

## Run

```bash
CGO_ENABLED=1 go run . --model parakeet-tdt-0.6b-v3-q8_0
```

## Configuration

| Variable             | Description                     | Default |
| -------------------- | ------------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather (env takes precedence over config file) | (required) |
| `DEBUG`              | Enable debug logging            | `false` |
| `USER_ID`            | Restrict to single user         | `0` (all) |
| `ASR_DEFAULT_MODEL`  | Default model variant           | `parakeet-tdt-0.6b-v3-q4_k` |
| `ASR_LANGUAGE`       | Source language (ISO 639-1)     | (none) |
| `CRISPASR_THREADS`   | CPU threads for CrispASR        | `4` |

Model paths are always inferred from the selected model variant (via `--model` flag or `ASR_DEFAULT_MODEL`).

## Pull models

Model downloads use direct HTTP with a live progress bar:

```bash
go run . pull --model cohere-transcribe-q8_0
go run . pull --model cohere-transcribe-q4_k
go run . pull --model cohere-transcribe-f16
go run . pull --model parakeet-tdt-0.6b-v3-q4_k
go run . pull --model parakeet-tdt-0.6b-v3-q8_0
go run . pull --model parakeet-tdt-0.6b-v3-q5_0
go run . pull --model parakeet-tdt-0.6b-v3

# List downloaded models
go run . list
```

## Build CrispASR C library

The CrispASR C library is pulled as a pre-built binary from GitHub releases. `go generate` downloads and extracts it:

```bash
CGO_ENABLED=1 go generate ./internal/asr/
go build .
```

The pre-built tarball is cached in `lib-imported/` — no CMake or submodule needed.

## Docker

### AVX-512 SIGILL fix

The pre-built `libggml*.so*` files bundled with CrispASR contain AVX-512 instructions that cause SIGILL on CPUs without AVX-512 support (e.g. Intel i7-1355U).

The Dockerfile works around this by rebuilding ggml from the CrispASR vendored source (in `ggml-build` stage) with `-DGGML_NATIVE=OFF` (AVX-512 defaults to OFF). The resulting .so files replace the pre-built ones before the Go binary is linked.

## Project structure

```
main.go              # entry point, graceful shutdown
cmd/                 # cobra commands (root, pull, list, run)
config/config.go     # env-based config
internal/asr/        # Engine interface + backends (crispasr)
internal/bot/bot.go  # bot struct, long-polling updates loop
internal/handlers/   # message & command handlers
lib-imported/        # Pre-built CrispASR tarball (downloaded by go generate)
```

## Models directory

Each model variant lives in its own subdirectory named after the model `.gguf` file:

```
/opt/go-asr-bot/models/
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

- Model download uses `net/http` directly, writing files straight to the destination path without extra directory nesting. A terminal progress bar shows percentage and sizes.
- If a model directory contains a subdirectory instead of the expected file (e.g. from an older download), `resolveModelPath` in `cmd/pull.go` drills one level deeper automatically.

## Conventions

- **No comments** in code unless necessary for non-obvious logic
- **No emojis** in code or commit messages
- Imports grouped: stdlib first, then third-party, then internal
- Error handling: return errors up, log at call site
- Handlers go in `internal/handlers/`, one file per concern if they grow
- Bot lifecycle managed in `internal/bot/bot.go`
