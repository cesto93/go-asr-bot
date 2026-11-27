package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cesto93/go-asr-bot/config"
)

type fakeServer struct {
	model string
	lang  string
}

func (f fakeServer) SetLanguage(string) error { return nil }
func (f fakeServer) SetModel(string) error    { return nil }
func (f fakeServer) CurrentModel() string     { return f.model }
func (f fakeServer) CurrentLanguage() string  { return f.lang }

func TestModelKeyboardLabelsAndData(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MODELS_DIR", dir)

	const current = "cohere-transcribe-q8_0"
	h := &Handler{server: fakeServer{model: current}}

	kb := h.modelKeyboard()
	if len(kb.InlineKeyboard) == 0 {
		t.Fatal("keyboard is empty")
	}

	var buttons int
	for _, row := range kb.InlineKeyboard {
		buttons += len(row)
	}
	if want := len(config.ModelVariants) + 1; buttons != want {
		t.Fatalf("got %d buttons, want %d", buttons, want)
	}

	var active, closeSeen int
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData != nil && *btn.CallbackData == callbackModelClose {
				closeSeen++
				continue
			}
			if btn.CallbackData == nil || !strings.HasPrefix(*btn.CallbackData, callbackModelSelect) {
				t.Fatalf("unexpected button data: %v", btn.CallbackData)
			}

			name := strings.TrimPrefix(*btn.CallbackData, callbackModelSelect)
			if _, ok := config.ModelVariants[name]; !ok {
				t.Fatalf("callback data references unknown model %q", name)
			}

			if name == current {
				active++
				if !strings.Contains(btn.Text, "(active)") {
					t.Fatalf("active model label missing marker: %q", btn.Text)
				}
				continue
			}

			if !strings.Contains(btn.Text, "(download)") {
				t.Fatalf("non-downloaded model label missing marker: %q", btn.Text)
			}
		}
	}

	if active != 1 {
		t.Fatalf("got %d active models, want 1", active)
	}
	if closeSeen != 1 {
		t.Fatalf("got %d close buttons, want 1", closeSeen)
	}
}

func TestModelKeyboardMarksDownloadedModels(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MODELS_DIR", dir)

	const downloaded = "parakeet-tdt-0.6b-v3-q5_0"
	v := config.ModelVariants[downloaded]
	if err := os.MkdirAll(filepath.Join(dir, v.ModelFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, v.ModelFile, v.ModelFile), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	h := &Handler{server: fakeServer{model: "cohere-transcribe-f16"}}

	for _, row := range h.modelKeyboard().InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == nil || *btn.CallbackData != callbackModelSelect+downloaded {
				continue
			}
			if strings.Contains(btn.Text, "(download)") {
				t.Fatalf("downloaded model marked for download: %q", btn.Text)
			}
			return
		}
	}
	t.Fatalf("button for %q not found", downloaded)
}

func TestStatusKeyboardOpensMenu(t *testing.T) {
	h := &Handler{server: fakeServer{model: "cohere-transcribe-f16"}}
	kb := h.statusKeyboard()

	if len(kb.InlineKeyboard) != 1 || len(kb.InlineKeyboard[0]) != 1 {
		t.Fatalf("unexpected status keyboard layout")
	}
	btn := kb.InlineKeyboard[0][0]
	if btn.CallbackData == nil || *btn.CallbackData != callbackModelMenu {
		t.Fatalf("status button does not open menu: %v", btn.CallbackData)
	}
}
