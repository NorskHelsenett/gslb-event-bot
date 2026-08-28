package webhooks

import (
	"context"
	"fmt"

	"github.com/NorskHelsenett/gslb-event-bot/internal/config"
	"github.com/NorskHelsenett/gslb-event-bot/internal/model"
	"github.com/NorskHelsenett/gslb-event-bot/pkg/mq"
	"github.com/NorskHelsenett/gslb-event-bot/pkg/mq/rabbitmq"
)

type WebhooksBroker struct {
	client mq.MessageBroker[model.Webhook]
}

func NewWebhooksBroker(ctx context.Context) *WebhooksBroker {
	mqCfg := config.MQ()
	return &WebhooksBroker{
		client: rabbitmq.New(ctx,
			fmt.Sprintf(
				"amqp://%s:%s@%s",
				mqCfg.User(),
				mqCfg.Pass(),
				mqCfg.Endpoint(),
			),
			rabbitmq.WithExchange[model.Webhook]("ex.gslb.webhooks-registration"),
			rabbitmq.WithQueue[model.Webhook]("q.gslb.webhooks-registration"),
			rabbitmq.WithFanout[model.Webhook](),
		),
	}
}

func (wb *WebhooksBroker) Publish(ctx context.Context, wh model.Webhook) error {
	return wb.client.Publish(ctx, wh)
}
