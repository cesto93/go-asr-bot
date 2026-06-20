package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/cesto93/go-asr-bot/internal/asr"
	"github.com/spf13/cobra"
)

var (
	runModel string
	runLang  string
)

var runCmd = &cobra.Command{
	Use:   "run <audio-file>",
	Short: "Transcribe an audio file using the ASR engine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		audioFile := args[0]
		cfg := config.Load()

		if runLang != "" {
			cfg.Language = runLang
		}

		modelName := runModel
		if modelName == "" {
			modelName = cfg.DefaultModel
		}
		v, ok := modelVariants[modelName]
		if !ok {
			fmt.Printf("unknown model %q\n\navailable variants:\n", modelName)
			keys := make([]string, 0, len(modelVariants))
			for k := range modelVariants {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("  %s\n", k)
			}
			os.Exit(1)
		}

		modelPath := resolveModelPath(v, v.modelFile)
		var mmprojPath string
		if v.mmprojFile != "" {
			mmprojPath = resolveModelPath(v, v.mmprojFile)
		}
		backend := v.backend

		if _, err := os.Stat(modelPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Model file not found at %s\n", modelPath)
			os.Exit(1)
		}
		if backend == "yzma" {
			if _, err := os.Stat(mmprojPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Multimodal projector file not found at %s\n", mmprojPath)
				os.Exit(1)
			}
		}

		engine, err := asr.NewFromConfig(cfg, modelPath, mmprojPath, backend)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize ASR engine: %v\n", err)
			os.Exit(1)
		}
		defer engine.Close()

		fmt.Fprintf(os.Stderr, "Transcribing %s...\n", audioFile)

		pcm, err := asr.AudioToPCM(audioFile, engine.SampleRate())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to convert audio: %v\n", err)
			os.Exit(1)
		}

		var lang string
		if runLang != "" {
			lang = runLang
		} else {
			lang = cfg.Language
		}

		text, err := engine.TranscribeLang(pcm, lang)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to transcribe: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(text)
	},
}

func init() {
	runCmd.Flags().StringVar(&runModel, "model", "", "ASR model name (one of: qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16, qwen3-asr-1.7b-q8_0, qwen3-asr-1.7b-bf16, cohere-transcribe-f16, cohere-transcribe-q8_0, cohere-transcribe-q4_k, parakeet-tdt-0.6b-v3, parakeet-tdt-0.6b-v3-q8_0, parakeet-tdt-0.6b-v3-q5_0, parakeet-tdt-0.6b-v3-q4_k)")
	runCmd.Flags().StringVar(&runLang, "lang", "", "Source language (ISO 639-1, e.g. en, fr, de)")
	rootCmd.AddCommand(runCmd)
}
