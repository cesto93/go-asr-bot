package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/spf13/cobra"
)

var (
	configSetModel string
	configSetLang  string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or modify configuration",
	Long: `Display current configuration from config file, environment variables, and defaults.

With --set-default-model or --set-language, update the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		changed := false

		if configSetModel != "" {
			if _, ok := modelVariants[configSetModel]; !ok {
				fmt.Printf("unknown model %q\n\navailable variants:\n", configSetModel)
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
			cfg.DefaultModel = configSetModel
			changed = true
		}

		if configSetLang != "" {
			cfg.Language = configSetLang
			changed = true
		}

		if changed {
			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Saved config to %s\n\n", config.ConfigPath)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "Key\tValue")
		fmt.Fprintln(w, "---\t-----")
		fmt.Fprintf(w, "TelegramToken\t%s\n", maskToken(cfg.TelegramToken))
		fmt.Fprintf(w, "Debug\t%t\n", cfg.Debug)
		fmt.Fprintf(w, "UserID\t%d\n", cfg.UserID)
		fmt.Fprintf(w, "Language\t%s\n", cfg.Language)
		fmt.Fprintf(w, "DefaultModel\t%s\n", cfg.DefaultModel)
		fmt.Fprintf(w, "CrispasrThreads\t%d\n", cfg.CrispasrThreads)
		fmt.Fprintf(w, "ConfigFile\t%s\n", config.ConfigPath)
		w.Flush()
	},
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func init() {
	configCmd.Flags().StringVar(&configSetModel, "set-default-model", "", "Set default ASR model variant in config file")
	configCmd.Flags().StringVar(&configSetLang, "set-language", "", "Set source language (ISO 639-1) in config file")
	rootCmd.AddCommand(configCmd)
}
