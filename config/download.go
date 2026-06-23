package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadModel(variant string) error {
	v, ok := ModelVariants[variant]
	if !ok {
		return fmt.Errorf("unknown model variant %q", variant)
	}

	modelDir := filepath.Join(ModelsDir(), v.ModelFile)
	if err := os.MkdirAll(modelDir, 0775); err != nil {
		return fmt.Errorf("create directory %s: %w", modelDir, err)
	}

	type fileSpec struct{ name, url string }
	files := []fileSpec{{v.ModelFile, v.BaseURL + "/" + v.ModelFile}}
	if v.MMProjFile != "" {
		files = append(files, fileSpec{v.MMProjFile, v.BaseURL + "/" + v.MMProjFile})
	}

	for _, f := range files {
		destPath := filepath.Join(modelDir, f.name)
		if _, err := os.Stat(destPath); err == nil {
			continue
		}
		if err := downloadFile(f.url, destPath); err != nil {
			return fmt.Errorf("downloading %s: %w", f.name, err)
		}
	}

	return nil
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
