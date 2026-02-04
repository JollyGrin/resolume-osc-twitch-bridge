# Skills

## /spike

Test OSC connectivity to Resolume. Sends a test message to Layer 1, Clip 1.

**Prerequisites**: Resolume running with OSC enabled on port 7000, text clip on Layer 1 Clip 1.

```bash
go run ./cmd/spike
```

## /run

Run the main bridge application.

```bash
go run ./cmd/bridge
```

## /build

Build the bridge binary.

```bash
go build -o bin/bridge ./cmd/bridge
```

## /test-ws

Test WebSocket connection to Twitch backend (dev mode).

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
