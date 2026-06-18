package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listModelPath string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List downloaded ASR models and their backends",
	Run: func(cmd *cobra.Command, args []string) {
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
			if v.mmprojFile != "" {
				return "yzma"
			}
			return "crispasr"
		}
	}
	if strings.Contains(name, "Qwen3-ASR") || strings.Contains(name, "qwen3-asr") {
		return "yzma"
	}
	return "crispasr"
}

func init() {
	listModelPath = "/opt/go-asr-bot/models"

	listCmd.Flags().StringVar(&listModelPath, "model-path", listModelPath, "directory to scan for models")
	rootCmd.AddCommand(listCmd)
}
