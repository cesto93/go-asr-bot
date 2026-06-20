package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/cesto93/go-asr-bot/internal/asr"
	"github.com/spf13/cobra"
)

var runModel string

var runCmd = &cobra.Command{
	Use:   "run <audio-file>",
	Short: "Transcribe an audio file using the ASR engine",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		audioFile := args[0]
		cfg := config.Load()

		if runModel != "" {
			v, ok := modelVariants[runModel]
			if !ok {
				fmt.Printf("unknown model %q\n\navailable variants:\n", runModel)
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

			cfg.ASRBackend = v.backend

			switch cfg.ASRBackend {
			case "yzma":
				cfg.ModelPath = resolveModelPath(v, v.modelFile)
				if v.mmprojFile != "" {
					cfg.MMProjPath = resolveModelPath(v, v.mmprojFile)
				}
			case "crispasr":
				cfg.CrispasrModelPath = resolveModelPath(v, v.modelFile)
			}
		}

		if cfg.ASRBackend == "yzma" {
			if _, err := os.Stat(cfg.ModelPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Model file not found at %s\n", cfg.ModelPath)
				os.Exit(1)
			}
			if _, err := os.Stat(cfg.MMProjPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Multimodal projector file not found at %s\n", cfg.MMProjPath)
				os.Exit(1)
			}
		}
		if cfg.ASRBackend == "crispasr" {
			if _, err := os.Stat(cfg.CrispasrModelPath); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "CrispASR model file not found at %s\n", cfg.CrispasrModelPath)
				os.Exit(1)
			}
		}

		engine, err := asr.NewFromConfig(cfg)
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

		text, err := engine.Transcribe(pcm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to transcribe: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(text)
	},
}

func init() {
	runCmd.Flags().StringVar(&runModel, "model", "", "ASR model name (one of: qwen3-asr-0.6b-q8_0, qwen3-asr-0.6b-bf16, qwen3-asr-1.7b-q8_0, qwen3-asr-1.7b-bf16, cohere-transcribe-f16, cohere-transcribe-q8_0, cohere-transcribe-q4_k)")
	rootCmd.AddCommand(runCmd)
}
