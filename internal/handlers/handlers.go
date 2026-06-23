package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/cesto93/go-asr-bot/internal/asr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server interface {
	SetLanguage(lang string) error
	SetModel(name string) error
	CurrentModel() string
	CurrentLanguage() string
}

type Handler struct {
	bot    *tgbotapi.BotAPI
	asr    asr.Engine
	server Server
}

var BotCommands = []tgbotapi.BotCommand{
	{Command: "start", Description: "Start the bot"},
	{Command: "help", Description: "Show available commands"},
	{Command: "list", Description: "List downloaded models"},
	{Command: "status", Description: "Show current model and language"},
	{Command: "setmodel", Description: "Set ASR model (downloads if needed)"},
	{Command: "setlang", Description: "Set language (ISO 639-1, empty to clear)"},
}

func HelpText() string {
	var b strings.Builder
	b.WriteString("Available commands:\n")
	for _, cmd := range BotCommands {
		desc := cmd.Description
		if cmd.Command == "setmodel" {
			desc = "<name> - Set ASR model (downloads if needed)"
		} else if cmd.Command == "setlang" {
			desc = "<code> - Set language (ISO 639-1, empty to clear)"
		}
		b.WriteString("/" + cmd.Command + " " + desc + "\n")
	}
	b.WriteString("\nSend a voice message to transcribe it.")
	return b.String()
}

func New(bot *tgbotapi.BotAPI, asr asr.Engine, server Server) *Handler {
	return &Handler{bot: bot, asr: asr, server: server}
}

func (h *Handler) SetASR(asr asr.Engine) {
	h.asr = asr
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
		return h.sendText(msg.Chat.ID, HelpText())
	case "list":
		return h.handleList(msg)
	case "status":
		return h.handleStatus(msg)
	case "setmodel":
		return h.handleSetModel(msg)
	case "setlang":
		return h.handleSetLang(msg)
	default:
		return h.sendText(msg.Chat.ID, "Unknown command.")
	}
}

func (h *Handler) handleStatus(msg *tgbotapi.Message) error {
	model := h.server.CurrentModel()
	lang := h.server.CurrentLanguage()

	text := fmt.Sprintf("Model: %s\nLanguage: %s", model, lang)
	if lang == "" {
		text = fmt.Sprintf("Model: %s\nLanguage: (not set)", model)
	}
	if h.asr == nil {
		text += "\nASR: unavailable"
	}
	return h.sendText(msg.Chat.ID, text)
}

func (h *Handler) handleSetModel(msg *tgbotapi.Message) error {
	args := strings.Fields(msg.CommandArguments())
	if len(args) == 0 {
		names := make([]string, 0, len(config.ModelVariants))
		for k := range config.ModelVariants {
			names = append(names, k)
		}
		return h.sendText(msg.Chat.ID, "Usage: /setmodel <name>\nAvailable models: "+strings.Join(names, ", "))
	}

	name := args[0]
	if _, ok := config.ModelVariants[name]; !ok {
		return h.sendText(msg.Chat.ID, "Unknown model. Use /setmodel without arguments to list available models.")
	}

	if err := h.server.SetModel(name); err != nil {
		return h.sendText(msg.Chat.ID, fmt.Sprintf("Failed to change model: %v", err))
	}

	return h.sendText(msg.Chat.ID, fmt.Sprintf("Model set to %s", name))
}

func (h *Handler) handleSetLang(msg *tgbotapi.Message) error {
	args := strings.Fields(msg.CommandArguments())

	var lang string
	if len(args) > 0 {
		lang = args[0]
	}

	if err := h.server.SetLanguage(lang); err != nil {
		return h.sendText(msg.Chat.ID, fmt.Sprintf("Failed to set language: %v", err))
	}

	if lang == "" {
		return h.sendText(msg.Chat.ID, "Language cleared.")
	}
	return h.sendText(msg.Chat.ID, fmt.Sprintf("Language set to %s", lang))
}

func findGGUF(dir string) (models []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			models = append(models, findGGUF(path)...)
		} else if strings.HasSuffix(e.Name(), ".gguf") && !strings.HasPrefix(e.Name(), "mmproj-") {
			models = append(models, path)
		}
	}
	return models
}

func modelVariantName(filename string) string {
	for k, v := range config.ModelVariants {
		if v.ModelFile == filename {
			return k
		}
	}
	return ""
}

func (h *Handler) sendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) handleList(msg *tgbotapi.Message) error {
	dir := config.ModelsDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return h.sendText(msg.Chat.ID, fmt.Sprintf("Failed to read models directory: %v", err))
	}

	var models []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		ggufs := findGGUF(filepath.Join(dir, e.Name()))
		for _, p := range ggufs {
			name := modelVariantName(filepath.Base(p))
			if name != "" {
				models = append(models, name)
			}
		}
	}

	sort.Strings(models)

	if len(models) == 0 {
		return h.sendText(msg.Chat.ID, "No downloaded models found.")
	}

	text := "Downloaded models:\n"
	for _, m := range models {
		text += "  • " + m + "\n"
	}

	return h.sendText(msg.Chat.ID, text)
}

func (h *Handler) sendTranscribing(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Transcribing...")
	_, err := h.bot.Send(msg)
	return err
}
