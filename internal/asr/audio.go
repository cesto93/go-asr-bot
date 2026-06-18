package asr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
)

func AudioToPCM(audioPath string, sampleRate int) ([]float32, error) {
	pcmPath := audioPath + ".pcm"
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", audioPath,
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", "1",
		pcmPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %s: %w", string(out), err)
	}
	defer os.Remove(pcmPath)

	data, err := os.ReadFile(pcmPath)
	if err != nil {
		return nil, err
	}
	samples := make([]float32, len(data)/4)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &samples); err != nil {
		return nil, err
	}
	return samples, nil
}
