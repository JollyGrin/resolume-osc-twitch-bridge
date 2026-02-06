# Implementation Plan

> **Session Start**: Read `CLAUDE.md` → `.claude/skills.md` → find next unchecked step below

---

## Phase 0: Project Setup

- [x] **0.1 Initialize Go module**
  ```bash
  go mod init github.com/waterhouse/resolume-twitch-osc
  ```

- [x] **0.2 Create directory structure**
  ```
  cmd/spike/
  cmd/bridge/
  internal/config/
  internal/osc/
  internal/websocket/
  internal/events/
  internal/tui/
  ```

- [x] **0.3 Create config.example.yaml**
  - OSC host/port settings
  - WebSocket URL
  - Event mappings with templates

---

## Phase 1: OSC Spike

Goal: Verify we can send text to Resolume from Go.

- [x] **1.1 Install go-osc dependency**
  ```bash
  go get github.com/hypebeast/go-osc/osc
  ```

- [x] **1.2 Create minimal spike** (`cmd/spike/main.go`)
  - Connect to localhost:7000
  - Send text to Layer 1, Clip 1
  - Send trigger to play clip

  **Research**: [hypebeast/go-osc README](https://github.com/hypebeast/go-osc)
  ```go
  client := osc.NewClient("127.0.0.1", 7000)
  msg := osc.NewMessage("/composition/layers/1/clips/1/video/effects/textblock/effect/text/params/lines")
  msg.Append("Test from Go!")
  client.Send(msg)
  ```

- [x] **1.3 Test spike with Resolume**
  - Resolume running, OSC enabled on port 7000
  - Text clip exists on Layer 1, Clip 1
  - Run spike, verify text appears

---

## Phase 2: Config System

- [x] **2.1 Install yaml dependency**
  ```bash
  go get gopkg.in/yaml.v3
  ```

- [x] **2.2 Create config structs** (`internal/config/config.go`)
  ```go
  type Config struct {
      WebSocket WebSocketConfig
      OSC       OSCConfig
      Mappings  map[string]EventMapping
  }
  ```

- [x] **2.3 Create config loader**
  - Load from `./config.yaml`
  - Validate required fields
  - Return typed config

- [x] **2.4 Create config.example.yaml with all event mappings**

---

## Phase 3: OSC Client Wrapper

- [x] **3.1 Create OSC client** (`internal/osc/client.go`)
  - Wrap hypebeast/go-osc
  - Method: `SendText(address, text string)`
  - Method: `SendTrigger(address string)`

- [x] **3.2 Add error handling**
  - Log send failures (UDP is fire-and-forget, but log anyway)

---

## Phase 4: WebSocket Client

- [x] **4.1 Install gorilla/websocket**
  ```bash
  go get github.com/gorilla/websocket
  ```

- [x] **4.2 Create WebSocket client** (`internal/websocket/client.go`)
  - Connect to configured URL
  - Read-only (no sending)
  - Return messages via channel

  **Research**: [gorilla/websocket examples](https://github.com/gorilla/websocket/tree/main/examples)

- [x] **4.3 Add reconnection logic**
  - Exponential backoff: 1s → 2s → 4s → ... → 30s max
  - Emit connection state changes

  **Research**: Simple pattern without external deps
  ```go
  delay := time.Second
  for {
      err := connect()
      if err == nil { delay = time.Second; continue }
      time.Sleep(delay)
      delay = min(delay*2, 30*time.Second)
  }
  ```

- [x] **4.4 Handle ping/pong keepalive**
  - Server sends ping every 54s
  - Client must respond with pong

---

## Phase 5: Event System

- [x] **5.1 Create event types** (`internal/events/types.go`)
  - Base envelope: `type`, `data`, `timestamp`
  - Follow event struct

  **Research**: See `docs/draft_websocket-to-resolume.md` for schemas

- [x] **5.2 Create event parser** (`internal/events/parser.go`)
  - Parse JSON envelope
  - Route by `type` field
  - Return typed event

- [x] **5.3 Create event handler for Follow** (`internal/events/handlers.go`)
  - Extract `user_name` from follow event
  - Apply template: `"{user_name} just followed!"`
  - Call OSC client to send text + trigger

- [ ] **5.4 Test end-to-end: WebSocket → Event → OSC**
  - Use test endpoint or mock
  - Verify text appears in Resolume

---

## Phase 6: Basic TUI

- [x] **6.1 Install bubbletea**
  ```bash
  go get github.com/charmbracelet/bubbletea
  go get github.com/charmbracelet/bubbles
  go get github.com/charmbracelet/lipgloss
  ```

- [x] **6.2 Create TUI model** (`internal/tui/model.go`)
  ```go
  type model struct {
      connectionStatus string  // "connected", "disconnected", "reconnecting"
      recentEvents     []string // Last 10 events
      err              error
  }
  ```

  **Research**: [bubbletea realtime example](https://github.com/charmbracelet/bubbletea/blob/main/examples/realtime/main.go)

- [x] **6.3 Create TUI view**
  - Connection status with color indicator
  - Scrollable event log
  - Quit instruction (q to quit)

- [x] **6.4 Wire WebSocket events to TUI**
  - Channel from WebSocket client → TUI updates
  - Pattern: `waitForActivity()` command

- [x] **6.5 Add test trigger**
  - Press 't' to send test follow event
  - Useful for testing without real Twitch events

---

## Phase 7: Main Application

- [x] **7.1 Create main entrypoint** (`cmd/bridge/main.go`)
  - Load config
  - Initialize OSC client
  - Initialize WebSocket client
  - Initialize TUI
  - Wire everything together

- [x] **7.2 Add graceful shutdown**
  - Handle Ctrl+C
  - Close WebSocket cleanly

- [ ] **7.3 Test full flow**
  - Start app
  - Trigger test event from Twitch debug panel
  - Verify text appears in Resolume

---

## Phase 8: Remaining Events

- [x] **8.1 Add subscribe event handler**
  - Template: `"{user_name} subscribed (Tier {tier})!"`

- [x] **8.2 Add gift_sub event handler**
  - Template: `"{user_name} gifted {total} subs!"`
  - Handle anonymous gifters

- [x] **8.3 Add cheer event handler**
  - Template: `"{user_name} cheered {bits} bits!"`
  - Handle anonymous cheers

- [x] **8.4 Add raid event handler**
  - Template: `"{from_broadcaster_user_name} raided with {viewers} viewers!"`

- [x] **8.5 Add chat event handler** (optional for MVP)
  - Template: `"{chatter_user_name}: {message.text}"`

- [x] **8.6 Add stream_start/stream_end handlers**

---

## Phase 9: Polish

- [ ] **9.1 Add logging**
  - Log file for debugging
  - Configurable log level

- [ ] **9.2 Update README with final instructions**

- [ ] **9.3 Create release binary**
  ```bash
  go build -o bin/resolume-bridge ./cmd/bridge
  ```

---

## Research References

| Topic | Link |
|-------|------|
| go-osc usage | https://github.com/hypebeast/go-osc |
| gorilla/websocket | https://github.com/gorilla/websocket |
| bubbletea realtime | https://github.com/charmbracelet/bubbletea/blob/main/examples/realtime/main.go |
| bubbletea + websocket | https://github.com/Askath/Go-ChatAndClient |
| Resolume text OSC | https://github.com/anthonyeden/ProPresenter-Resolume |
| WebSocket event schemas | `docs/draft_websocket-to-resolume.md` |
| Resolume setup | `docs/resolume-setup.md` |
