package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hybridgroup/yzma/pkg/download"
	"github.com/spf13/cobra"
)

var (
	pullLibPath     string
	pullProcessor   string
	pullVersion     string
	pullUpgrade     bool
	pullModel       string
	pullModelPath   string
)

type modelVariant struct {
	modelFile string
	mmprojFile string
}

var modelVariants = map[string]modelVariant{
	"qwen3-asr-0.6b-q8_0": {
		modelFile:  "Qwen3-ASR-0.6B-Q8_0.gguf",
		mmprojFile: "mmproj-Qwen3-ASR-0.6B-Q8_0.gguf",
	},
	"qwen3-asr-0.6b-bf16": {
		modelFile:  "Qwen3-ASR-0.6B-bf16.gguf",
		mmprojFile: "mmproj-Qwen3-ASR-0.6B-bf16.gguf",
	},
}

const modelBaseURL = "https://huggingface.co/ggml-org/Qwen3-ASR-0.6B-GGUF/resolve/main"

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
		return fmt.Errorf("unknown model variant %q (available: qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16)", variant)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", destDir, err)
	}

	files := []struct {
		name string
		url  string
	}{
		{v.modelFile, modelBaseURL + "/" + v.modelFile},
		{v.mmprojFile, modelBaseURL + "/" + v.mmprojFile},
	}

	for _, f := range files {
		destPath := filepath.Join(destDir, f.name)
		if !upgrade {
			if _, err := os.Stat(destPath); err == nil {
				fmt.Println(f.name, "already exists at", destPath)
				continue
			}
		}
		fmt.Println("downloading", f.name, "...")
		if err := downloadFile(f.url, destPath); err != nil {
			return fmt.Errorf("downloading %s: %w", f.name, err)
		}
		fmt.Println("downloaded", f.name)
	}

	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func init() {
	pullCmd.Flags().StringVar(&pullLibPath, "lib-path", "llamacpp", "destination directory for llama.cpp libraries")
	pullCmd.Flags().StringVar(&pullProcessor, "processor", "", "processor type: cpu, cuda, vulkan, rocm, metal (auto-detected if empty)")
	pullCmd.Flags().StringVar(&pullVersion, "version", "latest", "llama.cpp version to download (e.g. b1234)")
	pullCmd.Flags().BoolVar(&pullUpgrade, "upgrade", false, "force re-download even if already installed")
	pullCmd.Flags().StringVar(&pullModel, "model", "", "ASR model variant to download (qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16)")
	pullCmd.Flags().StringVar(&pullModelPath, "model-path", "models", "destination directory for model files")
	rootCmd.AddCommand(pullCmd)
}
