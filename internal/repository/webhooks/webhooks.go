package webhooks

import (
	"github.com/NorskHelsenett/gslb-event-bot/internal/model"
	"github.com/NorskHelsenett/gslb-event-bot/pkg/persistence"
)

type WebhookRepo interface {
	Get(id string) (model.Webhook, error)
	Store(wh model.Webhook) error
}

var repo WebhookRepo

func Init(store persistence.Store[model.Webhook]) {
	repo = NewPersistedRepo(store)
}

func Get(id string) (model.Webhook, error) {
	return repo.Get(id)
}

func Store(wh model.Webhook) error {
	return repo.Store(wh)
}
