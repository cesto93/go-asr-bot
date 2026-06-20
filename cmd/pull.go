package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hybridgroup/yzma/pkg/download"
	"github.com/spf13/cobra"
)

type progressReader struct {
	r        io.Reader
	total    int64
	current  int64
	label    string
	lastPct  int
	barWidth int
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.current += int64(n)
	pct := int(pr.current * 100 / pr.total)
	if pct > 100 {
		pct = 100
	}
	if pct != pr.lastPct {
		pr.lastPct = pct
		pr.print()
	}
	return n, err
}

func (pr *progressReader) print() {
	cur := humanSize(pr.current)
	tot := humanSize(pr.total)
	filled := pr.barWidth * pr.lastPct / 100
	bar := strings.Repeat("=", filled)
	if filled < pr.barWidth {
		bar += ">" + strings.Repeat(" ", pr.barWidth-filled-1)
	} else {
		bar += "="
	}
	fmt.Fprintf(os.Stderr, "\r  %s [%s] %3d%% (%s/%s)", pr.label, bar, pr.lastPct, cur, tot)
}



var (
	pullLibPath     string
	pullProcessor   string
	pullVersion     string
	pullUpgrade     bool
	pullModel       string
	pullModelPath   string
)

type modelVariant struct {
	modelFile  string
	mmprojFile string
	baseURL    string
}

var modelVariants = map[string]modelVariant{
	"qwen3-asr-0.6b-q8_0": {
		modelFile:  "Qwen3-ASR-0.6B-Q8_0.gguf",
		mmprojFile: "mmproj-Qwen3-ASR-0.6B-Q8_0.gguf",
		baseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-0.6B-GGUF/resolve/main",
	},
	"qwen3-asr-0.6b-bf16": {
		modelFile:  "Qwen3-ASR-0.6B-bf16.gguf",
		mmprojFile: "mmproj-Qwen3-ASR-0.6B-bf16.gguf",
		baseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-0.6B-GGUF/resolve/main",
	},
	"qwen3-asr-1.7b-q8_0": {
		modelFile:  "Qwen3-ASR-1.7B-Q8_0.gguf",
		mmprojFile: "mmproj-Qwen3-ASR-1.7B-Q8_0.gguf",
		baseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-1.7B-GGUF/resolve/main",
	},
	"qwen3-asr-1.7b-bf16": {
		modelFile:  "Qwen3-ASR-1.7B-bf16.gguf",
		mmprojFile: "mmproj-Qwen3-ASR-1.7B-bf16.gguf",
		baseURL:    "https://huggingface.co/ggml-org/Qwen3-ASR-1.7B-GGUF/resolve/main",
	},
	"cohere-transcribe-f16": {
		modelFile:  "cohere-transcribe.gguf",
		baseURL:    "https://huggingface.co/cstr/cohere-transcribe-03-2026-GGUF/resolve/main",
	},
	"cohere-transcribe-q8_0": {
		modelFile:  "cohere-transcribe-q8_0.gguf",
		baseURL:    "https://huggingface.co/cstr/cohere-transcribe-03-2026-GGUF/resolve/main",
	},
	"cohere-transcribe-q4_k": {
		modelFile:  "cohere-transcribe-q4_k.gguf",
		baseURL:    "https://huggingface.co/cstr/cohere-transcribe-03-2026-GGUF/resolve/main",
	},
}

func resolveModelPath(v modelVariant, filename string) string {
	path := filepath.Join(modelsDir, v.modelFile, filename)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		path = filepath.Join(path, filename)
	}
	return path
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download llama.cpp libraries and ASR models",
	Run: func(cmd *cobra.Command, args []string) {
		if pullModel != "" {
			if err := downloadModel(pullModel, pullModelPath, pullUpgrade); err != nil {
				fmt.Println("failed to download model:", err.Error())
				os.Exit(1)
			}
			return
		}

		if !pullUpgrade {
			if download.AlreadyInstalled(pullLibPath) {
				fmt.Println("llama.cpp already installed at", pullLibPath)
				return
			}
		}

		if pullProcessor == "" {
			pullProcessor = "cpu"
			if cudaInstalled, cudaVersion := download.HasCUDA(); cudaInstalled {
				fmt.Printf("CUDA detected (version %s), using CUDA build\n", cudaVersion)
				pullProcessor = "cuda"
			}
		}

		if pullVersion == "" || pullVersion == "latest" {
			fmt.Println("installing latest llama.cpp version to", pullLibPath)
		} else {
			fmt.Println("installing llama.cpp version", pullVersion, "to", pullLibPath)
		}

		if err := download.Get(runtime.GOARCH, runtime.GOOS, pullProcessor, pullVersion, pullLibPath); err != nil {
			fmt.Println("failed to download llama.cpp:", err.Error())
			os.Exit(1)
		}

		fmt.Println("done.")
	},
}

func downloadModel(variant, destDir string, upgrade bool) error {
	v, ok := modelVariants[variant]
	if !ok {
		return fmt.Errorf("unknown model variant %q (available: qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16, qwen3-asr-1.7b-q8_0, qwen3-asr-1.7b-bf16, cohere-transcribe-f16, cohere-transcribe-q8_0, cohere-transcribe-q4_k)", variant)
	}

	modelDir := filepath.Join(destDir, v.modelFile)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", modelDir, err)
	}

	var files []struct {
		name string
		url  string
	}
	files = append(files, struct {
		name string
		url  string
	}{v.modelFile, v.baseURL + "/" + v.modelFile})
	if v.mmprojFile != "" {
		files = append(files, struct {
			name string
			url  string
		}{v.mmprojFile, v.baseURL + "/" + v.mmprojFile})
	}

	for _, f := range files {
		destPath := filepath.Join(modelDir, f.name)
		if !upgrade {
			if _, err := os.Stat(destPath); err == nil {
				fmt.Fprintf(os.Stderr, "  %s already exists\n", f.name)
				continue
			}
		}
		if err := downloadFile(f.url, destPath); err != nil {
			return fmt.Errorf("downloading %s: %w", f.name, err)
		}
	}

	return nil
}

func downloadFile(url, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	label := filepath.Base(dest)
	r := io.Reader(resp.Body)
	if total := resp.ContentLength; total > 0 {
		r = &progressReader{
			r:        resp.Body,
			total:    total,
			label:    label,
			barWidth: 30,
		}
		defer fmt.Fprintln(os.Stderr)
	}

	_, err = io.Copy(out, r)
	return err
}

func init() {
	pullLibPath = "/opt/go-asr-bot/llamacpp"
	pullModelPath = "/opt/go-asr-bot/models"

	pullCmd.Flags().StringVar(&pullLibPath, "lib-path", pullLibPath, "destination directory for llama.cpp libraries")
	pullCmd.Flags().StringVar(&pullProcessor, "processor", "", "processor type: cpu, cuda, vulkan, rocm, metal (auto-detected if empty)")
	pullCmd.Flags().StringVar(&pullVersion, "version", "latest", "llama.cpp version to download (e.g. b1234)")
	pullCmd.Flags().BoolVar(&pullUpgrade, "upgrade", false, "force re-download even if already installed")
	pullCmd.Flags().StringVar(&pullModel, "model", "", "ASR model variant to download (qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16, qwen3-asr-1.7b-q8_0, qwen3-asr-1.7b-bf16, cohere-transcribe-f16, cohere-transcribe-q8_0, cohere-transcribe-q4_k)")
	pullCmd.Flags().StringVar(&pullModelPath, "model-path", pullModelPath, "destination directory for model files")
	rootCmd.AddCommand(pullCmd)
}
