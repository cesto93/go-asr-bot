# go-asr-bot

Go-based Telegram bot using `github.com/go-telegram-bot-api/telegram-bot-api/v5`.

ASR transcription via **CrispASR**: CrispASR C library wrapping 26+ ASR backends (parakeet, canary, whisper, cohere-transcribe, etc.)

## Run

```bash
TELEGRAM_BOT_TOKEN=your_token CGO_ENABLED=1 go run .
```

Model can be overridden with `--model` flag or `ASR_DEFAULT_MODEL` env var. Language can be set with `--lang` flag (ISO 639-1, e.g. `--lang it`).

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

`Language` and `UserID` changes hot-reload via fsnotify.

## Telegram bot commands

| Command | Description |
| ------- | ----------- |
| `/start` | Start the bot |
| `/help` | Show available commands |
| `/status` | Show current model, language, and downloaded models (with "Change model" button) |
| `/models` | Choose ASR model from an inline keyboard |
| `/setmodel [name]` | Set ASR model by name (downloads if needed); no name opens the model menu |
| `/setlang <code>` | Set language (ISO 639-1, empty to clear) |

## CLI commands

| Command | Description |
| ------- | ----------- |
| `go run . [--model <name>] [--lang <code>]` | Run the bot (long-polling) |
| `go run . pull --model <name> [--upgrade] [--model-path <dir>]` | Download a model |
| `go run . list [--model-path <dir>]` | List downloaded models |
| `go run . list --available` | Show models available for download |
| `go run . rm <name> [--model-path <dir>]` | Delete a downloaded model |
| `go run . run <audio-file> [--model <name>] [--lang <code>]` | Transcribe a single file from CLI |

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

The CrispASR C library is pulled as a pre-built binary from GitHub releases. `go generate` extracts it and rebuilds ggml:

```bash
CGO_ENABLED=1 go generate ./internal/asr/
go build .
```

The pre-built tarball is cached in `lib-imported/` — no CMake or submodule needed. Download it first (`libcrispasr-linux-x86_64.tar.gz` or `libcrispasr-linux-arm64.tar.gz`) and place it in `lib-imported/`; the script aborts with the exact URL if it is missing.

The CrispASR version is defined in `scripts/crispasr-version.sh`, which `scripts/build-crispasr.sh` sources. The Dockerfile also has its own default (`ARG CRISPASR_VERSION`) and must be bumped in the same commit.

Since v0.8.44 the release tarball ships `crispasr_session.h` alongside `crispasr.h`. The CGO bridge still declares the session prototypes inline (`internal/asr/crispasr_cgo.go`), so copying headers is not part of the flow.

Note: `crispasr_audio_load` (the C ABI path used by the bot) still resamples with miniaudio's linear interpolator. The polyphase resampler added in v0.8.30 is only wired into the CLI's `read_audio_data()`; `crispasr_audio_resample()` is available to callers that need it.

The build script handles both amd64 and arm64 via the `TARGETARCH` environment variable (defaults to `amd64`).

The ggml submodule commit is resolved automatically at build time from the CrispASR GitHub API (`/git/trees/{tag}`), so updating `CRISPASR_VERSION` alone is sufficient. If the API call fails, the script falls back to the pre-built ggml libraries from the CrispASR tarball (skipping the SIGILL-safe rebuild).

The ggml rebuild itself lives in `scripts/build-ggml.sh` (usage: `build-ggml.sh <output-dir>`; `CRISPASR_VERSION` is taken from the environment when set, otherwise from `scripts/crispasr-version.sh`). `build-crispasr.sh` calls it unless `GGML_PREBUILT_DIR` points to an existing sanitized build, in which case those libraries are used verbatim (the Docker build relies on this).

## Docker

### AVX-512 SIGILL fix

The pre-built `libggml*.so*` files bundled with CrispASR contain AVX-512 instructions that cause SIGILL on CPUs without AVX-512 support (e.g. Intel i7-1355U).

The build script (`scripts/build-crispasr.sh`) works around this by rebuilding ggml from the CrispASR vendored source with `-DGGML_NATIVE=OFF` (AVX-512 defaults to OFF). The resulting .so files replace the pre-built ones before the Go binary is linked.

In the Dockerfile the rebuild happens in its own `ggml-build` stage, which only depends on `scripts/build-ggml.sh`, `scripts/crispasr-version.sh` and `ARG CRISPASR_VERSION`. Go source changes therefore never invalidate the cmake layer. The `build` stage consumes it through `GGML_PREBUILT_DIR=/ggml` and no longer installs cmake.

