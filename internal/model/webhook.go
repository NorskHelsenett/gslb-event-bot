package model

type Webhook struct {
	ChannelID    string             `json:"id"`
	Subscription EventsSubscription `json:"subscription"`
	Options      WebhookOptions     `json:"options"`
}

type EventsSubscription struct {
	Events  []string       `json:"events"`
	Options map[string]any `json:"options"`
}

type WebhookOptions struct {
	Format string `json:"format"`
}
