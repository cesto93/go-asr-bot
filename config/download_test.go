package config

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadModelTo(t *testing.T) {
	const body = "fake gguf data"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	defer srv.Close()

	const variant = "test-download"
	ModelVariants[variant] = ModelVariant{ModelFile: "test.gguf", BaseURL: srv.URL}
	defer delete(ModelVariants, variant)

	destDir := filepath.Join(t.TempDir(), "nested", "models")
	if err := DownloadModelTo(variant, destDir); err != nil {
		t.Fatalf("DownloadModelTo: %v", err)
	}

	dest := filepath.Join(destDir, "test.gguf", "test.gguf")
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(got) != body {
		t.Fatalf("downloaded content = %q, want %q", got, body)
	}

	if _, err := os.Stat(dest + ".part"); !os.IsNotExist(err) {
		t.Fatalf("temporary file was not removed: %v", err)
	}
}

func TestEnsureModelsDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing", "models")
	t.Setenv("MODELS_DIR", dir)

	if err := EnsureModelsDir(); err != nil {
		t.Fatalf("EnsureModelsDir: %v", err)
	}

	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat models dir: %v", err)
	}
	if !fi.IsDir() {
		t.Fatalf("%s is not a directory", dir)
	}
}
