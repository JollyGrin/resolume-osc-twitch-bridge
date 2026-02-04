# Project Context

Twitch WebSocket → Resolume OSC bridge. Receives Twitch events, sends OSC to control Resolume Arena text overlays.

## Session Start

1. Read this file (`CLAUDE.md`)
2. Read `.claude/skills.md`
3. Read `PLAN.md` → find next unchecked `[ ]` step
4. Work on that step, check it off `[x]` when done

## Tech Decisions

| Component | Choice | Import Path |
|-----------|--------|-------------|
| OSC | hypebeast/go-osc | `github.com/hypebeast/go-osc/osc` |
| WebSocket | gorilla/websocket | `github.com/gorilla/websocket` |
| TUI | bubbletea | `github.com/charmbracelet/bubbletea` |
| Config | YAML | `gopkg.in/yaml.v3` |

## Project Structure

```
cmd/bridge/      - Main app entry point
cmd/spike/       - OSC connectivity test
internal/config/ - YAML config loading
internal/osc/    - OSC client wrapper
internal/websocket/ - WebSocket client with reconnect
internal/events/ - Event parsing/routing
internal/tui/    - Bubbletea UI
```

## Key OSC Addresses (Resolume)

```
# Trigger clip (Layer 1, Clip 1)
/composition/layers/1/clips/1/connect  → int 1

# Set text on clip's text source
/composition/layers/1/clips/1/video/source/textgenerator/params/text/value → string
```

## WebSocket Events

Source: `wss://twitch-api.waterhousestudios.nl/ws`

Events: `follow`, `subscribe`, `gift_sub`, `cheer`, `raid`, `chat`, `stream_start`, `stream_end`

Envelope: `{"type": "<event>", "data": {...}, "timestamp": number}`

## Constraints

- macOS, same machine as Resolume
- Single user, localhost OSC (127.0.0.1:7000)
- Keep concurrency simple (limited Go experience)
- Config at `./config.yaml`

## Current Phase

MVP complete. All event handlers implemented. Remaining: manual testing and polish (Phase 9).

## Key Files

| File | Purpose |
|------|---------|
| `cmd/bridge/main.go` | Main app - wires config, OSC, WebSocket, TUI |
| `cmd/spike/main.go` | Standalone OSC test |
| `internal/events/handlers.go` | Event → OSC text formatting |
| `internal/events/types.go` | Twitch event structs |
| `internal/websocket/client.go` | Auto-reconnecting WebSocket |
| `internal/tui/model.go` | Bubbletea UI model |
