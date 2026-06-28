package bot

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cesto93/go-asr-bot/config"
	"github.com/cesto93/go-asr-bot/internal/asr"
	"github.com/cesto93/go-asr-bot/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api              *tgbotapi.BotAPI
	handlers         *handlers.Handler
	engine           asr.Engine
	authorizedUserID int64
	cfg              *config.Config
	modelName        string
	modelPath        string
	mmprojPath       string
	backend          string
}

func New(cfg *config.Config, modelPath, mmprojPath, modelName, backend string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}

	api.Debug = cfg.Debug
	slog.Info("authorized on account", "bot", api.Self.UserName)

	b := &Bot{
		api:              api,
		cfg:              cfg,
		modelName:        modelName,
		modelPath:        modelPath,
		mmprojPath:       mmprojPath,
		backend:          backend,
		authorizedUserID: cfg.UserID,
	}

	if modelPath != "" {
		b.engine, err = asr.NewFromConfig(cfg, modelPath, mmprojPath, backend)
		if err != nil {
			return nil, err
		}
	}

	b.handlers = handlers.New(api, b.engine, b)

	if err := b.registerCommands(); err != nil {
		slog.Error("failed to register commands", "err", err)
	}

	return b, nil
}

func NewWithoutASR(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}

	api.Debug = cfg.Debug
	slog.Info("authorized on account", "bot", api.Self.UserName)

	b := &Bot{
		api:              api,
		cfg:              cfg,
		authorizedUserID: cfg.UserID,
	}

	b.handlers = handlers.New(api, nil, b)

	if err := b.registerCommands(); err != nil {
		slog.Error("failed to register commands", "err", err)
	}

	return b, nil
}

func (b *Bot) Reload() {
	cfg := config.Load()
	slog.Info("reloading config", "path", config.ConfigPath())

	if cfg.Language != "" && b.engine != nil {
		b.engine.SetLanguage(cfg.Language)
		slog.Info("updated language", "lang", cfg.Language)
	}

	b.authorizedUserID = cfg.UserID
}

func (b *Bot) CurrentModel() string {
	return b.modelName
}

func (b *Bot) CurrentLanguage() string {
	if b.cfg != nil {
		return b.cfg.Language
	}
	return ""
}

func (b *Bot) SetLanguage(lang string) error {
	if b.engine != nil {
		b.engine.SetLanguage(lang)
	}
	if b.cfg != nil {
		b.cfg.Language = lang
	}
	return config.Save(b.cfg)
}

func (b *Bot) SetModel(name string) error {
	v, ok := config.ModelVariants[name]
	if !ok {
		return fmt.Errorf("unknown model %q", name)
	}

	modelPath := config.ResolveModelPath(v, v.ModelFile)
	var mmprojPath string
	if v.MMProjFile != "" {
		mmprojPath = config.ResolveModelPath(v, v.MMProjFile)
	}

	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		if dlerr := config.DownloadModel(name); dlerr != nil {
			return fmt.Errorf("model not found and download failed: %w", dlerr)
		}
	}
	cfg := config.Load()

	if b.engine != nil {
		b.engine.Close()
		b.engine = nil
	}

	engine, err := asr.NewFromConfig(cfg, modelPath, mmprojPath, v.Backend)
	if err != nil {
		return fmt.Errorf("create engine: %w", err)
	}

	b.engine = engine
	b.handlers.SetASR(engine)
	b.modelName = name
	b.modelPath = modelPath
	b.mmprojPath = mmprojPath
	b.backend = v.Backend
	b.cfg = cfg

	cfg.DefaultModel = name
	return config.Save(cfg)
}

func (b *Bot) registerCommands() error {
	_, err := b.api.Request(tgbotapi.NewSetMyCommands(handlers.BotCommands...))
	return err
}

func messageType(msg *tgbotapi.Message) string {
	switch {
	case msg.IsCommand():
		return "command:" + msg.Command()
	case msg.Voice != nil:
		return "voice"
	case msg.Text != "":
		return "text"
	default:
		return "unknown"
	}
}

func (b *Bot) Run() error {
	config.Watch(func(cfg *config.Config) {
		b.Reload()
	})

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if b.authorizedUserID != 0 && update.Message.From.ID != b.authorizedUserID {
				continue
			}
			slog.Info("message received",
				"user", update.Message.From.UserName,
				"chat", update.Message.Chat.ID,
				"type", messageType(update.Message))
			if err := b.handlers.HandleMessage(update.Message); err != nil {
				slog.Error("error handling message", "err", err)
				b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error: %v", err)))
			}
		}
	}

	return nil
}


