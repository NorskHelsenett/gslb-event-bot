package webhooks

import (
	"github.com/NorskHelsenett/gslb-event-bot/internal/model"
)

func NewInMemoryRepo() WebhookRepo {
	return &inMemoryRepo{
		cache: make(map[string]model.Webhook),
	}
}

type inMemoryRepo struct {
	cache map[string]model.Webhook
}

func (ir *inMemoryRepo) Get(id string) (model.Webhook, error) {
	return ir.cache[id], nil
}

func (ir *inMemoryRepo) Store(wh model.Webhook) error {
	ir.cache[wh.ChannelID] = wh
	return nil
}
