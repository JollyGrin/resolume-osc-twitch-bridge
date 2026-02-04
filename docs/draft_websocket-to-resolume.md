# Twitch WebSocket to Resolume OSC Integration

## Overview

This document describes the WebSocket message structure from the Twitch overlay backend, intended for building a local application that converts these events into OSC signals for Resolume Arena/Avenue to display custom text overlays.

---

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐     ┌──────────────┐
│  Twitch API     │────▶│  Go Backend      │────▶│  Local OSC App  │────▶│  Resolume    │
│  (EventSub)     │     │  (WebSocket)     │     │  (To Be Built)  │     │  Arena       │
└─────────────────┘     └──────────────────┘     └─────────────────┘     └──────────────┘
```

### Connection Details

| Environment | WebSocket URL |
|-------------|---------------|
| Development | `ws://localhost:8080/ws` |
| Production  | `wss://twitch-api.waterhousestudios.nl/ws` |

**Connection Behavior:**
- WebSocket is **receive-only** (client cannot send commands)
- Ping/pong keepalive every 54 seconds
- Auto-reconnect with exponential backoff (1s → 2s → 4s → ... → 30s max)

---

## Message Envelope

Every WebSocket message follows this structure:

```json
{
  "type": "<event_type>",
  "data": { ... },
  "timestamp": 1738675496
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Event type identifier (see below) |
| `data` | object | Event-specific payload |
| `timestamp` | number | Unix timestamp (seconds since epoch) |

---

## Event Types Reference

### Summary Table

| Type | Twitch Trigger | Key Data Fields | Suggested Resolume Use |
|------|----------------|-----------------|------------------------|
| `chat` | Message in chat | `chatter_user_name`, `message.text`, `color` | Scrolling chat, highlighted messages |
| `follow` | New follower | `user_name`, `followed_at` | "Thanks for following!" alert |
| `subscribe` | New subscription | `user_name`, `tier` | Sub alert with tier badge |
| `gift_sub` | Gifted subs | `user_name`, `total`, `tier` | Gift bomb animation |
| `cheer` | Bits donation | `user_name`, `bits`, `message` | Bit alert with amount |
| `raid` | Incoming raid | `from_broadcaster_user_name`, `viewers` | Raid alert with viewer count |
| `stream_start` | Stream goes live | `broadcaster_user_name`, `started_at` | "We're live!" graphic |
| `stream_end` | Stream ends | `broadcaster_user_name` | "Stream ended" graphic |

---

## Detailed Event Schemas

### 1. Chat Message (`chat`)

Triggered when someone sends a message in chat.

```typescript
{
  type: "chat",
  data: {
    // Broadcaster info
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string,

    // Chatter info
    chatter_user_id: string,
    chatter_user_login: string,
    chatter_user_name: string,

    // Message content
    message_id: string,
    message: {
      text: string,
      fragments: Array<{
        type: string,        // "text", "emote", "cheermote", etc.
        text: string,
        emote?: {
          id: string,
          emote_set_id: string
        }
      }>
    },

    // Styling
    color: string,           // Hex color, e.g. "#FF0000"

    // Badges (subscriber, moderator, etc.)
    badges: Array<{
      set_id: string,        // e.g. "subscriber", "moderator"
      id: string,            // badge tier/version
      info: string
    }>,

    // Message metadata
    message_type: string,    // "text", "channel_points_highlighted", etc.

    // Optional fields
    cheer?: {
      bits: number
    } | null,

    reply?: {
      parent_message_id: string,
      parent_message_body: string,
      parent_user_id: string,
      parent_user_name: string,
      parent_user_login: string,
      thread_message_id: string,
      thread_user_id: string,
      thread_user_name: string,
      thread_user_login: string
    } | null,

    channel_points_custom_reward_id?: string | null
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "chat",
  "data": {
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios",
    "chatter_user_id": "12345678",
    "chatter_user_name": "CoolViewer",
    "message": {
      "text": "Hello everyone!",
      "fragments": [
        { "type": "text", "text": "Hello everyone!" }
      ]
    },
    "color": "#8A2BE2",
    "badges": [
      { "set_id": "subscriber", "id": "3", "info": "3" }
    ],
    "message_type": "text"
  },
  "timestamp": 1738675496
}
```

**OSC Mapping Suggestions:**
- `/resolume/text/chat/username` → `chatter_user_name`
- `/resolume/text/chat/message` → `message.text`
- `/resolume/color/chat` → `color` (convert hex to RGB)

---

### 2. Follow Event (`follow`)

Triggered when someone follows the channel.

```typescript
{
  type: "follow",
  data: {
    user_id: string,
    user_login: string,
    user_name: string,
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string,
    followed_at: string       // ISO 8601, e.g. "2025-02-04T12:34:56Z"
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "follow",
  "data": {
    "user_id": "87654321",
    "user_login": "newfollower",
    "user_name": "NewFollower",
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios",
    "followed_at": "2025-02-04T12:34:56Z"
  },
  "timestamp": 1738675496
}
```

**OSC Mapping Suggestions:**
- `/resolume/text/follow/username` → `user_name`
- `/resolume/trigger/follow` → 1 (trigger clip)

---

### 3. Subscribe Event (`subscribe`)

Triggered when someone subscribes (non-gift).

```typescript
{
  type: "subscribe",
  data: {
    user_id: string,
    user_login: string,
    user_name: string,
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string,
    tier: string,             // "1000" = Tier 1, "2000" = Tier 2, "3000" = Tier 3
    is_gift: boolean
  },
  timestamp: number
}
```

**Tier Values:**
| Tier Value | Meaning | Price |
|------------|---------|-------|
| `"1000"` | Tier 1 | $4.99 |
| `"2000"` | Tier 2 | $9.99 |
| `"3000"` | Tier 3 | $24.99 |

**Example:**
```json
{
  "type": "subscribe",
  "data": {
    "user_id": "11111111",
    "user_name": "LoyalFan",
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios",
    "tier": "1000",
    "is_gift": false
  },
  "timestamp": 1738675500
}
```

**OSC Mapping Suggestions:**
- `/resolume/text/sub/username` → `user_name`
- `/resolume/text/sub/tier` → `tier` (or convert to "Tier 1", "Tier 2", "Tier 3")
- `/resolume/trigger/sub` → 1

---

### 4. Gift Subscription Event (`gift_sub`)

Triggered when someone gifts subscriptions.

```typescript
{
  type: "gift_sub",
  data: {
    user_id: string,
    user_login: string,
    user_name: string,
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string,
    total: number,            // Number of subs gifted in this event
    tier: string,             // "1000", "2000", or "3000"
    cumulative_total?: number, // Total gifts from this user all-time
    is_anonymous: boolean
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "gift_sub",
  "data": {
    "user_id": "22222222",
    "user_name": "GenerousGifter",
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios",
    "total": 5,
    "tier": "1000",
    "cumulative_total": 50,
    "is_anonymous": false
  },
  "timestamp": 1738675505
}
```

**OSC Mapping Suggestions:**
- `/resolume/text/giftsub/username` → `user_name` (or "Anonymous" if `is_anonymous`)
- `/resolume/text/giftsub/count` → `total`
- `/resolume/trigger/giftsub` → 1

---

### 5. Cheer Event (`cheer`)

Triggered when someone cheers with bits.

```typescript
{
  type: "cheer",
  data: {
    is_anonymous: boolean,
    user_id?: string,         // Absent if anonymous
    user_login?: string,
    user_name?: string,
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string,
    message: string,          // The chat message with the cheer
    bits: number              // Number of bits cheered
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "cheer",
  "data": {
    "is_anonymous": false,
    "user_id": "33333333",
    "user_name": "BitsBoss",
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios",
    "message": "cheer100 Great stream!",
    "bits": 100
  },
  "timestamp": 1738675510
}
```

**OSC Mapping Suggestions:**
- `/resolume/text/cheer/username` → `user_name` (or "Anonymous")
- `/resolume/text/cheer/bits` → `bits`
- `/resolume/text/cheer/message` → `message`
- `/resolume/trigger/cheer` → 1

---

### 6. Raid Event (`raid`)

Triggered when another streamer raids the channel.

```typescript
{
  type: "raid",
  data: {
    from_broadcaster_user_id: string,
    from_broadcaster_user_login: string,
    from_broadcaster_user_name: string,
    to_broadcaster_user_id: string,
    to_broadcaster_user_login: string,
    to_broadcaster_user_name: string,
    viewers: number           // Number of viewers in the raid
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "raid",
  "data": {
    "from_broadcaster_user_id": "44444444",
    "from_broadcaster_user_name": "FriendlyStreamer",
    "to_broadcaster_user_id": "1299430554",
    "to_broadcaster_user_name": "waterhousestudios",
    "viewers": 150
  },
  "timestamp": 1738675515
}
```

**OSC Mapping Suggestions:**
- `/resolume/text/raid/username` → `from_broadcaster_user_name`
- `/resolume/text/raid/viewers` → `viewers`
- `/resolume/trigger/raid` → 1

---

### 7. Stream Start Event (`stream_start`)

Triggered when the stream goes live.

```typescript
{
  type: "stream_start",
  data: {
    id: string,               // Stream ID
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string,
    type: string,             // Usually "live"
    started_at: string        // ISO 8601 timestamp
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "stream_start",
  "data": {
    "id": "stream123456",
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios",
    "type": "live",
    "started_at": "2025-02-04T18:00:00Z"
  },
  "timestamp": 1738695600
}
```

---

### 8. Stream End Event (`stream_end`)

Triggered when the stream goes offline.

```typescript
{
  type: "stream_end",
  data: {
    broadcaster_user_id: string,
    broadcaster_user_login: string,
    broadcaster_user_name: string
  },
  timestamp: number
}
```

**Example:**
```json
{
  "type": "stream_end",
  "data": {
    "broadcaster_user_id": "1299430554",
    "broadcaster_user_name": "waterhousestudios"
  },
  "timestamp": 1738702800
}
```

---

## OSC Integration Notes

### Resolume OSC Basics

Resolume Arena/Avenue listens for OSC messages on a configurable port (default: **7000**).

**Common OSC Address Patterns:**
- `/composition/layers/[layer]/clips/[clip]/connect` → Trigger clip (value: 1)
- `/composition/layers/[layer]/video/opacity` → Set opacity (0.0-1.0)
- `/composition/selectedclip/video/source/textgenerator/text` → Set text on selected text clip

### Suggested Application Flow

```
1. Connect to WebSocket
   ↓
2. Parse incoming JSON message
   ↓
3. Extract event type and relevant data
   ↓
4. Map to OSC address + value
   ↓
5. Send OSC packet to Resolume
```

### Example OSC Mappings

```
Event: follow
  → /composition/layers/1/clips/1/connect = 1      (trigger follow alert clip)
  → /composition/layers/1/clips/1/video/source/textgenerator/params/text/value = "NewFollower just followed!"

Event: cheer (100 bits)
  → /composition/layers/2/clips/1/connect = 1      (trigger cheer clip)
  → /composition/layers/2/clips/1/video/source/textgenerator/params/text/value = "BitsBoss cheered 100 bits!"

Event: raid (150 viewers)
  → /composition/layers/3/clips/1/connect = 1      (trigger raid clip)
  → /composition/layers/3/clips/1/video/source/textgenerator/params/text/value = "FriendlyStreamer raided with 150 viewers!"
```

### Text Formatting Templates

Suggested text templates for each event type:

| Event | Template |
|-------|----------|
| `follow` | `"{user_name} just followed!"` |
| `subscribe` | `"{user_name} subscribed (Tier {tier})!"` |
| `gift_sub` | `"{user_name} gifted {total} subs!"` |
| `cheer` | `"{user_name} cheered {bits} bits!"` |
| `raid` | `"{from_broadcaster_user_name} raided with {viewers} viewers!"` |
| `chat` | `"{chatter_user_name}: {message.text}"` |

---

## Implementation Checklist

### WebSocket Client Requirements

- [ ] Connect to WebSocket URL (dev or prod)
- [ ] Handle connection errors gracefully
- [ ] Implement auto-reconnect with exponential backoff
- [ ] Parse JSON messages
- [ ] Route by `type` field to appropriate handler

### OSC Client Requirements

- [ ] Configure Resolume OSC port (default 7000)
- [ ] Support sending string values (for text)
- [ ] Support sending float values (for triggers/opacity)
- [ ] Handle OSC bundle batching (optional, for performance)

### Event Handlers

- [ ] `chat` → Extract username, message, color
- [ ] `follow` → Extract username
- [ ] `subscribe` → Extract username, tier
- [ ] `gift_sub` → Extract username, count, handle anonymous
- [ ] `cheer` → Extract username, bits, message, handle anonymous
- [ ] `raid` → Extract raider name, viewer count
- [ ] `stream_start` → Trigger "going live" graphic
- [ ] `stream_end` → Trigger "stream ended" graphic

### Configuration Options (Suggested)

```yaml
websocket:
  url: "wss://twitch-api.waterhousestudios.nl/ws"
  reconnect_delay_ms: 1000
  max_reconnect_delay_ms: 30000

osc:
  host: "127.0.0.1"
  port: 7000

mappings:
  follow:
    trigger_address: "/composition/layers/1/clips/1/connect"
    text_address: "/composition/layers/1/clips/1/video/source/textgenerator/params/text/value"
    template: "{user_name} just followed!"

  subscribe:
    trigger_address: "/composition/layers/1/clips/2/connect"
    text_address: "/composition/layers/1/clips/2/video/source/textgenerator/params/text/value"
    template: "{user_name} subscribed (Tier {tier})!"

  # ... etc
```

---

## Testing

### Debug Mode

Add `?test=true` to the overlay URL to use local WebSocket:
```
http://localhost:5173/?test=true
```

### Debug Panel

The backend provides a debug panel at `/debug` (when running in debug mode) that lets you trigger test events:

```bash
# Start backend in debug mode
DEBUG_MODE=true go run cmd/server/main.go
```

Then visit: `http://localhost:8080/debug`

### Manual Test Event

```bash
curl -X POST http://localhost:8080/debug/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "type": "follow",
    "data": {
      "user_name": "TestFollower",
      "user_id": "12345",
      "broadcaster_user_name": "waterhousestudios",
      "broadcaster_user_id": "1299430554",
      "followed_at": "2025-02-04T12:00:00Z"
    }
  }'
```

---

## Appendix: TypeScript Type Definitions

For reference, here are the complete TypeScript types from the client overlay:

```typescript
// Base message envelope
interface BaseMessage {
  type: string;
  data: unknown;
  timestamp: number;
}

// Union of all event types
type TwitchEvent =
  | ChatMessage
  | FollowEvent
  | RaidEvent
  | SubscribeEvent
  | GiftSubEvent
  | CheerEvent
  | StreamOnlineEvent
  | StreamOfflineEvent;

// Type guard functions available:
function isChatMessage(event: TwitchEvent): event is ChatMessage
function isFollowEvent(event: TwitchEvent): event is FollowEvent
function isRaidEvent(event: TwitchEvent): event is RaidEvent
function isSubscribeEvent(event: TwitchEvent): event is SubscribeEvent
function isGiftSubEvent(event: TwitchEvent): event is GiftSubEvent
function isCheerEvent(event: TwitchEvent): event is CheerEvent
function isStreamOnlineEvent(event: TwitchEvent): event is StreamOnlineEvent
function isStreamOfflineEvent(event: TwitchEvent): event is StreamOfflineEvent
```

---

## Source Files Reference

| File | Description |
|------|-------------|
| `backend/cmd/server/main.go` | Server entry point, WebSocket setup |
| `backend/internal/websocket/hub.go` | WebSocket hub, broadcast logic |
| `backend/internal/websocket/client.go` | WebSocket client connection handling |
| `backend/internal/twitch/webhook.go` | EventSub webhook receiver |
| `backend/internal/twitch/client.go` | Twitch API integration |
| `client-overlay/src/lib/types/twitch.ts` | TypeScript type definitions |
| `client-overlay/src/lib/stores/websocket.ts` | Client WebSocket connection |
