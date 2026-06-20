package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	Debug         bool
	UserID        int64

	// Backend selection
	ASRBackend string // "yzma" (default) or "crispasr"

	// yzma/Qwen3-ASR backend (used when ASRBackend == "yzma")
	ModelPath  string
	MMProjPath string
	YzmaLib    string

	// CrispASR backend (used when ASRBackend == "crispasr")
	CrispasrModelPath string
	CrispasrThreads   int
}

func Load() *Config {
	godotenv.Load()

	userID, _ := strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		Debug:         os.Getenv("DEBUG") == "true",
		UserID:        userID,

		ASRBackend: "yzma",

		ModelPath:  envOrDefault("MODEL_PATH", "/opt/go-asr-bot/models/Qwen3-ASR-0.6B-Q8_0.gguf/Qwen3-ASR-0.6B-Q8_0.gguf"),
		MMProjPath: envOrDefault("MMPROJ_PATH", "/opt/go-asr-bot/models/Qwen3-ASR-0.6B-Q8_0.gguf/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf"),
		YzmaLib:    envOrDefault("YZMA_LIB", "/opt/go-asr-bot/llamacpp"),

		CrispasrModelPath: envOrDefault("CRISPASR_MODEL_PATH", "/opt/go-asr-bot/models/parakeet-tdt-0.6b-v3.gguf/parakeet-tdt-0.6b-v3.gguf"),
		CrispasrThreads:   envOrDefaultInt("CRISPASR_THREADS", 4),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrDefaultInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
