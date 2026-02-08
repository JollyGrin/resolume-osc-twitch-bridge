// Package tui provides a terminal user interface using bubbletea.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/waterhouse/resolume-twitch-osc/internal/websocket"
)

const maxEvents = 100

// Model is the bubbletea model for the TUI.
type Model struct {
	connectionStatus websocket.ConnectionState
	recentEvents     []eventEntry
	viewport         viewport.Model
	ready            bool
	width            int
	height           int
	err              error

	// Channels for receiving updates
	eventsChan <-chan eventEntry
	stateChan  <-chan websocket.ConnectionState

	// Test triggers keyed by number (1-8)
	testTriggers map[string]func()
}

type eventEntry struct {
	timestamp time.Time
	eventType string
	text      string
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	statusConnected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	statusDisconnected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	statusConnecting = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true)

	eventTypeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// NewModel creates a new TUI model.
// testTriggers maps key strings (e.g., "1", "2") to trigger functions.
func NewModel(eventsChan <-chan eventEntry, stateChan <-chan websocket.ConnectionState, testTriggers map[string]func()) Model {
	return Model{
		connectionStatus: websocket.Disconnected,
		recentEvents:     make([]eventEntry, 0, maxEvents),
		eventsChan:       eventsChan,
		stateChan:        stateChan,
		testTriggers:     testTriggers,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForEvent(m.eventsChan),
		waitForState(m.stateChan),
	)
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6", "7", "8":
			if fn, ok := m.testTriggers[key]; ok {
				fn()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 4
		footerHeight := 2
		viewportHeight := m.height - headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(m.width, viewportHeight)
			m.viewport.SetContent(m.renderEvents())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = viewportHeight
		}

	case eventMsg:
		m.addEvent(eventEntry(msg))
		if m.ready {
			m.viewport.SetContent(m.renderEvents())
			m.viewport.GotoBottom()
		}
		cmds = append(cmds, waitForEvent(m.eventsChan))

	case stateMsg:
		m.connectionStatus = websocket.ConnectionState(msg)
		cmds = append(cmds, waitForState(m.stateChan))
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("Resolume Twitch OSC Bridge"))
	b.WriteString("\n")
	b.WriteString(m.renderStatus())
	b.WriteString("\n\n")

	// Event log
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Footer
	b.WriteString(helpStyle.Render("1:follow 2:sub 3:gift 4:cheer 5:raid 6:chat 7:start 8:end | q:quit"))

	return b.String()
}

func (m *Model) addEvent(e eventEntry) {
	m.recentEvents = append(m.recentEvents, e)
	if len(m.recentEvents) > maxEvents {
		m.recentEvents = m.recentEvents[1:]
	}
}

func (m Model) renderStatus() string {
	var status string
	switch m.connectionStatus {
	case websocket.Connected:
		status = statusConnected.Render("● Connected")
	case websocket.Connecting:
		status = statusConnecting.Render("○ Connecting...")
	default:
		status = statusDisconnected.Render("○ Disconnected")
	}
	return fmt.Sprintf("Status: %s", status)
}

func (m Model) renderEvents() string {
	if len(m.recentEvents) == 0 {
		return helpStyle.Render("No events yet. Waiting for Twitch events...")
	}

	var b strings.Builder
	for _, e := range m.recentEvents {
		ts := timestampStyle.Render(e.timestamp.Format("15:04:05"))
		evType := eventTypeStyle.Render(fmt.Sprintf("[%s]", e.eventType))
		b.WriteString(fmt.Sprintf("%s %s %s\n", ts, evType, e.text))
	}
	return b.String()
}

// Message types for tea.Cmd
type eventMsg eventEntry
type stateMsg websocket.ConnectionState

func waitForEvent(ch <-chan eventEntry) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-ch
		if !ok {
			return nil
		}
		return eventMsg(e)
	}
}

func waitForState(ch <-chan websocket.ConnectionState) tea.Cmd {
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return nil
		}
		return stateMsg(s)
	}
}

// EventEntry creates an event entry for the TUI.
func EventEntry(eventType, text string) eventEntry {
	return eventEntry{
		timestamp: time.Now(),
		eventType: eventType,
		text:      text,
	}
}

// EventsChan type alias for external use.
type EventsChan = chan eventEntry
