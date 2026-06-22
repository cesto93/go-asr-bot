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
	configSetModel           string
	configSetLang            string
	configSetUserID          int64
	configSetCrispasrThreads int
	configSetToken           string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display or modify configuration",
	Long: `Display current configuration from config file, environment variables, and defaults.

With --set-default-model, --set-language, --set-user-id, --set-crispasr-threads, or --set-telegram-token, update the config file.

NOTE: The Telegram token will be stored in plaintext in the config file.
The file will be created with restricted permissions (0600).`,
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

		if cmd.Flags().Changed("set-user-id") {
			cfg.UserID = configSetUserID
			changed = true
		}

		if cmd.Flags().Changed("set-crispasr-threads") {
			cfg.CrispasrThreads = configSetCrispasrThreads
			changed = true
		}

		if cmd.Flags().Changed("set-telegram-token") {
			cfg.TelegramToken = configSetToken
			changed = true
		}

		if changed {
			if err := config.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Saved config to %s\n", config.ConfigPath)
			if configSetLang != "" {
				fmt.Println("Language will take effect immediately (auto-detected via fsnotify)")
			}
			if configSetModel != "" {
				fmt.Println("Default model change requires a restart")
			}
			if cmd.Flags().Changed("set-user-id") {
				fmt.Println("UserID restriction will take effect immediately (auto-detected via fsnotify)")
			}
			if cmd.Flags().Changed("set-crispasr-threads") {
				fmt.Println("CrispasrThreads change requires a restart")
			}
			if cmd.Flags().Changed("set-telegram-token") {
				fmt.Println("Telegram token change requires a restart")
			}
			fmt.Println()
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
	configCmd.Flags().Int64Var(&configSetUserID, "set-user-id", 0, "Restrict bot to a single Telegram user ID (0 = allow all)")
	configCmd.Flags().IntVar(&configSetCrispasrThreads, "set-crispasr-threads", 0, "Set CPU threads for CrispASR backend")
	configCmd.Flags().StringVar(&configSetToken, "set-telegram-token", "", "Set Telegram bot token in config file (stored in plaintext)")
	rootCmd.AddCommand(configCmd)
}
