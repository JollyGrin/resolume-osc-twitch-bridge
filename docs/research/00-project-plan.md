# Project Plan: Twitch WebSocket to Resolume OSC Bridge

## Project Summary

Build a Go CLI application with TUI that receives Twitch events via WebSocket and converts them to OSC commands for Resolume Arena text overlays.

---

## Constraints & Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Language | Go | Specified in README |
| OS | macOS | Development and production |
| Deployment | Same machine as Resolume | Single-user, localhost OSC |
| Concurrency | Keep simple | Limited Go experience |
| Config format | YAML | Infrequent changes, human-readable |
| MVP scope | Follow event only | Then expand to all 8 event types |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        TUI Interface                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ Connection  │  │   Recent    │  │   Test Trigger      │  │
│  │   Status    │  │   Events    │  │     Panel           │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Core Application                        │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │  WebSocket   │───▶│    Event     │───▶│     OSC      │   │
│  │   Client     │    │   Router     │    │   Sender     │   │
│  └──────────────┘    └──────────────┘    └──────────────┘   │
│         │                   │                    │          │
│         ▼                   ▼                    ▼          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │ Reconnect    │    │   Config     │    │  Resolume    │   │
│  │   Logic      │    │   (YAML)     │    │  (UDP:7000)  │   │
│  └──────────────┘    └──────────────┘    └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## Feature Breakdown

### Phase 1: MVP (Follow Event Only)

#### 1.1 WebSocket Client
- Connect to `wss://twitch-api.waterhousestudios.nl/ws`
- Parse JSON message envelope (`type`, `data`, `timestamp`)
- Auto-reconnect with exponential backoff (1s → 30s max)
- Handle ping/pong keepalive

#### 1.2 OSC Sender
- Send UDP packets to `127.0.0.1:7000`
- Support string values (for text)
- Support float/int values (for triggers)

#### 1.3 Event Handler: Follow
- Parse follow event data
- Format text: `"{user_name} just followed!"`
- Send to configured OSC address

#### 1.4 YAML Configuration
```yaml
websocket:
  url: "wss://twitch-api.waterhousestudios.nl/ws"

osc:
  host: "127.0.0.1"
  port: 7000

mappings:
  follow:
    trigger_address: "/composition/layers/1/clips/1/connect"
    text_address: "/composition/layers/1/clips/1/video/source/textgenerator/params/text/value"
    template: "{user_name} just followed!"
```

#### 1.5 Basic TUI
- Connection status indicator (connected/disconnected/reconnecting)
- Recent events log (last 10 events)
- Test trigger button (simulate follow event)

---

### Phase 2: All Event Types

Add handlers for:
- `subscribe` - subscription alerts
- `gift_sub` - gift subscription alerts
- `cheer` - bits/cheer alerts
- `raid` - raid alerts
- `chat` - chat message display
- `stream_start` - going live
- `stream_end` - stream ended

---

### Phase 3: Enhanced TUI (Future)

- Scene/clip selector for each event type
- Live preview of OSC addresses
- Config editing within TUI

---

## Technical Components to Research

| Component | What to Learn | Priority |
|-----------|---------------|----------|
| Go WebSocket client | gorilla/websocket or nhooyr/websocket | High |
| Go OSC library | hypebeast/go-osc or similar | High |
| Go TUI framework | charmbracelet/bubbletea or tview | High |
| YAML parsing | gopkg.in/yaml.v3 | Medium |
| Exponential backoff | Pattern implementation | Medium |

---

## Open Questions

1. **OSC Address Discovery**: How do we know which Resolume layer/clip to target?
   - Manual setup in Resolume first, then configure YAML?
   - Or query Resolume's OSC output to discover addresses?

2. **Text Clip Setup**: Does Resolume need a pre-configured text clip?
   - Likely yes—user creates text source in Resolume, we just update it

3. **Trigger Behavior**: When we "connect" a clip, does it auto-play?
   - Need to test: send `/composition/layers/X/clips/Y/connect` with value 1

4. **Error Visibility**: If OSC fails silently (UDP), how do we know it worked?
   - TUI could show "sent" but not "received"
   - Resolume has OSC monitor for debugging

---

## Next Steps

1. **Research**: Find comparable Go repos for WebSocket→OSC bridges
2. **Research**: Document Go OSC library options
3. **Research**: Document Go TUI framework options
4. **Spike**: Minimal OSC "hello world" to Resolume
5. **Build**: MVP with follow event only
