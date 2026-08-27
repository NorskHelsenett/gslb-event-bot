package hslack

import (
	"context"
	"log/slog"

	"github.com/NorskHelsenett/gslb-event-bot/internal/config"
	"github.com/slack-go/slack"
)

type Handler interface {
	Run(context.Context) error
}

func Setup() Handler {
	client := slack.New(
		config.Slack().BotToken(),
		slack.OptionAppLevelToken(config.Slack().AppToken()),
		slack.OptionLog(slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)),
	)

	return NewSocketHandler(client)
}
