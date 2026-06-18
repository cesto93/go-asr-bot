package asr

import (
	"fmt"

	"github.com/cesto93/go-asr-bot/config"
)

type backendFactory func(modelPath string, threads int) (Engine, error)

var backends = map[string]backendFactory{}

func NewFromConfig(cfg *config.Config) (Engine, error) {
	switch cfg.ASRBackend {
	case "yzma":
		e := newYzma(cfg.ModelPath, cfg.MMProjPath, cfg.YzmaLib)
		if err := e.Init(); err != nil {
			return nil, fmt.Errorf("yzma init: %w", err)
		}
		return e, nil

	case "crispasr":
		fn, ok := backends["crispasr"]
		if !ok {
			return nil, fmt.Errorf("crispasr backend not available; rebuild with CGO_ENABLED=1")
		}
		return fn(cfg.CrispasrModelPath, cfg.CrispasrThreads)

	default:
		return nil, fmt.Errorf("unknown ASR backend: %q", cfg.ASRBackend)
	}
}
