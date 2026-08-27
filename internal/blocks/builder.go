package blocks

import (
	"fmt"
	"strings"

	"github.com/NorskHelsenett/gslb-event-bot/internal/config"
	"github.com/NorskHelsenett/gslb-event-bot/internal/repository/webhooks"
	"github.com/slack-go/slack"
)

type SlackMessageBuilder struct{}

func NewSlackMessageBuilder() *SlackMessageBuilder {
	return &SlackMessageBuilder{}
}

func (sb *SlackMessageBuilder) NewSubscriptionModal(id string) (slack.ModalViewRequest, error) {
	hook, err := webhooks.Get(id)
	if err != nil {
		return slack.ModalViewRequest{}, fmt.Errorf("failed to fetch webhook: %w", err)
	}

	events := config.Webhooks().Events()

	options := make([]*slack.OptionBlockObject, 0, len(events))
	for _, event := range events {
		options = append(
			options,
			slack.NewOptionBlockObject(event, slack.NewTextBlockObject(slack.PlainTextType, event, false, false), nil),
		)
	}

	multiSelect := slack.NewOptionsMultiSelectBlockElement(
		slack.MultiOptTypeStatic,
		slack.NewTextBlockObject(slack.PlainTextType, "Select events", false, false),
		"multi_static_select-action",
		options...,
	)

	var initialValue string
	initialOptions := make([]*slack.OptionBlockObject, 0, len(hook.Subscription.Events))
	if raw, ok := hook.Subscription.Options["memberOfs"]; ok {
		initialValue = strings.Join(toStringSlice(raw), ",")
	}

	for _, event := range hook.Subscription.Events {
		initialOptions = append(
			initialOptions,
			slack.NewOptionBlockObject(event, slack.NewTextBlockObject(slack.PlainTextType, event, false, false), nil),
		)
	}
	multiSelect.InitialOptions = initialOptions

	multiLineMemberOfs := slack.PlainTextInputBlockElement{
		Type:         slack.METPlainTextInput,
		ActionID:     "plain_text_input-action",
		Multiline:    true,
		InitialValue: initialValue,
		Placeholder:  slack.NewTextBlockObject(slack.PlainTextType, "Comma-separated list of FQDNs, e.g. host1.nhn.no,host2.nhn.no", false, false),
	}

	return slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: "gslb_events_subscription",
		Title:      slack.NewTextBlockObject(slack.PlainTextType, "Events subscription", true, false),
		Close:      slack.NewTextBlockObject(slack.PlainTextType, "Cancel", true, false),
		Submit:     slack.NewTextBlockObject(slack.PlainTextType, "Submit", true, false),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewInputBlock(
					"gslb_event_options",
					slack.NewTextBlockObject(slack.PlainTextType, "Select GSLB events", false, false),
					nil,
					multiSelect,
				).WithOptional(false),
				slack.NewInputBlock(
					"gslb_memberOfs",
					slack.NewTextBlockObject(slack.PlainTextType, "Enter member-of FQDNs", false, false),
					nil,
					multiLineMemberOfs,
				).WithOptional(true),
			},
		},
	}, nil
}

// toStringSlice normalizes memberOfs stored as []string (in-memory) or []interface{} (JSON-decoded persisted data).
func toStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}
