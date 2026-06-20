package asr

import (
	"fmt"

	"github.com/cesto93/go-asr-bot/config"
)

type backendFactory func(modelPath string, threads int, lang string) (Engine, error)

var backends = map[string]backendFactory{}

func NewFromConfig(cfg *config.Config) (Engine, error) {
	if fn, ok := backends["crispasr"]; ok {
		return fn(cfg.CrispasrModelPath, cfg.CrispasrThreads, cfg.Language)
	}

	e := newYzma(cfg.ModelPath, cfg.MMProjPath, cfg.YzmaLib, cfg.Language)
	if err := e.Init(); err != nil {
		return nil, fmt.Errorf("yzma init: %w", err)
	}
	return e, nil
}
