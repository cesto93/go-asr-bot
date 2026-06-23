package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/spf13/cobra"
)



var (
	pullUpgrade   bool
	pullModel     string
	pullModelPath string
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download ASR models",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			pullModel = args[0]
		}

		if pullModel == "" {
			fmt.Println("no model specified. Use --model to specify a model variant.")
			os.Exit(1)
		}

		if err := downloadModel(pullModel, pullModelPath, pullUpgrade); err != nil {
			fmt.Println("failed to download model:", err.Error())
			os.Exit(1)
		}
	},
}

func downloadModel(variant, destDir string, upgrade bool) error {
	v, ok := config.ModelVariants[variant]
	if !ok {
		return fmt.Errorf("unknown model variant %q", variant)
	}

	modelDir := filepath.Join(destDir, v.ModelFile)
	if upgrade {
		os.RemoveAll(modelDir)
	}

	return config.DownloadModel(variant)
}

func init() {
	pullModelPath = config.ModelsDir()

	pullCmd.Flags().BoolVar(&pullUpgrade, "upgrade", false, "force re-download even if already installed")
	pullCmd.Flags().StringVar(&pullModel, "model", "", "ASR model variant to download")
	pullCmd.Flags().StringVar(&pullModelPath, "model-path", pullModelPath, "destination directory for model files")
	rootCmd.AddCommand(pullCmd)
}
