package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	Debug         bool
	ModelPath     string
	MMProjPath    string
	LibPath       string
	UserID        int64
}

func Load() *Config {
	godotenv.Load()

	userID, _ := strconv.ParseInt(os.Getenv("USER_ID"), 10, 64)

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		Debug:         os.Getenv("DEBUG") == "true",
		ModelPath:     envOrDefault("MODEL_PATH", "/opt/go-asr-bot/models/Qwen3-ASR-0.6B-Q8_0.gguf"),
		MMProjPath:    envOrDefault("MMPROJ_PATH", "/opt/go-asr-bot/models/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf"),
		LibPath:       envOrDefault("YZMA_LIB", "/opt/go-asr-bot/llamacpp"),
		UserID:        userID,
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
