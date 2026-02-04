# Twitch to Resolume OSC Bridge

A Go CLI application that receives Twitch events via WebSocket and converts them to OSC commands for Resolume Arena text overlays.

## Features

- Receives Twitch events (follow, subscribe, raid, cheer, etc.)
- Converts events to OSC messages for Resolume
- Updates text sources with event information
- TUI interface showing connection status and recent events

## Prerequisites

- Go 1.21+
- Resolume Arena/Avenue with OSC enabled
- Access to Twitch WebSocket feed

## Quick Start

### 1. Set Up Resolume

Enable OSC in Resolume and create a text clip:

1. **Enable OSC**: Resolume → Preferences → OSC → Enable Input on port `7000`
2. **Create text clip**: Add a Text Block source to Layer 1, Clip 1

See [docs/resolume-setup.md](docs/resolume-setup.md) for detailed instructions.

### 2. Configure the Bridge

Copy the example config and edit:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` with your settings.

### 3. Run the Application

```bash
go run ./cmd/bridge
```

### TUI Controls

| Key | Action |
|-----|--------|
| `t` | Send test event to Resolume |
| `q` | Quit |

## Configuration

Configuration is stored in `config.yaml` in the project directory:

```yaml
websocket:
  url: "wss://twitch-api.waterhousestudios.nl/ws"

osc:
  host: "127.0.0.1"
  port: 7000

mappings:
  follow:
    layer: 1
    clip: 1
    template: "{user_name} just followed!"
  subscribe:
    layer: 1
    clip: 1
    template: "{user_name} subscribed (Tier {tier})!"
  # See config.example.yaml for all event mappings
```

### Event Templates

Templates support these placeholders:

| Event | Placeholders |
|-------|-------------|
| `follow` | `{user_name}` |
| `subscribe` | `{user_name}`, `{tier}` |
| `gift_sub` | `{user_name}`, `{total}`, `{tier}` |
| `cheer` | `{user_name}`, `{bits}`, `{message}` |
| `raid` | `{from_broadcaster_user_name}`, `{viewers}` |
| `chat` | `{chatter_user_name}`, `{message_text}` |

## Development

### Project Structure

```
.
├── cmd/
│   ├── bridge/          # Main application entry point
│   └── spike/           # OSC connectivity test
├── internal/
│   ├── config/          # YAML config loading
│   ├── osc/             # OSC client wrapper
│   ├── websocket/       # WebSocket client with auto-reconnect
│   ├── events/          # Event parsing, types, and handlers
│   └── tui/             # Bubbletea TUI
├── config.yaml          # Runtime configuration (not in git)
├── config.example.yaml  # Example configuration
└── docs/
    ├── resolume-setup.md
    └── research/
```

### Testing OSC Connection

Run the spike to verify OSC connectivity:

```bash
go run ./cmd/spike
```

This sends a test message to Resolume without needing the WebSocket connection.

## Documentation

- [Resolume Setup Guide](docs/resolume-setup.md) - How to configure Resolume for OSC
- [WebSocket Protocol](docs/draft_websocket-to-resolume.md) - Twitch event message formats
- [Research Notes](docs/research/) - Technical research and decisions

## Tech Stack

- **Language**: Go
- **TUI**: [Bubbletea](https://github.com/charmbracelet/bubbletea)
- **OSC**: [hypebeast/go-osc](https://github.com/hypebeast/go-osc)
- **WebSocket**: [gorilla/websocket](https://github.com/gorilla/websocket)
