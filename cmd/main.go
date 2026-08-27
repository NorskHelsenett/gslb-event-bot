package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NorskHelsenett/gslb-event-bot/internal/api/handlers/hslack"
	"github.com/NorskHelsenett/gslb-event-bot/internal/config"
	"github.com/NorskHelsenett/gslb-event-bot/internal/model"
	"github.com/NorskHelsenett/gslb-event-bot/internal/repository/webhooks"
	"github.com/NorskHelsenett/gslb-event-bot/pkg/bslog"
	valkeyStore "github.com/NorskHelsenett/gslb-event-bot/pkg/persistence/store/valkey"
	"github.com/valkey-io/valkey-go"
)

var (
	version   string
	buildDate string
)

func main() {
	bslog.Info("running gslb-event-bot", slog.String("version", version), slog.String("buildDate", buildDate))

	valkeyClient, err := valkeyStore.NewClient(
		valkey.ClientOption{
			InitAddress: []string{config.Valkey().Address()},
			Username:    config.Valkey().User(),
			Password:    config.Valkey().Password(),
		},
	)
	if err != nil {
		bslog.Fatal("failed to establish valkey connection", slog.String("reason", err.Error()))
	}
	webhooksStore, err := valkeyStore.NewStore[model.Webhook](valkeyClient, "event_bot:webhooks", time.Hour)
	if err != nil {
		bslog.Fatal("failed to create webhooks store", slog.String("reason", err.Error()))
	}

	webhooks.Init(webhooksStore)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigterm := make(chan os.Signal, 1)
	serverErr := make(chan error, 1)
	signal.Notify(sigterm, os.Interrupt, syscall.SIGTERM)

	handler := hslack.Setup()
	go func() {
		err := handler.Run(ctx)
		if err != nil {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		bslog.Fatal("un-expected error ocurred, no longer serving slack requests", slog.String("reason", err.Error()))
	case <-sigterm:
		bslog.Info("gracefully shutting down...")
	}
}
