package config

import (
	"os"
	"path/filepath"
)

type ModelVariant struct {
	ModelFile  string
	MMProjFile string
	BaseURL    string
	Backend    string
}

var ModelVariants = map[string]ModelVariant{
	"qwen3-asr-0.6b-q8_0": {
		ModelFile:  "Qwen3-ASR-0.6B-Q8_0.gguf",
		MMProjFile: "mmproj-Qwen3-ASR-0.6B-Q8_0.gguf",
		BaseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-0.6B-GGUF/resolve/main",
		Backend:    "yzma",
	},
	"qwen3-asr-0.6b-bf16": {
		ModelFile:  "Qwen3-ASR-0.6B-bf16.gguf",
		MMProjFile: "mmproj-Qwen3-ASR-0.6B-bf16.gguf",
		BaseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-0.6B-GGUF/resolve/main",
		Backend:    "yzma",
	},
	"qwen3-asr-1.7b-q8_0": {
		ModelFile:  "Qwen3-ASR-1.7B-Q8_0.gguf",
		MMProjFile: "mmproj-Qwen3-ASR-1.7B-Q8_0.gguf",
		BaseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-1.7B-GGUF/resolve/main",
		Backend:    "yzma",
	},
	"qwen3-asr-1.7b-bf16": {
		ModelFile:  "Qwen3-ASR-1.7B-bf16.gguf",
		MMProjFile: "mmproj-Qwen3-ASR-1.7B-bf16.gguf",
		BaseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-1.7B-GGUF/resolve/main",
		Backend:    "yzma",
	},
	"cohere-transcribe-f16": {
		ModelFile: "cohere-transcribe.gguf",
		BaseURL:   "https://huggingface.co/cstr/cohere-transcribe-03-2026-GGUF/resolve/main",
		Backend:   "crispasr",
	},
	"cohere-transcribe-q8_0": {
		ModelFile: "cohere-transcribe-q8_0.gguf",
		BaseURL:   "https://huggingface.co/cstr/cohere-transcribe-03-2026-GGUF/resolve/main",
		Backend:   "crispasr",
	},
	"cohere-transcribe-q4_k": {
		ModelFile: "cohere-transcribe-q4_k.gguf",
		BaseURL:   "https://huggingface.co/cstr/cohere-transcribe-03-2026-GGUF/resolve/main",
		Backend:   "crispasr",
	},
	"parakeet-tdt-0.6b-v3": {
		ModelFile: "parakeet-tdt-0.6b-v3.gguf",
		BaseURL:   "https://huggingface.co/cstr/parakeet-tdt-0.6b-v3-GGUF/resolve/main",
		Backend:   "crispasr",
	},
	"parakeet-tdt-0.6b-v3-q8_0": {
		ModelFile: "parakeet-tdt-0.6b-v3-q8_0.gguf",
		BaseURL:   "https://huggingface.co/cstr/parakeet-tdt-0.6b-v3-GGUF/resolve/main",
		Backend:   "crispasr",
	},
	"parakeet-tdt-0.6b-v3-q5_0": {
		ModelFile: "parakeet-tdt-0.6b-v3-q5_0.gguf",
		BaseURL:   "https://huggingface.co/cstr/parakeet-tdt-0.6b-v3-GGUF/resolve/main",
		Backend:   "crispasr",
	},
	"parakeet-tdt-0.6b-v3-q4_k": {
		ModelFile: "parakeet-tdt-0.6b-v3-q4_k.gguf",
		BaseURL:   "https://huggingface.co/cstr/parakeet-tdt-0.6b-v3-GGUF/resolve/main",
		Backend:   "crispasr",
	},
}

func ResolveModelPath(v ModelVariant, filename string) string {
	path := filepath.Join(ModelsDir(), v.ModelFile, filename)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		path = filepath.Join(path, filename)
	}
	return path
}
