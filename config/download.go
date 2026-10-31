package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadModel(variant string) error {
	return DownloadModelTo(variant, ModelsDir())
}

func DownloadModelTo(variant, destDir string) error {
	v, ok := ModelVariants[variant]
	if !ok {
		return fmt.Errorf("unknown model variant %q", variant)
	}

	if err := os.MkdirAll(destDir, 0775); err != nil {
		return fmt.Errorf("create models directory %s: %w", destDir, err)
	}

	modelDir := filepath.Join(destDir, v.ModelFile)
	if err := os.MkdirAll(modelDir, 0775); err != nil {
		return fmt.Errorf("create model directory %s: %w", modelDir, err)
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

	tmp := dest + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}

	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, dest)
}
