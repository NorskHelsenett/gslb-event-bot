package config

import (
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/NorskHelsenett/gslb-event-bot/pkg/bslog"
)

var cfg *Config

type Config struct {
	Server        server   `mapstructure:"server"`
	Valkey        valkey   `mapstructre:"valkey"`
	MQ            mq       `mapstructure:"mq"`
	Slack         slack    `mapstructure:"slack"`
	Webhooks      webhooks `mapstructure:"webhooks"`
	secretsLoaded int
	secretsTotal  int
}

func (c *Config) LogValue() slog.Value {

	return slog.GroupValue()
}

func Server() *server {
	return &cfg.Server
}

func Valkey() *valkey {
	return &cfg.Valkey
}

func MQ() *mq {
	return &cfg.MQ
}

func Slack() *slack {
	return &cfg.Slack
}

func Webhooks() *webhooks {
	return &cfg.Webhooks
}

func init() {
	var err error

	cfg, err = new()
	if err != nil {
		log.Fatalf("unable to load config: %s", err.Error())
	}

	var handler slog.Handler
	handlerOpts := &slog.HandlerOptions{
		Level:       cfg.Server.LogLevel(),
		ReplaceAttr: bslog.BaseReplaceAttr,
	}

	switch cfg.Server.ENV {
	case "dev", "development", "DEV", "DEVELOPMENT", "local", "LOCAL":
		handler = bslog.NewHandler(
			os.Stdout, // log output
			// slog handler factory
			func(w io.Writer) slog.Handler {
				return slog.NewTextHandler(w, handlerOpts)
			},
			// options
			bslog.InDevMode(),
			bslog.WithColor(),
		)

	case "prod", "production", "PROD", "PRODUCTION":
		handler = bslog.NewHandler(
			os.Stdout,
			func(w io.Writer) slog.Handler {
				return slog.NewJSONHandler(w, handlerOpts)
			},
			//bslog.WithSplunkMultiHandler("<secret>", "<splunk_index>", slog.LevelInfo),
		)
	default:
		handler = bslog.NewHandler(
			os.Stdout, // log output
			// slog handler factory
			func(w io.Writer) slog.Handler {
				return slog.NewTextHandler(w, handlerOpts)
			},
			// options
			bslog.InDevMode(),
			bslog.WithColor(),
		)
	}

	slog.SetDefault(slog.New(handler))
	bslog.Info("config-loaded", slog.Any("config", cfg))
}

type server struct {
	ENV       string `mapstructure:"env"`
	LOG_LEVEL string `mapstructure:"log_level"`
}

func (s *server) Environment() string {
	return s.ENV
}

func (s *server) LogLevel() slog.Level {
	switch s.LOG_LEVEL {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "WARN":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	case "fatal", "FATAL":
		return bslog.LevelFatal
	default:
		return slog.LevelDebug
	}
}

type valkey struct {
	Addr string `mapstructure:"addr"`
	USER string `mapstructure:"user"`
	PASS string `mapstructure:"pass"`
}

func (v *valkey) Address() string {
	return v.Addr
}

func (v *valkey) User() string {
	return v.USER
}

func (v *valkey) Password() string {
	return v.PASS
}

type mq struct {
	Enable   bool   `mapstructure:"enabled"`
	Usr      string `mapstructure:"user"`
	Passwd   string `mapstructure:"pass"`
	EndPoint string `mapstructure:"endpoint"`
}

func (mq *mq) Enabled() bool {
	return mq.Enable
}

func (mq *mq) User() string {
	return mq.Usr
}

func (mq *mq) Pass() string {
	return mq.Passwd
}

func (mq *mq) Endpoint() string {
	return mq.EndPoint
}

type slack struct {
	Enable         bool   `mapstructure:"enabled"`
	APP_TOKEN      string `mapstructure:"app_token"`
	BOT_TOKEN      string `mapstructure:"bot_token"`
	SIGNING_SECRET string `mapstructure:"signing_secret"`
}

func (s *slack) Enabled() bool {
	return s.Enable
}

func (s *slack) AppToken() string {
	return s.APP_TOKEN
}

func (s *slack) BotToken() string {
	return s.BOT_TOKEN
}

func (s *slack) SigningSecret() string {
	return s.SIGNING_SECRET
}

type webhooks struct {
	EVENTS []string `mapstructure:"events"`
}

func (w *webhooks) Events() []string {
	return w.EVENTS
}
