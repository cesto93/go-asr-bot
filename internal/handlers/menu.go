package handlers

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/cesto93/go-asr-bot/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	callbackModelMenu   = "mdl:menu"
	callbackModelClose  = "mdl:close"
	callbackModelSelect = "mdl:set:"
)

func (h *Handler) HandleCallback(cq *tgbotapi.CallbackQuery) error {
	switch {
	case cq.Data == callbackModelMenu:
		return h.handleModelMenuCallback(cq)
	case cq.Data == callbackModelClose:
		return h.handleModelCloseCallback(cq)
	case strings.HasPrefix(cq.Data, callbackModelSelect):
		return h.handleModelSelectCallback(cq)
	default:
		return h.answerCallback(cq.ID, "Unknown action")
	}
}

func (h *Handler) showModelMenu(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, h.modelMenuText())
	msg.ReplyMarkup = h.modelKeyboard()
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) handleModelMenuCallback(cq *tgbotapi.CallbackQuery) error {
	if err := h.answerCallback(cq.ID, ""); err != nil {
		slog.Error("failed to answer callback", "err", err)
	}
	return h.editMenu(cq)
}

func (h *Handler) handleModelCloseCallback(cq *tgbotapi.CallbackQuery) error {
	if err := h.answerCallback(cq.ID, ""); err != nil {
		slog.Error("failed to answer callback", "err", err)
	}
	if cq.Message == nil {
		return nil
	}
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, "Model selection closed.")
	_, err := h.bot.Send(edit)
	return err
}

func (h *Handler) handleModelSelectCallback(cq *tgbotapi.CallbackQuery) error {
	name := strings.TrimPrefix(cq.Data, callbackModelSelect)
	if _, ok := config.ModelVariants[name]; !ok {
		return h.answerCallback(cq.ID, "Unknown model.")
	}

	current := h.server.CurrentModel()
	if name == current {
		return h.answerCallback(cq.ID, name+" is already active.")
	}

	if err := h.answerCallback(cq.ID, "Switching to "+name+"..."); err != nil {
		slog.Error("failed to answer callback", "err", err)
	}

	if cq.Message != nil {
		h.editText(cq.Message.Chat.ID, cq.Message.MessageID, "Switching to "+name+"...\nDownloading the model if needed, this can take a while.")
	}

	if err := h.server.SetModel(name); err != nil {
		slog.Error("failed to change model", "model", name, "err", err)
		if cq.Message != nil {
			text := fmt.Sprintf("Failed to switch to %s: %v\n\n%s", name, err, h.modelMenuText())
			h.editMenuWithText(cq, text)
		}
		return nil
	}

	return h.editMenu(cq)
}

func (h *Handler) editMenu(cq *tgbotapi.CallbackQuery) error {
	if cq.Message == nil {
		return nil
	}
	return h.editMenuWithText(cq, h.modelMenuText())
}

func (h *Handler) editMenuWithText(cq *tgbotapi.CallbackQuery, text string) error {
	edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, text)
	kb := h.modelKeyboard()
	edit.ReplyMarkup = &kb
	_, err := h.bot.Send(edit)
	return err
}

func (h *Handler) editText(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	empty := tgbotapi.NewInlineKeyboardMarkup()
	edit.ReplyMarkup = &empty
	if _, err := h.bot.Send(edit); err != nil {
		slog.Error("failed to edit message", "err", err)
	}
}

func (h *Handler) answerCallback(id, text string) error {
	_, err := h.bot.Request(tgbotapi.NewCallback(id, text))
	return err
}

func (h *Handler) modelMenuText() string {
	current := h.server.CurrentModel()
	lang := h.server.CurrentLanguage()
	if lang == "" {
		lang = "not set"
	}

	var b strings.Builder
	b.WriteString("Select an ASR model\n\n")
	b.WriteString("Current: " + current + "\n")
	b.WriteString("Language: " + lang + "\n\n")
	b.WriteString("(active) = in use, (download) = not stored yet and will be downloaded")

	return b.String()
}

func (h *Handler) modelKeyboard() tgbotapi.InlineKeyboardMarkup {
	current := h.server.CurrentModel()
	downloaded := map[string]bool{}
	for _, m := range downloadedModels() {
		downloaded[m] = true
	}

	names := make([]string, 0, len(config.ModelVariants))
	for name := range config.ModelVariants {
		names = append(names, name)
	}
	sort.Strings(names)

	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton
	for _, name := range names {
		label := name
		switch {
		case name == current:
			label += " (active)"
		case !downloaded[name]:
			label += " (download)"
		}

		row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, callbackModelSelect+name))
		if len(row) == 2 {
			rows = append(rows, row)
			row = nil
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Close", callbackModelClose),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func (h *Handler) statusKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Change model", callbackModelMenu),
		),
	)
}
