package handlers

import (
	"github.com/cesto93/go-asr-bot/internal/asr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot *tgbotapi.BotAPI
	asr asr.Engine
}

func New(bot *tgbotapi.BotAPI, asr asr.Engine) *Handler {
	return &Handler{bot: bot, asr: asr}
}

func (h *Handler) HandleMessage(msg *tgbotapi.Message) error {
	if msg.IsCommand() {
		return h.handleCommand(msg)
	}

	if msg.Voice != nil {
		return h.handleVoice(msg)
	}

	if msg.Text != "" {
		return h.sendText(msg.Chat.ID, "Echo: "+msg.Text)
	}

	return nil
}

func (h *Handler) handleCommand(msg *tgbotapi.Message) error {
	switch msg.Command() {
	case "start":
		return h.sendText(msg.Chat.ID, "Hello! I'm a Telegram bot. Send me a voice message to transcribe.")
	case "help":
		return h.sendText(msg.Chat.ID, "Available commands:\n/start - Start the bot\n/help - Show this help\n\nSend a voice message to transcribe it.")
	default:
		return h.sendText(msg.Chat.ID, "Unknown command.")
	}
}

func (h *Handler) sendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) sendTranscribing(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Transcribing...")
	_, err := h.bot.Send(msg)
	return err
}
