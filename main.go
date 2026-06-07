package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pier/go-asr-bot/config"
	"github.com/pier/go-asr-bot/internal/bot"
)

func main() {
	cfg := config.Load()
	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is not set")
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
}
