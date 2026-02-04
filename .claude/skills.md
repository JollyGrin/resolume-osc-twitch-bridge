# Skills

## /spike

Test OSC connectivity to Resolume. Sends "Test from Go!" to Layer 1, Clip 1.

**Prerequisites**: Resolume running with OSC enabled on port 7000, text clip on Layer 1 Clip 1.

```bash
go run ./cmd/spike
```

## /run

Run the main bridge application with TUI.

**Prerequisites**: `config.yaml` exists (copy from `config.example.yaml`).

```bash
go run ./cmd/bridge
```

**TUI Controls**: `t` = test event, `q` = quit

## /build

Build the bridge binary.

```bash
go build -o bin/bridge ./cmd/bridge
```

## /test-ws

Test WebSocket connection to Twitch backend.

```bash
# Local dev
wscat -c ws://localhost:8080/ws

# Production
wscat -c wss://twitch-api.waterhousestudios.nl/ws
```

## /resolume-setup

Show Resolume setup instructions.

Read `docs/resolume-setup.md` for:
- Enabling OSC in Resolume preferences
- Creating a text clip on Layer 1
- Finding OSC addresses

## /add-event

Add a new event type handler:

1. Add struct to `internal/events/types.go`
2. Add case to `Parse()` in `internal/events/parser.go`
3. Add case to `formatText()` in `internal/events/handlers.go`
4. Add mapping to `config.example.yaml`
