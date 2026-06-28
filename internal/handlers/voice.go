package handlers

import (
	"fmt"
	"log/slog"

	"github.com/cesto93/go-asr-bot/internal/asr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleVoice(msg *tgbotapi.Message) error {
	if h.asr == nil {
		slog.Warn("ASR model not available for voice message", "chat", msg.Chat.ID)
		return h.sendText(msg.Chat.ID, "ASR model not available. Please download a model first using `go run . pull --model <name>`.")
	}

	h.sendTranscribing(msg.Chat.ID)

	oggData, err := h.downloadVoiceBytes(msg.Voice.FileID)
	if err != nil {
		return fmt.Errorf("download voice: %w", err)
	}

	pcm, err := asr.AudioToPCMBytes(oggData, h.asr.SampleRate())
	if err != nil {
		return fmt.Errorf("convert audio: %w", err)
	}

	text, err := h.asr.Transcribe(pcm)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
	}

	slog.Info("voice transcribed", "chat", msg.Chat.ID, "text", text)

	return h.sendText(msg.Chat.ID, text)
}

func (h *Handler) downloadVoiceBytes(fileID string) ([]byte, error) {
	file, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}
	return downloadBytes(file.Link(h.bot.Token))
}
