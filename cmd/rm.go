package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var rmModelPath string

var rmCmd = &cobra.Command{
	Use:   "rm <model>",
	Short: "Delete a downloaded ASR model",
	Long: `Delete a downloaded ASR model by its variant name.

Use "go-asr-bot list" to see available models, or
"go-asr-bot pull --model <name>" to download one first.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		variant := args[0]

		v, ok := modelVariants[variant]
		if !ok {
			fmt.Printf("unknown model %q\n\navailable variants:\n", variant)
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

		modelDir := filepath.Join(rmModelPath, v.modelFile)

		if _, err := os.Stat(modelDir); os.IsNotExist(err) {
			fmt.Printf("model %q is not downloaded (directory %s does not exist)\n", variant, modelDir)
			return
		}

		var removed []string
		entries, _ := os.ReadDir(modelDir)
		for _, e := range entries {
			removed = append(removed, filepath.Join(modelDir, e.Name()))
		}

		fmt.Printf("removing %s...\n", modelDir)
		if err := os.RemoveAll(modelDir); err != nil {
			fmt.Printf("failed to remove %s: %v\n", modelDir, err)
			os.Exit(1)
		}
		for _, path := range removed {
			fmt.Printf("  deleted %s\n", path)
		}
		fmt.Printf("deleted %q\n", variant)
	},
}

func init() {
	rmModelPath = "/opt/go-asr-bot/models"

	rmCmd.Flags().StringVar(&rmModelPath, "model-path", rmModelPath, "directory where models are stored")
	rootCmd.AddCommand(rmCmd)
}
