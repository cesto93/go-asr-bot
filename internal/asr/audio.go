package asr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

var crispASRLoad func(path string) ([]float32, error)

func AudioToPCM(audioPath string, sampleRate int) ([]float32, error) {
	if crispASRLoad != nil {
		pcm, err := crispASRLoad(audioPath)
		if err == nil {
			return pcm, nil
		}
		slog.Warn("crispasr_audio_load failed, falling back to ffmpeg", "path", audioPath, "err", err)
	}

	return audioToPCMFFmpeg(audioPath, sampleRate)
}

func AudioToPCMBytes(data []byte, sampleRate int) ([]float32, error) {
	if crispASRLoad != nil {
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

		pcm, err := crispASRLoad(tmpPath)
		if err == nil {
			return pcm, nil
		}
		slog.Warn("crispasr_audio_load failed, falling back to ffmpeg", "path", tmpPath, "err", err)
	}

	return audioToPCMFFmpegBytes(data, sampleRate)
}

func audioToPCMFFmpeg(audioPath string, sampleRate int) ([]float32, error) {
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

func audioToPCMFFmpegBytes(data []byte, sampleRate int) ([]float32, error) {
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
