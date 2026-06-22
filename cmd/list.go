package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/spf13/cobra"
)

var (
	listModelPath string
	listAvailable bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List ASR models",
	Long: `List downloaded ASR models and their backends.

With --available, show models available for download instead.`,
	Run: func(cmd *cobra.Command, args []string) {
		if listAvailable {
			listAvailableModels()
			return
		}
		if err := listModels(listModelPath); err != nil {
			fmt.Println("error:", err.Error())
			os.Exit(1)
		}
	},
}

type classifiedModel struct {
	name       string
	backend    string
	size       int64
	mmprojSize int64
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func listModels(dir string) error {
	if err := os.MkdirAll(dir, 0775); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", dir, err)
	}

	var models []classifiedModel

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(dir, e.Name())

		modelPaths, mmprojPaths := findGGUF(dirPath)

		mmprojByModel := map[string]string{}
		for _, mmp := range mmprojPaths {
			mmprojByModel[filepath.Base(mmp)] = mmp
		}

		for _, mp := range modelPaths {
			name := filepath.Base(mp)
			backend := determineBackend(name)

			fi, err := os.Stat(mp)
			var size int64
			if err == nil {
				size = fi.Size()
			}

			mmprojName := "mmproj-" + name
			var mmprojSize int64
			if p, ok := mmprojByModel[mmprojName]; ok {
				fi, err := os.Stat(p)
				if err == nil {
					mmprojSize = fi.Size()
				}
			}

			models = append(models, classifiedModel{
				name:       name,
				backend:    backend,
				size:       size,
				mmprojSize: mmprojSize,
			})
		}
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].name < models[j].name
	})

	fmt.Printf("Models in %s:\n\n", dir)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Backend\tModel\tSize\tProjector")
	fmt.Fprintln(w, "-------\t-----\t----\t---------")

	for _, m := range models {
		size := humanSize(m.size)
		proj := "-"
		if m.mmprojSize > 0 {
			proj = humanSize(m.mmprojSize)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.backend, modelVariantName(m.name), size, proj)
	}
	w.Flush()

	if len(models) == 0 {
		fmt.Println("  (no models found)")
	}

	return nil
}

func listAvailableModels() {
	names := make([]string, 0, len(modelVariants))
	for k := range modelVariants {
		names = append(names, k)
	}
	sort.Strings(names)

	fmt.Println("Available models for download:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Backend\tVariant\tFiles")
	fmt.Fprintln(w, "-------\t-------\t-----")

	for _, name := range names {
		v := modelVariants[name]
		files := v.modelFile
		if v.mmprojFile != "" {
			files += ", " + v.mmprojFile
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", v.backend, name, files)
	}
	w.Flush()
}

func findGGUF(dir string) (models, mmprojs []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil
	}

	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if e.IsDir() {
			m, mm := findGGUF(path)
			models = append(models, m...)
			mmprojs = append(mmprojs, mm...)
		} else if strings.HasSuffix(e.Name(), ".gguf") {
			if strings.HasPrefix(e.Name(), "mmproj-") {
				mmprojs = append(mmprojs, path)
			} else {
				models = append(models, path)
			}
		}
	}
	return models, mmprojs
}

func modelVariantName(filename string) string {
	for k, v := range modelVariants {
		if v.modelFile == filename {
			return k
		}
	}
	return filename
}

func determineBackend(name string) string {
	for _, v := range modelVariants {
		if v.modelFile == name {
			return v.backend
		}
	}
	if strings.Contains(name, "Qwen3-ASR") || strings.Contains(name, "qwen3-asr") {
		return "yzma"
	}
	return "crispasr"
}

func init() {
	listModelPath = config.ModelsDir()

	listCmd.Flags().StringVar(&listModelPath, "model-path", listModelPath, "directory to scan for models")
	listCmd.Flags().BoolVar(&listAvailable, "available", false, "show models available for download instead of installed ones")
	rootCmd.AddCommand(listCmd)
}
