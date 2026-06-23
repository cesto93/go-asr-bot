package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/spf13/cobra"
)

type progressReader struct {
	r        io.Reader
	total    int64
	current  int64
	label    string
	lastPct  int
	barWidth int
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.current += int64(n)
	pct := int(pr.current * 100 / pr.total)
	if pct > 100 {
		pct = 100
	}
	if pct != pr.lastPct {
		pr.lastPct = pct
		pr.print()
	}
	return n, err
}

func (pr *progressReader) print() {
	cur := humanSize(pr.current)
	tot := humanSize(pr.total)
	filled := pr.barWidth * pr.lastPct / 100
	bar := strings.Repeat("=", filled)
	if filled < pr.barWidth {
		bar += ">" + strings.Repeat(" ", pr.barWidth-filled-1)
	} else {
		bar += "="
	}
	fmt.Fprintf(os.Stderr, "\r  %s [%s] %3d%% (%s/%s)", pr.label, bar, pr.lastPct, cur, tot)
}



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
	if err := os.MkdirAll(modelDir, 0775); err != nil {
		return fmt.Errorf("create directory %s: %w", modelDir, err)
	}

	var files []struct {
		name string
		url  string
	}
	files = append(files, struct {
		name string
		url  string
	}{v.ModelFile, v.BaseURL + "/" + v.ModelFile})
	if v.MMProjFile != "" {
		files = append(files, struct {
			name string
			url  string
		}{v.MMProjFile, v.BaseURL + "/" + v.MMProjFile})
	}

	for _, f := range files {
		destPath := filepath.Join(modelDir, f.name)
		if !upgrade {
			if _, err := os.Stat(destPath); err == nil {
				fmt.Fprintf(os.Stderr, "  %s already exists\n", f.name)
				continue
			}
		}
		if err := downloadFile(f.url, destPath); err != nil {
			return fmt.Errorf("downloading %s: %w", f.name, err)
		}
	}

	return nil
}

func downloadFile(url, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	out.Chmod(0664)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	label := filepath.Base(dest)
	r := io.Reader(resp.Body)
	if total := resp.ContentLength; total > 0 {
		r = &progressReader{
			r:        resp.Body,
			total:    total,
			label:    label,
			barWidth: 30,
		}
		defer fmt.Fprintln(os.Stderr)
	}

	_, err = io.Copy(out, r)
	return err
}

func init() {
	pullModelPath = config.ModelsDir()

	pullCmd.Flags().BoolVar(&pullUpgrade, "upgrade", false, "force re-download even if already installed")
	pullCmd.Flags().StringVar(&pullModel, "model", "", "ASR model variant to download")
	pullCmd.Flags().StringVar(&pullModelPath, "model-path", pullModelPath, "destination directory for model files")
	rootCmd.AddCommand(pullCmd)
}
