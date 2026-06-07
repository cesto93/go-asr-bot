package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	Debug         bool
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		Debug:         os.Getenv("DEBUG") == "true",
	}
}
