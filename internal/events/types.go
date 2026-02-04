// Package events handles parsing and routing of Twitch WebSocket events.
package events

import "encoding/json"

// Envelope is the common structure for all WebSocket messages.
type Envelope struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

// FollowEvent represents a new follower.
type FollowEvent struct {
	UserID                string `json:"user_id"`
	UserLogin             string `json:"user_login"`
	UserName              string `json:"user_name"`
	BroadcasterUserID     string `json:"broadcaster_user_id"`
	BroadcasterUserLogin  string `json:"broadcaster_user_login"`
	BroadcasterUserName   string `json:"broadcaster_user_name"`
	FollowedAt            string `json:"followed_at"`
}

// SubscribeEvent represents a new subscription.
type SubscribeEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Tier                 string `json:"tier"`
	IsGift               bool   `json:"is_gift"`
}

// GiftSubEvent represents gifted subscriptions.
type GiftSubEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Total                int    `json:"total"`
	Tier                 string `json:"tier"`
	CumulativeTotal      int    `json:"cumulative_total,omitempty"`
	IsAnonymous          bool   `json:"is_anonymous"`
}

// CheerEvent represents a bits cheer.
type CheerEvent struct {
	IsAnonymous          bool   `json:"is_anonymous"`
	UserID               string `json:"user_id,omitempty"`
	UserLogin            string `json:"user_login,omitempty"`
	UserName             string `json:"user_name,omitempty"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Message              string `json:"message"`
	Bits                 int    `json:"bits"`
}

// RaidEvent represents an incoming raid.
type RaidEvent struct {
	FromBroadcasterUserID    string `json:"from_broadcaster_user_id"`
	FromBroadcasterUserLogin string `json:"from_broadcaster_user_login"`
	FromBroadcasterUserName  string `json:"from_broadcaster_user_name"`
	ToBroadcasterUserID      string `json:"to_broadcaster_user_id"`
	ToBroadcasterUserLogin   string `json:"to_broadcaster_user_login"`
	ToBroadcasterUserName    string `json:"to_broadcaster_user_name"`
	Viewers                  int    `json:"viewers"`
}

// ChatMessage represents a chat message (nested structure).
type ChatMessage struct {
	Text string `json:"text"`
}

// ChatEvent represents a chat message event.
type ChatEvent struct {
	BroadcasterUserID    string      `json:"broadcaster_user_id"`
	BroadcasterUserLogin string      `json:"broadcaster_user_login"`
	BroadcasterUserName  string      `json:"broadcaster_user_name"`
	ChatterUserID        string      `json:"chatter_user_id"`
	ChatterUserLogin     string      `json:"chatter_user_login"`
	ChatterUserName      string      `json:"chatter_user_name"`
	MessageID            string      `json:"message_id"`
	Message              ChatMessage `json:"message"`
	Color                string      `json:"color"`
	MessageType          string      `json:"message_type"`
}

// StreamStartEvent represents stream going live.
type StreamStartEvent struct {
	ID                   string `json:"id"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Type                 string `json:"type"`
	StartedAt            string `json:"started_at"`
}

// StreamEndEvent represents stream going offline.
type StreamEndEvent struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

// TierToString converts Twitch tier codes to human-readable strings.
func TierToString(tier string) string {
	switch tier {
	case "1000":
		return "1"
	case "2000":
		return "2"
	case "3000":
		return "3"
	default:
		return tier
	}
}
