package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	Debug         bool
	ModelPath     string
	MMProjPath    string
	LibPath       string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		Debug:         os.Getenv("DEBUG") == "true",
		ModelPath:     envOrDefault("MODEL_PATH", "models/Qwen3-ASR-0.6B-Q8_0.gguf"),
		MMProjPath:    envOrDefault("MMPROJ_PATH", "models/mmproj-Qwen3-ASR-0.6B-Q8_0.gguf"),
		LibPath:       envOrDefault("YZMA_LIB", "llamacpp"),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
