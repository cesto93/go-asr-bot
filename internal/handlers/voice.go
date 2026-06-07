package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleVoice(msg *tgbotapi.Message) error {
	h.sendTranscribing(msg.Chat.ID)

	tmpDir, err := os.MkdirTemp("", "asr-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	audioPath, err := h.downloadVoice(msg.Voice.FileID, tmpDir)
	if err != nil {
		return fmt.Errorf("download voice: %w", err)
	}

	text, err := h.asr.Transcribe(audioPath)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
	}

	return h.sendText(msg.Chat.ID, text)
}

func (h *Handler) downloadVoice(fileID, destDir string) (string, error) {
	file, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("get file: %w", err)
	}

	url := file.Link(h.bot.Token)
	dest := filepath.Join(destDir, fileID+".ogg")

	if err := downloadFile(url, dest); err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	return dest, nil
}
