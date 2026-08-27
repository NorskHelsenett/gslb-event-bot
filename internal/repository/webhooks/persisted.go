package webhooks

import (
	"github.com/NorskHelsenett/gslb-event-bot/internal/model"
	"github.com/NorskHelsenett/gslb-event-bot/pkg/persistence"
)

type PersistedRepo struct {
	store persistence.Store[model.Webhook]
}

func NewPersistedRepo(store persistence.Store[model.Webhook]) WebhookRepo {
	return &PersistedRepo{
		store: store,
	}
}

func (pr *PersistedRepo) Get(id string) (model.Webhook, error) {
	return pr.store.Load(id)
}

func (pr *PersistedRepo) Store(wh model.Webhook) error {
	return pr.store.Save(wh.ChannelID, wh)
}
