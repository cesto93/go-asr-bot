package asr

import (
	"fmt"

	"github.com/cesto93/go-asr-bot/config"
)

type backendFactory func(modelPath string, threads int, lang string) (Engine, error)

var backends = map[string]backendFactory{}

func NewFromConfig(cfg *config.Config, modelPath, mmprojPath, backend string) (Engine, error) {
	switch backend {
	case "crispasr":
		fn, ok := backends["crispasr"]
		if !ok {
			return nil, fmt.Errorf("crispasr backend not available (CGO disabled?)")
		}
		return fn(modelPath, cfg.CrispasrThreads, cfg.Language)
	default:
		e := newYzma(modelPath, mmprojPath, config.DefaultYZMALib, cfg.Language)
		if err := e.Init(); err != nil {
			return nil, fmt.Errorf("yzma init: %w", err)
		}
		return e, nil
	}
}
