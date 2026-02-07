// Resolume Twitch OSC Bridge
//
// Receives Twitch events via WebSocket and sends OSC to Resolume.
//
// Usage: go run ./cmd/bridge

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/waterhouse/resolume-twitch-osc/internal/config"
	"github.com/waterhouse/resolume-twitch-osc/internal/events"
	"github.com/waterhouse/resolume-twitch-osc/internal/osc"
	"github.com/waterhouse/resolume-twitch-osc/internal/tui"
	"github.com/waterhouse/resolume-twitch-osc/internal/websocket"
)

func main() {
	// Load config
	cfg, err := config.Load("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize OSC client
	oscClient := osc.NewClient(cfg.OSC.Host, cfg.OSC.Port)

	// Initialize event handler
	eventHandler := events.NewHandler(oscClient, cfg.Mappings, cfg.Defaults)

	// Initialize WebSocket client
	wsClient := websocket.NewClient(cfg.WebSocket.URL)

	// Channels for TUI
	tuiEvents := make(tui.EventsChan, 100)
	tuiState := make(chan websocket.ConnectionState, 10)

	// Test trigger function
	testTrigger := func() {
		text := "TestUser just followed!"
		if err := oscClient.TriggerClipWithText(1, 1, text); err != nil {
			log.Printf("Test trigger failed: %v", err)
		}
		tuiEvents <- tui.EventEntry("test", text)
	}

	// Start WebSocket connection
	wsClient.Connect()

	// Forward WebSocket state to TUI
	go func() {
		for state := range wsClient.State() {
			select {
			case tuiState <- state:
			default:
			}
		}
	}()

	// Process WebSocket messages
	go func() {
		for msg := range wsClient.Messages() {
			event, err := events.Parse(msg)
			if err != nil {
				log.Printf("Failed to parse event: %v", err)
				continue
			}

			text, err := eventHandler.Handle(event)
			if err != nil {
				log.Printf("Failed to handle event: %v", err)
				continue
			}

			if text != "" {
				tuiEvents <- tui.EventEntry(event.Type, text)
			}
		}
	}()

	// Create TUI
	model := tui.NewModel(tuiEvents, tuiState, testTrigger)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		wsClient.Close()
		p.Quit()
	}()

	// Run TUI
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}

	// Cleanup
	eventHandler.Close()
	wsClient.Close()
	time.Sleep(100 * time.Millisecond) // Allow cleanup
}
