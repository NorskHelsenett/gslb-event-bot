package hslack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/NorskHelsenett/gslb-event-bot/internal/blocks"
	whBroker "github.com/NorskHelsenett/gslb-event-bot/internal/brokers/webhooks"
	"github.com/NorskHelsenett/gslb-event-bot/internal/model"
	whRepo "github.com/NorskHelsenett/gslb-event-bot/internal/repository/webhooks"
	"github.com/NorskHelsenett/gslb-event-bot/pkg/bslog"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// SocketHandler manages Slack Socket Mode connections and event handling.
type SocketHandler struct {
	handler *socketmode.SocketmodeHandler
	builder *blocks.SlackMessageBuilder
	broker  *whBroker.WebhooksBroker
}

func NewSocketHandler(client *slack.Client) *SocketHandler {
	handler := socketmode.NewSocketmodeHandler(
		socketmode.New(
			client,
			// route socketmode's internal debug/connection logs through the app's shared slog handler
			socketmode.OptionLog(slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)),
		),
	)

	return &SocketHandler{
		handler: handler,
		builder: blocks.NewSlackMessageBuilder(),
		broker:  whBroker.NewWebhooksBroker(context.Background()),
	}
}

// Run starts the Socket Mode handler and begins listening for incoming events.
// It returns an error if the connection fails or if an error occurs during event processing.
func (sh *SocketHandler) Run(ctx context.Context) error {
	sh.register()
	return sh.handler.RunEventLoopContext(ctx)
}

// register registers the Socket Mode handler with the appropriate event listeners
// or routing mechanisms to enable event handling.
func (sh *SocketHandler) register() {
	sh.handler.HandleSlashCommand("/subscribe", sh.HandleSlashCommandSubscribe)
	sh.handler.HandleInteraction(slack.InteractionTypeViewSubmission, sh.HandleInteractionSubmission)
}

func (sh *SocketHandler) HandleSlashCommandSubscribe(e *socketmode.Event, c *socketmode.Client) {
	c.Ack(*e.Request)

	cmd, ok := e.Data.(slack.SlashCommand)
	if !ok {
		return
	}
	logger := bslog.With(slog.String("channel_id", cmd.ChannelID))

	logger.Debug("received subscription command")

	channel, err := c.Client.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: cmd.ChannelID,
	})
	if err != nil {
		logger.Error("failed to fetch channel info", slog.String("reason", err.Error()))
		slack.PostWebhook(cmd.ResponseURL, &slack.WebhookMessage{Text: "failed to get channel information, Please contact #drift-lastbalansering"})
		return
	}

	if !channel.IsMember {
		logger.Info("skipping webhooks registration: event-bot is not a member in channel")
		slack.PostWebhook(cmd.ResponseURL, &slack.WebhookMessage{Text: "gslb-event-bot needs to be a member in the channel to register webhooks, please invite me in your channel to continue."})
		return
	}

	view, _ := sh.builder.NewSubscriptionModal(cmd.ChannelID)
	view.PrivateMetadata = cmd.ChannelID

	res, err := c.OpenView(cmd.TriggerID, view)
	if err != nil {
		logger.Error("failed to open modal view for client", slog.String("reason", err.Error()))

		// notify the channel since the user never saw a modal to explain the failure
		if _, _, postErr := c.PostMessage(cmd.ChannelID, slack.MsgOptionText("Something unexpected happened while processing your request. Please contact #drift-lastbalansering", false)); postErr != nil {
			logger.Error("failed to post error message to channel", slog.String("reason", postErr.Error()))
		}
		return
	}

	if !res.SlackResponse.Ok {
		logger.Error("slack response failed", slog.String("reason", res.Error), slog.String("request_id", e.Request.EnvelopeID))
		return
	}
}

func (sh *SocketHandler) HandleInteractionSubmission(e *socketmode.Event, c *socketmode.Client) {
	c.Ack(*e.Request)
	payload, ok := e.Data.(slack.InteractionCallback)
	if !ok {
		return
	}

	switch payload.View.CallbackID {
	case "gslb_events_subscription":
		sh.registerWebhook(payload, c)
	default:
	}
}

func (sh *SocketHandler) registerWebhook(payload slack.InteractionCallback, c *socketmode.Client) {
	logger := bslog.With(slog.String("channel_id", payload.View.PrivateMetadata))
	webhook := model.Webhook{
		ChannelID: payload.View.PrivateMetadata,
		Options: model.WebhookOptions{
			Format: "slack",
		},
	}

	for _, option := range payload.View.State.Values["gslb_event_options"]["multi_static_select-action"].SelectedOptions {
		webhook.Subscription.Events = append(webhook.Subscription.Events, option.Value)
	}

	memberOfs := make([]string, 0)
	for memberOf := range strings.SplitSeq(payload.View.State.Values["gslb_memberOfs"]["plain_text_input-action"].Value, ",") {
		memberOfs = append(memberOfs, memberOf)
	}

	webhook.Subscription.Options = map[string]any{
		"memberOfs": memberOfs,
	}

	whRepo.Store(webhook)
	err := sh.broker.Publish(context.Background(), webhook)
	if err != nil {
		logger.Error("failed to publish webhook", slog.String("reason", err.Error()), slog.Any("webhook", webhook))
		if _, _, postErr := c.PostMessage(payload.View.PrivateMetadata, slack.MsgOptionText("Something unexpected happened while processing your request. Please contact #drift-lastbalansering", false)); postErr != nil {
			logger.Error("failed to post error message to channel", slog.String("reason", postErr.Error()))
			return
		}
	}

	respChannel, respTimeStamp, err := c.PostMessage(
		payload.View.PrivateMetadata,
		slack.MsgOptionText("Succesfully sent webhooks registration to gslb-operators. Registration information in :thread:", false),
	)
	if err != nil {
		logger.Error("failed to post registration message to channel", slog.String("reason", err.Error()))
		return
	}

	details := fmt.Sprintf(
		"*Events:*\n%s\n\n*Member-of FQDNs:*\n%s",
		strings.Join(webhook.Subscription.Events, ", "),
		strings.Join(memberOfs, ", "),
	)

	if _, _, err := c.PostMessage(
		respChannel,
		slack.MsgOptionText(details, false),
		slack.MsgOptionTS(respTimeStamp),
	); err != nil {
		logger.Error("failed to post registration details reply", slog.String("reason", err.Error()))
	}

}
