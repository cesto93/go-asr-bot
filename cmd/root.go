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

const modelsDir = "/opt/go-asr-bot/models"

var flagModel string

func init() {
	rootCmd.Flags().StringVar(&flagModel, "model", "", "ASR model name (one of: qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16, qwen3-asr-1.7b-q8_0, qwen3-asr-1.7b-bf16, cohere-transcribe-f16, cohere-transcribe-q8_0, cohere-transcribe-q4_k)")
}

var rootCmd = &cobra.Command{
	Use:   "go-asr-bot",
	Short: "Telegram bot for ASR transcription",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()

		if flagModel != "" {
			v, ok := modelVariants[flagModel]
			if !ok {
				log.Fatalf("unknown model %q (available: qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16, qwen3-asr-1.7b-q8_0, qwen3-asr-1.7b-bf16, cohere-transcribe-f16, cohere-transcribe-q8_0, cohere-transcribe-q4_k)", flagModel)
			}

			cfg.ASRBackend = v.backend

			switch cfg.ASRBackend {
			case "yzma":
				cfg.ModelPath = resolveModelPath(v, v.modelFile)
				if v.mmprojFile != "" {
					cfg.MMProjPath = resolveModelPath(v, v.mmprojFile)
				}
			case "crispasr":
				cfg.CrispasrModelPath = resolveModelPath(v, v.modelFile)
			}
		}
		if cfg.TelegramToken == "" {
			log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
		}

		if cfg.ASRBackend == "yzma" {
			if _, err := os.Stat(cfg.ModelPath); os.IsNotExist(err) {
				log.Fatalf("Model file not found at %s", cfg.ModelPath)
			}
			if _, err := os.Stat(cfg.MMProjPath); os.IsNotExist(err) {
				log.Fatalf("Multimodal projector file not found at %s", cfg.MMProjPath)
			}
		}
		if cfg.ASRBackend == "crispasr" {
			if _, err := os.Stat(cfg.CrispasrModelPath); os.IsNotExist(err) {
				log.Fatalf("CrispASR model file not found at %s", cfg.CrispasrModelPath)
			}
		}

		b, err := bot.New(cfg)
		if err != nil {
			log.Fatalf("Failed to create bot: %v", err)
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
