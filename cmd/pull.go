package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/hybridgroup/yzma/pkg/download"
	"github.com/spf13/cobra"
)

var (
	pullLibPath   string
	pullProcessor string
	pullVersion   string
	pullUpgrade   bool
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download llama.cpp libraries",
	Run: func(cmd *cobra.Command, args []string) {
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

func init() {
	pullCmd.Flags().StringVar(&pullLibPath, "lib-path", "llamacpp", "destination directory for llama.cpp libraries")
	pullCmd.Flags().StringVar(&pullProcessor, "processor", "", "processor type: cpu, cuda, vulkan, rocm, metal (auto-detected if empty)")
	pullCmd.Flags().StringVar(&pullVersion, "version", "latest", "llama.cpp version to download (e.g. b1234)")
	pullCmd.Flags().BoolVar(&pullUpgrade, "upgrade", false, "force re-download even if already installed")
	rootCmd.AddCommand(pullCmd)
}
