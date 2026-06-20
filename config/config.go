package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const DefaultYZMALib = "/opt/go-asr-bot/llamacpp"

type Config struct {
	TelegramToken string
	Debug         bool
	UserID        int64

	// Language hint for transcription (ISO 639-1, e.g. "en", "fr", "de")
	Language string

	// DefaultModel is the model variant used when no --model flag is given
	DefaultModel string

	CrispasrThreads int
}

func Load() *Config {
	godotenv.Load()

	userID, _ := strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)

	return &Config{
		TelegramToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		Debug:           os.Getenv("DEBUG") == "true",
		UserID:          userID,
		Language:        os.Getenv("ASR_LANGUAGE"),
		DefaultModel:    envOrDefault("ASR_DEFAULT_MODEL", "qwen3-asr-0.6b-q8_0"),
		CrispasrThreads: envOrDefaultInt("CRISPASR_THREADS", 4),
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
