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
