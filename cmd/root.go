package cmd

import (
	"log"
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
			log.Fatalf("unknown model %q", modelName)
		}

		modelPath := config.ResolveModelPath(v, v.ModelFile)
		var mmprojPath string
		if v.MMProjFile != "" {
			mmprojPath = config.ResolveModelPath(v, v.MMProjFile)
		}
		backend := v.Backend

		if cfg.TelegramToken == "" {
			log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
		}

		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			log.Printf("WARNING: Model file not found at %s — ASR unavailable", modelPath)
			modelPath = ""
			mmprojPath = ""
		}

		b, err := bot.New(cfg, modelPath, mmprojPath, modelName, backend)
		if err != nil {
			log.Printf("WARNING: Failed to create bot: %v — starting without ASR", err)
			b, err = bot.NewWithoutASR(cfg)
			if err != nil {
				log.Fatalf("Failed to create bot even without ASR: %v", err)
			}
		}

		go func() {
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			log.Println("Shutting down...")
			os.Exit(0)
		}()

		log.Println("Bot started")
		if err := b.Run(); err != nil {
			log.Fatalf("Bot error: %v", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
