# Research: Relevant GitHub Repositories

This document catalogs GitHub repos relevant to building our Twitch WebSocket → Resolume OSC bridge.

---

## Summary Table

| Category | Recommended | Alternative | Notes |
|----------|-------------|-------------|-------|
| OSC Library | hypebeast/go-osc | scgolang/osc | hypebeast is most used |
| WebSocket | gorilla/websocket | coder/websocket | gorilla is battle-tested |
| TUI Framework | charmbracelet/bubbletea | tview | bubbletea is modern/idiomatic |
| TUI Components | charmbracelet/bubbles | - | Spinner, textinput, list, etc. |

---

## 1. Go OSC Libraries

### hypebeast/go-osc ⭐ RECOMMENDED
- **URL**: https://github.com/hypebeast/go-osc
- **Stars**: ~500+
- **Status**: Mature, stable
- **Relevance**: PRIMARY - This is the standard Go OSC library

**Key Features**:
- Pure Go implementation (no CGO)
- Supports OSC 1.0 spec
- UDP transport
- Message types: int32, float32, string, blob, int64, timetag, double, true/false/nil
- Bundle support
- Address pattern matching with wildcards

**Example Usage**:
```go
import "github.com/hypebeast/go-osc/osc"

// Create client
client := osc.NewClient("127.0.0.1", 7000)

// Create and send message
msg := osc.NewMessage("/composition/layers/1/clips/1/video/source/textgenerator/params/text/value")
msg.Append("Hello from Go!")
client.Send(msg)
```

**Why Recommended**: Most documented, most used, simple API, covers all our needs.

---

### scgolang/osc
- **URL**: https://github.com/scgolang/osc
- **Status**: Active
- **Relevance**: Alternative if hypebeast has issues

**Pros**: Aims for full OSC 1.0 compliance
**Cons**: Less community adoption, fewer examples

---

## 2. Go WebSocket Libraries

### gorilla/websocket ⭐ RECOMMENDED
- **URL**: https://github.com/gorilla/websocket
- **Stars**: 22k+
- **Status**: Mature (6+ years), stable
- **Relevance**: PRIMARY - Battle-tested, widely used

**Key Features**:
- Fast and well-tested
- Supports text and binary messages
- Ping/pong handling
- Compression support
- Clear documentation

**Example Usage**:
```go
import "github.com/gorilla/websocket"

// Connect
dialer := websocket.Dialer{}
conn, _, err := dialer.Dial("wss://example.com/ws", nil)

// Read messages
for {
    _, message, err := conn.ReadMessage()
    // handle message
}
```

**Why Recommended**: Most documentation, most Stack Overflow answers, proven stability.

---

### coder/websocket (formerly nhooyr/websocket)
- **URL**: https://github.com/coder/websocket
- **Status**: Actively maintained (took over from nhooyr in 2024)
- **Relevance**: Modern alternative

**Pros**: More idiomatic Go API, context support
**Cons**: Newer, less community examples

---

## 3. TUI Frameworks

### charmbracelet/bubbletea ⭐ RECOMMENDED
- **URL**: https://github.com/charmbracelet/bubbletea
- **Stars**: 30k+
- **Status**: Very active, modern
- **Relevance**: PRIMARY - Best Go TUI framework

**Key Features**:
- Elm-inspired architecture (Model, Update, View)
- Composable components
- Great for real-time updates
- Beautiful output with lipgloss styling

**Architecture Pattern**:
```go
type model struct {
    status    string
    events    []string
    quitting  bool
}

func (m model) Init() tea.Cmd { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* handle msgs */ }
func (m model) View() string { return /* render UI */ }
```

**Why Recommended**: Modern, actively maintained, great examples, composable.

---

### charmbracelet/bubbles (Components)
- **URL**: https://github.com/charmbracelet/bubbles
- **Relevance**: Pre-built TUI components for bubbletea

