package handlers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cesto93/go-asr-bot/internal/asr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleVoice(msg *tgbotapi.Message) error {
	if h.asr == nil {
		log.Printf("ASR model not available for voice message from chat %d", msg.Chat.ID)
		return h.sendText(msg.Chat.ID, "ASR model not available. Please download a model first using `go run . pull --model <name>`.")
	}

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

	pcm, err := asr.AudioToPCM(audioPath, h.asr.SampleRate())
	if err != nil {
		return fmt.Errorf("convert audio: %w", err)
	}

	text, err := h.asr.Transcribe(pcm)
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
