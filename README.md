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

## Configuration

Configuration is stored in `config.yaml` in the same directory as the binary:

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

## Development

### Project Structure

```
.
├── cmd/
│   └── bridge/          # Main application entry point
├── internal/
│   ├── config/          # YAML config loading
│   ├── osc/             # OSC client wrapper
│   ├── websocket/       # WebSocket client
│   ├── events/          # Event parsing and routing
│   └── tui/             # Bubbletea TUI
├── config.yaml          # Runtime configuration
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
