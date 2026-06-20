package config

import (
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	DefaultYZMALib = "/opt/go-asr-bot/llamacpp"
	ConfigPath     = "/opt/go-asr/config.yaml"
)

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

	v := viper.New()
	v.SetConfigFile(ConfigPath)
	v.SetConfigType("yaml")

	v.SetDefault("debug", false)
	v.SetDefault("user_id", 0)
	v.SetDefault("language", "")
	v.SetDefault("default_model", "qwen3-asr-0.6b-q8_0")
	v.SetDefault("crispasr_threads", 4)

	v.BindEnv("debug", "DEBUG")
	v.BindEnv("user_id", "USER_ID")
	v.BindEnv("language", "ASR_LANGUAGE")
	v.BindEnv("default_model", "ASR_DEFAULT_MODEL")
	v.BindEnv("crispasr_threads", "CRISPASR_THREADS")

	v.ReadInConfig()

	return &Config{
		TelegramToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
		Debug:           v.GetBool("debug"),
		UserID:          v.GetInt64("user_id"),
		Language:        v.GetString("language"),
		DefaultModel:    v.GetString("default_model"),
		CrispasrThreads: v.GetInt("crispasr_threads"),
	}
}

func Save(cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(ConfigPath)
	v.SetConfigType("yaml")

	v.Set("debug", cfg.Debug)
	v.Set("user_id", cfg.UserID)
	v.Set("language", cfg.Language)
	v.Set("default_model", cfg.DefaultModel)
	v.Set("crispasr_threads", cfg.CrispasrThreads)

	os.MkdirAll("/opt/go-asr", 0755)

	return v.WriteConfig()
}

func Watch(callback func(*Config)) {
	v := viper.New()
	v.SetConfigFile(ConfigPath)
	v.SetConfigType("yaml")
	v.WatchConfig()
	v.OnConfigChange(func(_ fsnotify.Event) {
		callback(Load())
	})
}
