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
	return bytesToF32(data)
}

func AudioToPCMBytes(data []byte, sampleRate int) ([]float32, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", "pipe:0",
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", "1",
		"pipe:1",
	)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %s: %w", stderr.String(), err)
	}
	return bytesToF32(stdout.Bytes())
}

func bytesToF32(data []byte) ([]float32, error) {
	samples := make([]float32, len(data)/4)
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &samples); err != nil {
		return nil, err
	}
	return samples, nil
}
