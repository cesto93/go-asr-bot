package bot

import (
	"log"

	"github.com/pier/go-asr-bot/config"
	"github.com/pier/go-asr-bot/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	handlers *handlers.Handler
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}

	api.Debug = cfg.Debug
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:      api,
		handlers: handlers.New(api),
	}, nil
}

func (b *Bot) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if err := b.handlers.HandleMessage(update.Message); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}

	return nil
}