**Useful Components**:
- `spinner` - Loading/connection indicators
- `list` - Event log display
- `textinput` - For test trigger input
- `viewport` - Scrollable content
- `table` - For structured data

---

### Bubbletea Realtime Example ⭐ KEY REFERENCE
- **URL**: https://github.com/charmbracelet/bubbletea/blob/main/examples/realtime/main.go
- **Relevance**: CRITICAL - Shows exactly how to handle external events

**Pattern for WebSocket Integration**:
```go
// Create a channel for incoming events
sub := make(chan Event)

// Command that waits for channel messages
func waitForActivity(sub chan Event) tea.Cmd {
    return func() tea.Msg {
        return <-sub  // Blocks until event received
    }
}

// In Update, re-subscribe after receiving
case Event:
    m.events = append(m.events, msg)
    return m, waitForActivity(sub)
```

---

## 4. Reference Projects (Learn From These)

### anthonyeden/ProPresenter-Resolume ⭐ HIGHLY RELEVANT
- **URL**: https://github.com/anthonyeden/ProPresenter-Resolume
- **Language**: Python
- **Relevance**: Shows EXACTLY how to send text to Resolume's text generator

**What to Learn**:
- OSC address format for Resolume text clips
- How to trigger clips while setting text
- Timing between text set and clip trigger

**Key Insight**: This project sends lyrics to Resolume's text generator - same pattern we need for event text.

---

### xSpylon/resolume-osc-server
- **URL**: https://github.com/xSpylon/resolume-osc-server
- **Language**: Python
- **Relevance**: Simple Resolume OSC example

**What to Learn**:
- Basic Resolume OSC communication pattern
- Testing OSC connectivity

---

### Askath/Go-ChatAndClient ⭐ PATTERN REFERENCE
- **URL**: https://github.com/Askath/Go-ChatAndClient
- **Language**: Go
- **Relevance**: Shows bubbletea + gorilla/websocket integration

**What to Learn**:
- How to combine bubbletea TUI with WebSocket client
- Message routing pattern
- Connection state management in TUI

---

### AcChosen/EZTwitchOSCBot
- **URL**: https://github.com/AcChosen/EZTwitchOSCBot
- **Language**: Python
- **Relevance**: Twitch → OSC bridge (for VRChat, but pattern similar)

**What to Learn**:
- Event mapping structure
- Configuration patterns for Twitch→OSC

---

### AlexW-578/Spooder-Docker
- **URL**: https://github.com/AlexW-578/Spooder-Docker
- **Relevance**: Full Twitch integration with OSC

**What to Learn**:
- EventSub handling patterns
- OSC tunneling concepts

---

## 5. Resolume OSC Documentation

### Official Resolume OSC Support
- **URL**: https://resolume.com/support/en/osc
- **Relevance**: ESSENTIAL - Official documentation

**Key OSC Addresses for Text**:
```
# Trigger a clip
/composition/layers/[layer]/clips/[clip]/connect

# Set text on a text source (when clip is selected)
/composition/selectedclip/video/source/textgenerator/params/text/value

# Set text on specific clip's text source
/composition/layers/[layer]/clips/[clip]/video/source/textgenerator/params/text/value
```

---

## Next Steps

1. **Clone and study**: `anthonyeden/ProPresenter-Resolume` - closest to our use case
2. **Clone and study**: `Askath/Go-ChatAndClient` - bubbletea + websocket pattern
3. **Read**: Bubbletea realtime example for channel communication
4. **Spike**: Minimal Go program to send text to Resolume using hypebeast/go-osc

---

## Questions for Further Research

1. Does Resolume need the text clip to be "selected" to receive text updates?
   - Or can we target specific layer/clip directly?

2. What's the correct sequence: set text first, then trigger clip?
   - Or trigger clip, then set text while it plays?

3. How does Resolume handle rapid OSC messages?
   - Do we need debouncing/rate limiting?
