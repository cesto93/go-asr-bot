# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

ASR transcription via two backends:
- **yzma** (default): `github.com/hybridgroup/yzma` with llama.cpp and Qwen3-ASR
- **crispasr**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, cohere-transcribe, etc.)

## Run

```bash
# yzma (default)
TELEGRAM_BOT_TOKEN=your_token go run .

# CrispASR (requires building libcrispasr.so in lib/crispasr/)
CGO_ENABLED=1 ASR_BACKEND=crispasr go run .

# Or pick a model — backend is auto-detected from the variant
CGO_ENABLED=1 go run . --model parakeet-tdt-0.6b-v3-q4_k
```

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
| `CRISPASR_MODEL_PATH` | Path to any CrispASR GGUF    | `/opt/go-asr-bot/models/parakeet-tdt-0.6b-v3-q4_k.gguf/parakeet-tdt-0.6b-v3-q4_k.gguf` |
| `CRISPASR_THREADS`    | CPU threads                  | `4` |

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

```bash
git submodule add https://github.com/CrispStrobe/CrispASR lib/crispasr
cmake -S lib/crispasr -B lib/crispasr/build -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release
cmake --build lib/crispasr/build --target crispasr -j$(nproc)
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