### Multi-arch CI

A single workflow (`.github/workflows/docker.yml`) builds both `linux/amd64` and `linux/arm64` platforms and publishes a multi-arch manifest. The same tags (`latest`, `<sha>`) work on both architectures — Docker selects the right variant at pull time.

Each platform runs on a native runner (`ubuntu-latest` for amd64, `ubuntu-24.04-arm` for arm64, free for public repos), so no QEMU emulation is involved. The matrix carries `platform`/`runner`/`arch` per entry; cache scopes are per arch. Do not reintroduce `docker/setup-qemu-action` unless a runner is switched back to an x86 host.

### Release images

`.github/workflows/release.yml` publishes versioned images when a GitHub release is published (also runnable via `workflow_dispatch` with a `tag` input). The image is tagged with the release tag minus the leading `v` (e.g. release `v1.2.3` → `ghcr.io/<owner>/<repo>:1.2.3`) plus the release commit SHA; `latest` is only moved for non-prerelease releases. Per-arch staging tags are `<arch>-<version>` and are merged into a multi-arch manifest.

### ARM64

The build script supports `TARGETARCH=arm64` for ARM64 builds. Pre-built ARM64 binaries are downloaded from a separate CrispASR release URL. Build with:

```bash
make docker-build-arm64
# or
docker build --build-arg TARGETARCH=arm64 -t go-asr-bot:latest-arm64 .
```

For native ARM64 builds outside Docker, set `TARGETARCH=arm64`:

```bash
TARGETARCH=arm64 go generate ./internal/asr/
go build .
```

### Docker Compose

```bash
./scripts/install.sh
```

Or manually:

```bash
docker compose pull
docker compose up -d
```

Requires a `.env` file with `TELEGRAM_BOT_TOKEN` set.

Data lives in named volumes, never in the project directory: `asr-data` is mounted at `/opt/go-asr-bot` (config.yaml) and `asr-models` at `/opt/go-asr-bot/models`. Do not bind-mount the repository at `/opt/go-asr-bot`: if that host directory is deleted or replaced (e.g. `rm -rf` + re-clone) while the container keeps running, the bind mount is orphaned and every path below it fails with `ENOENT` (`mkdir /opt/go-asr-bot/models: no such file or directory`). The container then has to be recreated with `docker compose up -d --force-recreate`.

`scripts/install.sh` pulls the published multi-arch image from GHCR (it does not build locally).

## Runtime dependencies

- **libcrispasr.so**, **libggml*.so**, **libopenblas.so.0** — handled by `go generate` or Docker

## Project structure

```
main.go              # entry point, graceful shutdown
cmd/                 # cobra commands (root, pull, list, rm, run)
config/              # env/file config, ModelVariant definitions, HTTP download
internal/asr/        # Engine interface + CrispASR CGO bridge
internal/bot/        # bot struct, long-polling updates, hot-reload
internal/handlers/   # message & command handlers
scripts/             # build scripts, docker-entrypoint, install/uninstall
Makefile             # docker-build, docker-build-arm64, test targets
data/                # test audio files
lib-imported/        # Pre-built CrispASR tarball (generated by go generate)
.github/workflows/   # CI/CD (docker.yml — multi-arch build)
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

- Model download uses `net/http` directly (`http.Get` + `io.Copy`), writing files to `MODELS_DIR/<model-file>/<model-file>`. The file is written to `<name>.part` and renamed on success, so an interrupted download is never mistaken for a complete model.
- `DownloadModelTo(variant, dir)` backs both `DownloadModel` (uses `MODELS_DIR`) and `pull --model-path`.
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

## Telegram UI

- Model selection is graphical: `/models`, `/setmodel` (no args) and the "Change model" button in `/status` all send an inline keyboard built in `internal/handlers/menu.go`.
- Buttons use callback data prefixed with `mdl:` (`mdl:menu`, `mdl:close`, `mdl:set:<variant>`). Inline keyboards are limited to 64 bytes per callback payload, so keep the prefix short.
- Labels mark the active model with `(active)` and models not stored locally with `(download)`; selecting one triggers the download in `Bot.SetModel`.
- Callback queries are dispatched in `Bot.Run` (`internal/bot/bot.go`) with the same `USER_ID` authorization as messages, then handled by `Handler.HandleCallback`.
