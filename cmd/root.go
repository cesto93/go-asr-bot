package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/cesto93/go-asr-bot/internal/bot"
	"github.com/spf13/cobra"
)

var (
	flagModel string
	flagLang  string
)

func init() {
	rootCmd.Flags().StringVar(&flagModel, "model", "", "ASR model name")
	rootCmd.Flags().StringVar(&flagLang, "lang", "", "Source language (ISO 639-1, e.g. en, fr, de)")
}

var rootCmd = &cobra.Command{
	Use:   "go-asr-bot",
	Short: "Telegram bot for ASR transcription",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if flagLang != "" {
			cfg.Language = flagLang
		}

		modelName := flagModel
		if modelName == "" {
			modelName = cfg.DefaultModel
		}
		v, ok := config.ModelVariants[modelName]
		if !ok {
			slog.Error("unknown model", "name", modelName)
			os.Exit(1)
		}

		modelPath := config.ResolveModelPath(v, v.ModelFile)
		var mmprojPath string
		if v.MMProjFile != "" {
			mmprojPath = config.ResolveModelPath(v, v.MMProjFile)
		}
		backend := v.Backend

		if cfg.TelegramToken == "" {
			slog.Error("TELEGRAM_BOT_TOKEN environment variable is not set")
			os.Exit(1)
		}

		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			slog.Info("downloading model", "path", modelPath)
			if err := config.DownloadModel(modelName); err != nil {
				slog.Warn("failed to download model — ASR unavailable", "name", modelName, "err", err)
				modelPath = ""
				mmprojPath = ""
			}
		}

		b, err := bot.New(cfg, modelPath, mmprojPath, modelName, backend)
		if err != nil {
			slog.Warn("failed to create bot, starting without ASR", "err", err)
			b, err = bot.NewWithoutASR(cfg)
			if err != nil {
				slog.Error("failed to create bot even without ASR", "err", err)
				os.Exit(1)
			}
		}

		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			slog.Info("shutting down")
			os.Exit(0)
		}()

		slog.Info("bot started")
		if err := b.Run(); err != nil {
			slog.Error("bot error", "err", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("execute error", "err", err)
		os.Exit(1)
	}
}
