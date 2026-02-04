package events

import (
	"fmt"
	"strings"

	"github.com/waterhouse/resolume-twitch-osc/internal/config"
	"github.com/waterhouse/resolume-twitch-osc/internal/osc"
)

// Handler processes events and sends OSC messages.
type Handler struct {
	osc      *osc.Client
	mappings map[string]config.EventMapping
}

// NewHandler creates a new event handler.
func NewHandler(oscClient *osc.Client, mappings map[string]config.EventMapping) *Handler {
	return &Handler{
		osc:      oscClient,
		mappings: mappings,
	}
}

// Handle processes an event and sends the appropriate OSC message.
// Returns the formatted text that was sent, or empty string if no mapping exists.
func (h *Handler) Handle(event *Event) (string, error) {
	mapping, ok := h.mappings[event.Type]
	if !ok {
		return "", nil // No mapping for this event type
	}

	text := h.formatText(event, mapping.Template)
	if text == "" {
		return "", nil
	}

	if err := h.osc.TriggerClipWithText(mapping.Layer, mapping.Clip, text); err != nil {
		return "", fmt.Errorf("sending OSC for %s: %w", event.Type, err)
	}

	return text, nil
}

// formatText applies the template for the event type.
func (h *Handler) formatText(event *Event, template string) string {
	text := template

	switch e := event.Data.(type) {
	case FollowEvent:
		text = strings.ReplaceAll(text, "{user_name}", e.UserName)

	case SubscribeEvent:
		text = strings.ReplaceAll(text, "{user_name}", e.UserName)
		text = strings.ReplaceAll(text, "{tier}", TierToString(e.Tier))

	case GiftSubEvent:
		userName := e.UserName
		if e.IsAnonymous {
			userName = "Anonymous"
		}
		text = strings.ReplaceAll(text, "{user_name}", userName)
		text = strings.ReplaceAll(text, "{total}", fmt.Sprintf("%d", e.Total))
		text = strings.ReplaceAll(text, "{tier}", TierToString(e.Tier))

	case CheerEvent:
		userName := e.UserName
		if e.IsAnonymous {
			userName = "Anonymous"
		}
		text = strings.ReplaceAll(text, "{user_name}", userName)
		text = strings.ReplaceAll(text, "{bits}", fmt.Sprintf("%d", e.Bits))
		text = strings.ReplaceAll(text, "{message}", e.Message)

	case RaidEvent:
		text = strings.ReplaceAll(text, "{from_broadcaster_user_name}", e.FromBroadcasterUserName)
		text = strings.ReplaceAll(text, "{viewers}", fmt.Sprintf("%d", e.Viewers))

	case ChatEvent:
		text = strings.ReplaceAll(text, "{chatter_user_name}", e.ChatterUserName)
		text = strings.ReplaceAll(text, "{message_text}", e.Message.Text)

	case StreamStartEvent, StreamEndEvent:
		// These templates typically don't have placeholders

	default:
		return ""
	}

	return text
}
