package events

import (
	"fmt"
	"strings"

	"github.com/waterhouse/resolume-twitch-osc/internal/config"
	"github.com/waterhouse/resolume-twitch-osc/internal/osc"
)

// Handler processes events and sends OSC messages.
type Handler struct {
	osc             *osc.Client
	mappings        map[string]config.EventMapping
	defaults        config.Defaults
	debounceManager *DebounceManager
}

// NewHandler creates a new event handler.
func NewHandler(oscClient *osc.Client, mappings map[string]config.EventMapping, defaults config.Defaults) *Handler {
	return &Handler{
		osc:             oscClient,
		mappings:        mappings,
		defaults:        defaults,
		debounceManager: NewDebounceManager(oscClient, defaults.Group, defaults.SoloOnEvent),
	}
}

// Close cleans up resources. Call on shutdown.
func (h *Handler) Close() {
	h.debounceManager.CancelAll()
}

// Handle processes an event and sends the appropriate OSC messages.
// Returns the formatted text from the first action with a template, or empty string if no mapping exists.
func (h *Handler) Handle(event *Event) (string, error) {
	mapping, ok := h.mappings[event.Type]
	if !ok {
		return "", nil // No mapping for this event type
	}

	// Solo group for non-chat events that should return to scene
	if event.Type != "chat" && mapping.ShouldReturnToScene() {
		h.debounceManager.SoloGroupIfNeeded()
	}

	var firstText string
	var triggeredClips []struct{ layer, clip int }

	for _, action := range mapping.Actions {
		if action.Template != "" {
			text := h.formatText(event, action.Template)
			if text == "" {
				continue
			}
			if firstText == "" {
				firstText = text
			}
			if err := h.osc.TriggerClipWithText(action.Layer, action.Clip, text); err != nil {
				return "", fmt.Errorf("sending OSC for %s: %w", event.Type, err)
			}
		} else {
			// No template, just trigger the clip
			if err := h.osc.TriggerClip(action.Layer, action.Clip); err != nil {
				return "", fmt.Errorf("triggering clip for %s: %w", event.Type, err)
			}
		}
		triggeredClips = append(triggeredClips, struct{ layer, clip int }{action.Layer, action.Clip})
	}

	// Schedule debounce for triggered clips if return_to_scene is enabled
	if mapping.ShouldReturnToScene() {
		debounce := mapping.GetDebounce(h.defaults.Debounce)
		for _, tc := range triggeredClips {
			h.debounceManager.Schedule(tc.layer, tc.clip, debounce)
		}
	}

	return firstText, nil
}

// TriggerTestEvent triggers a test event with proper solo/debounce behavior.
// Returns the text that was displayed.
func (h *Handler) TriggerTestEvent(layer, clip int, text string) (string, error) {
	// Solo and unbypass the group
	h.debounceManager.SoloGroupIfNeeded()

	// Trigger the clip with text
	if err := h.osc.TriggerClipWithText(layer, clip, text); err != nil {
		return "", fmt.Errorf("triggering test clip: %w", err)
	}

	// Schedule debounce
	h.debounceManager.Schedule(layer, clip, h.defaults.Debounce.Duration())

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
