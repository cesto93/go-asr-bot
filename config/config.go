package config

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const DataDir = "/opt/go-asr-bot"



func ConfigPath() string {
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}
	return DataDir + "/config.yaml"
}

func ModelsDir() string {
	if d := os.Getenv("MODELS_DIR"); d != "" {
		return d
	}
	return DataDir + "/models"
}

type Config struct {
	TelegramToken string
	Debug         bool
	UserID        int64

	// Language hint for transcription (ISO 639-1, e.g. "en", "fr", "de")
	Language string

	// DefaultModel is the model variant used when no --model flag is given
	DefaultModel string

	CrispasrThreads int
	ModelsDir       string
}

func Load() *Config {
	godotenv.Load()

	v := viper.New()
	v.SetConfigFile(ConfigPath())
	v.SetConfigType("yaml")

	v.SetDefault("debug", false)
	v.SetDefault("user_id", 0)
	v.SetDefault("language", "")
	v.SetDefault("default_model", "parakeet-tdt-0.6b-v3-q8_0")
	v.SetDefault("crispasr_threads", 4)

	v.BindEnv("debug", "DEBUG")
	v.BindEnv("user_id", "USER_ID")
	v.BindEnv("language", "ASR_LANGUAGE")
	v.BindEnv("default_model", "ASR_DEFAULT_MODEL")
	v.BindEnv("crispasr_threads", "CRISPASR_THREADS")
	v.BindEnv("telegram_token", "TELEGRAM_BOT_TOKEN")

	v.ReadInConfig()

	return &Config{
		TelegramToken:   v.GetString("telegram_token"),
		Debug:           v.GetBool("debug"),
		UserID:          v.GetInt64("user_id"),
		Language:        v.GetString("language"),
		DefaultModel:    v.GetString("default_model"),
		CrispasrThreads: v.GetInt("crispasr_threads"),
		ModelsDir:       ModelsDir(),
	}
}

func Save(cfg *Config) error {
	v := viper.New()
	v.SetConfigFile(ConfigPath())
	v.SetConfigType("yaml")

	v.Set("debug", cfg.Debug)
	v.Set("user_id", cfg.UserID)
	v.Set("language", cfg.Language)
	v.Set("default_model", cfg.DefaultModel)
	v.Set("crispasr_threads", cfg.CrispasrThreads)
	v.Set("telegram_token", cfg.TelegramToken)

	if err := os.MkdirAll(filepath.Dir(ConfigPath()), 0775); err != nil {
		return err
	}

	if err := v.WriteConfig(); err != nil {
		return err
	}
	return os.Chmod(ConfigPath(), 0600)
}

func Watch(callback func(*Config)) {
	v := viper.New()
	v.SetConfigFile(ConfigPath())
	v.SetConfigType("yaml")
	v.WatchConfig()
	v.OnConfigChange(func(_ fsnotify.Event) {
		callback(Load())
	})
}
