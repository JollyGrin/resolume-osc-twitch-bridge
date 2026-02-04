package events

import (
	"encoding/json"
	"fmt"
)

// Event is a parsed event with its type and data.
type Event struct {
	Type      string
	Timestamp int64
	Data      any
}

// Parse parses a raw WebSocket message into a typed Event.
func Parse(data []byte) (*Event, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parsing envelope: %w", err)
	}

	event := &Event{
		Type:      env.Type,
		Timestamp: env.Timestamp,
	}

	var err error
	switch env.Type {
	case "follow":
		var e FollowEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "subscribe":
		var e SubscribeEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "gift_sub":
		var e GiftSubEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "cheer":
		var e CheerEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "raid":
		var e RaidEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "chat":
		var e ChatEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "stream_start":
		var e StreamStartEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	case "stream_end":
		var e StreamEndEvent
		err = json.Unmarshal(env.Data, &e)
		event.Data = e
	default:
		// Unknown event type - store raw data
		event.Data = env.Data
	}

	if err != nil {
		return nil, fmt.Errorf("parsing %s data: %w", env.Type, err)
	}

	return event, nil
}
