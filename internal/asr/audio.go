package asr

import (
	"fmt"
	"os"
)

var crispASRLoad func(path string) ([]float32, error)

func AudioToPCM(audioPath string, _ int) ([]float32, error) {
	return crispASRLoad(audioPath)
}

func AudioToPCMBytes(data []byte, _ int) ([]float32, error) {
	f, err := os.CreateTemp("", "asr-*.opus")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	f.Close()
	defer os.Remove(tmpPath)

	return crispASRLoad(tmpPath)
}
