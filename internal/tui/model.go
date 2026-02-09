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
const maxLogs = 200

// Tab represents which view is active
type Tab int

const (
	TabEvents Tab = iota
	TabLogs
)

// Model is the bubbletea model for the TUI.
type Model struct {
	connectionStatus websocket.ConnectionState
	recentEvents     []eventEntry
	recentLogs       []logEntry
	viewport         viewport.Model
	ready            bool
	width            int
	height           int
	err              error
	activeTab        Tab

	// Channels for receiving updates
	eventsChan <-chan eventEntry
	stateChan  <-chan websocket.ConnectionState
	logsChan   <-chan logEntry

	// Test triggers keyed by number (1-8)
	testTriggers map[string]func()
}

type logEntry struct {
	timestamp time.Time
	message   string
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

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Underline(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

// NewModel creates a new TUI model.
// testTriggers maps key strings (e.g., "1", "2") to trigger functions.
func NewModel(eventsChan <-chan eventEntry, stateChan <-chan websocket.ConnectionState, logsChan <-chan logEntry, testTriggers map[string]func()) Model {
	return Model{
		connectionStatus: websocket.Disconnected,
		recentEvents:     make([]eventEntry, 0, maxEvents),
		recentLogs:       make([]logEntry, 0, maxLogs),
		eventsChan:       eventsChan,
		stateChan:        stateChan,
		logsChan:         logsChan,
		activeTab:        TabEvents,
		testTriggers:     testTriggers,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForEvent(m.eventsChan),
		waitForState(m.stateChan),
		waitForLog(m.logsChan),
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
		case "l":
			if m.activeTab == TabEvents {
				m.activeTab = TabLogs
			} else {
				m.activeTab = TabEvents
			}
			if m.ready {
				m.viewport.SetContent(m.renderActiveTab())
				m.viewport.GotoBottom()
			}
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
			m.viewport.SetContent(m.renderActiveTab())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = viewportHeight
		}

	case eventMsg:
		m.addEvent(eventEntry(msg))
		if m.ready && m.activeTab == TabEvents {
			m.viewport.SetContent(m.renderEvents())
			m.viewport.GotoBottom()
		}
		cmds = append(cmds, waitForEvent(m.eventsChan))

	case logMsg:
		m.addLog(logEntry(msg))
		if m.ready && m.activeTab == TabLogs {
			m.viewport.SetContent(m.renderLogs())
			m.viewport.GotoBottom()
		}
		cmds = append(cmds, waitForLog(m.logsChan))

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
	b.WriteString("  ")
	b.WriteString(m.renderTabs())
	b.WriteString("\n")
	b.WriteString(m.renderStatus())
	b.WriteString("\n\n")

	// Content viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Footer
	b.WriteString(helpStyle.Render("1-8:test events | l:toggle logs | q:quit"))

	return b.String()
}

func (m Model) renderTabs() string {
	eventsTab := "Events"
	logsTab := "Logs"

	if m.activeTab == TabEvents {
		eventsTab = tabActiveStyle.Render(eventsTab)
		logsTab = tabInactiveStyle.Render(logsTab)
	} else {
		eventsTab = tabInactiveStyle.Render(eventsTab)
		logsTab = tabActiveStyle.Render(logsTab)
	}

	return fmt.Sprintf("[%s] [%s]", eventsTab, logsTab)
}

func (m Model) renderActiveTab() string {
	if m.activeTab == TabLogs {
		return m.renderLogs()
	}
	return m.renderEvents()
}

func (m *Model) addEvent(e eventEntry) {
	m.recentEvents = append(m.recentEvents, e)
	if len(m.recentEvents) > maxEvents {
		m.recentEvents = m.recentEvents[1:]
	}
}

func (m *Model) addLog(l logEntry) {
	m.recentLogs = append(m.recentLogs, l)
	if len(m.recentLogs) > maxLogs {
		m.recentLogs = m.recentLogs[1:]
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

func (m Model) renderLogs() string {
	if len(m.recentLogs) == 0 {
		return helpStyle.Render("No logs yet...")
	}

	var b strings.Builder
	for _, l := range m.recentLogs {
		ts := timestampStyle.Render(l.timestamp.Format("15:04:05"))
		b.WriteString(fmt.Sprintf("%s %s\n", ts, l.message))
	}
	return b.String()
}

// Message types for tea.Cmd
type eventMsg eventEntry
type stateMsg websocket.ConnectionState
type logMsg logEntry

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

func waitForLog(ch <-chan logEntry) tea.Cmd {
	return func() tea.Msg {
		l, ok := <-ch
		if !ok {
			return nil
		}
		return logMsg(l)
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

// LogsChan type alias for external use.
type LogsChan = chan logEntry

// LogEntry creates a log entry for the TUI.
func LogEntry(message string) logEntry {
	return logEntry{
		timestamp: time.Now(),
		message:   message,
	}
}
