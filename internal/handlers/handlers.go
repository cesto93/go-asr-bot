package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot *tgbotapi.BotAPI
}

func New(bot *tgbotapi.BotAPI) *Handler {
	return &Handler{bot: bot}
}

func (h *Handler) HandleMessage(msg *tgbotapi.Message) error {
	if msg.IsCommand() {
		return h.handleCommand(msg)
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "Echo: "+msg.Text)
	_, err := h.bot.Send(reply)
	return err
}

func (h *Handler) handleCommand(msg *tgbotapi.Message) error {
	switch msg.Command() {
	case "start":
		return h.sendText(msg.Chat.ID, "Hello! I'm a Telegram bot.")
	case "help":
		return h.sendText(msg.Chat.ID, "Available commands:\n/start - Start the bot\n/help - Show this help")
	default:
		return h.sendText(msg.Chat.ID, "Unknown command.")
	}
}

func (h *Handler) sendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}
