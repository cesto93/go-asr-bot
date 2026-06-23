# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

ASR transcription via **CrispASR**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, cohere-transcribe, etc.)

## Run

```bash
TELEGRAM_BOT_TOKEN=your_token CGO_ENABLED=1 go run .
```

Model can be overridden with `--model` flag or `ASR_DEFAULT_MODEL` env var.

## Configuration

Config loaded from YAML (`/opt/go-asr-bot/config.yaml`) + `.env` + environment variables. Env takes precedence.

| Variable             | Description                     | Default |
| -------------------- | ------------------------------- | ------- |
| `TELEGRAM_BOT_TOKEN` | Bot token from BotFather        | (required) |
| `DEBUG`              | Enable debug logging            | `false` |
| `USER_ID`            | Restrict to single user         | `0` (all) |
| `ASR_DEFAULT_MODEL`  | Default model variant           | `parakeet-tdt-0.6b-v3-q4_k` |
| `ASR_LANGUAGE`       | Source language (ISO 639-1)     | (none) |
| `CRISPASR_THREADS`   | CPU threads for CrispASR        | `4` |
| `MODELS_DIR`         | Model storage directory         | `/opt/go-asr-bot/models` |
| `CONFIG_PATH`        | Path to YAML config file        | `/opt/go-asr-bot/config.yaml` |

Model paths are always inferred from the selected model variant (via `--model` flag or `ASR_DEFAULT_MODEL`).

Config can be viewed/modified at runtime via `go run . config --set-*` flags. `Language` and `UserID` changes hot-reload via fsnotify.

## Telegram bot commands

| Command | Description |
|---|---|
| `/start` | Start the bot |
| `/help` | Show available commands |
| `/list` | List downloaded models |
| `/status` | Show current model and language |
| `/setmodel <name>` | Set ASR model (downloads if needed) |
| `/setlang <code>` | Set language (ISO 639-1, empty to clear) |

## CLI commands

| Command | Description |
|---|---|
| `go run .` | Run the bot (long-polling) |
| `go run . pull --model <name>` | Download a model |
| `go run . list` | List downloaded models |
| `go run . list --available` | Show models available for download |
| `go run . rm <name>` | Delete a downloaded model |
| `go run . run <audio-file>` | Transcribe a single file from CLI |
| `go run . config` | Display config |
| `go run . config --set-*` | Modify config settings |

## Pull models

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

### Docker Compose

```bash
docker compose up -d
```

Requires a `.env` file with `TELEGRAM_BOT_TOKEN` set.

## Runtime dependencies

- **ffmpeg** — audio conversion (ogg/wav → PCM f32le)
- **libcrispasr.so**, **libggml*.so**, **libopenblas.so.0** — handled by `go generate` or Docker

## Project structure

```
main.go              # entry point, graceful shutdown
cmd/                 # cobra commands (root, config, pull, list, rm, run)
config/config.go     # env/file-based config with viper + godotenv
config/models.go     # ModelVariant definitions, ResolveModelPath, ModelVariants map
config/download.go   # HTTP model download (net/http)
internal/asr/        # Engine interface + CrispASR CGO bridge
internal/bot/bot.go  # bot struct, long-polling updates loop, hot-reload
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

- Model download uses `net/http` directly (`http.Get` + `io.Copy`), writing files to `MODELS_DIR/<model-file>/<model-file>`.
- If a model directory contains a subdirectory instead of the expected file (e.g. from an older download), `ResolveModelPath` in `config/models.go` drills one level deeper automatically.
- Model paths also support an `mmproj-*` sidecar file (not currently used by CrispASR, but kept for potential future mmproj support).
- The `--model` flag and `/setmodel` command accept variant names (e.g. `parakeet-tdt-0.6b-v3-q8_0`), not raw file paths.

## Conventions

- **No comments** in code unless necessary for non-obvious logic
- **No emojis** in code or commit messages
- Imports grouped: stdlib first, then third-party, then internal
- Error handling: return errors up, log at call site
- Handlers go in `internal/handlers/`, one file per concern if they grow
- Bot lifecycle managed in `internal/bot/bot.go`
