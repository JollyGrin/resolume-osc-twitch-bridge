// Resolume Twitch OSC Bridge
//
// Receives Twitch events via WebSocket and sends OSC to Resolume.
//
// Usage: go run ./cmd/bridge

package main

import (
	"fmt"
	"log"
	"math/rand"
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
	oscClient := osc.NewClient(cfg.OSC.Host, cfg.OSC.Port, cfg.OSC.Delay.Duration())

	// Initialize event handler
	eventHandler := events.NewHandler(oscClient, cfg.Mappings, cfg.Defaults)

	// Initialize WebSocket client
	wsClient := websocket.NewClient(cfg.WebSocket.URL)

	// Channels for TUI
	tuiEvents := make(tui.EventsChan, 100)
	tuiState := make(chan websocket.ConnectionState, 10)
	tuiLogs := make(tui.LogsChan, 200)

	// Redirect log output to TUI
	log.SetOutput(tui.NewLogWriter(tuiLogs))
	log.SetFlags(0) // Remove default timestamp, we add our own

	// Test triggers map - keyed by number key string
	testTriggers := map[string]func(){
		// 1: follow
		"1": func() {
			userName := fmt.Sprintf("TestUser%02d", rand.Intn(100))
			text, err := eventHandler.HandleTestEvent("follow", events.FollowEvent{UserName: userName})
			if err != nil {
				log.Printf("Test follow failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("follow", text)
		},
		// 2: subscribe
		"2": func() {
			userName := fmt.Sprintf("TestUser%02d", rand.Intn(100))
			text, err := eventHandler.HandleTestEvent("subscribe", events.SubscribeEvent{UserName: userName, Tier: "1000"})
			if err != nil {
				log.Printf("Test subscribe failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("subscribe", text)
		},
		// 3: gift_sub
		"3": func() {
			userName := fmt.Sprintf("TestUser%02d", rand.Intn(100))
			total := rand.Intn(10) + 1
			text, err := eventHandler.HandleTestEvent("gift_sub", events.GiftSubEvent{UserName: userName, Total: total, Tier: "1000"})
			if err != nil {
				log.Printf("Test gift_sub failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("gift_sub", text)
		},
		// 4: cheer (random bits)
		"4": func() {
			userName := fmt.Sprintf("TestUser%02d", rand.Intn(100))
			bits := (rand.Intn(50) + 1) * 100 // 100-5000 bits
			text, err := eventHandler.HandleTestEvent("cheer", events.CheerEvent{UserName: userName, Bits: bits, Message: "Woohoo!"})
			if err != nil {
				log.Printf("Test cheer failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("cheer", text)
		},
		// 5: raid (random viewers)
		"5": func() {
			viewers := rand.Intn(500) + 10
			text, err := eventHandler.HandleTestEvent("raid", events.RaidEvent{FromBroadcasterUserName: "TestStreamer", Viewers: viewers})
			if err != nil {
				log.Printf("Test raid failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("raid", text)
		},
		// 6: chat
		"6": func() {
			messages := []string{
				"Hello everyone!",
				"This stream is awesome!",
				"LUL that was hilarious",
				"GG well played",
			}
			userName := fmt.Sprintf("Chatter%02d", rand.Intn(100))
			message := messages[rand.Intn(len(messages))]
			text, err := eventHandler.HandleTestEvent("chat", events.ChatEvent{
				ChatterUserName: userName,
				Message:         events.ChatMessage{Text: message},
			})
			if err != nil {
				log.Printf("Test chat failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("chat", text)
		},
		// 7: stream_start
		"7": func() {
			text, err := eventHandler.HandleTestEvent("stream_start", events.StreamStartEvent{})
			if err != nil {
				log.Printf("Test stream_start failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("stream_start", text)
		},
		// 8: stream_end
		"8": func() {
			text, err := eventHandler.HandleTestEvent("stream_end", events.StreamEndEvent{})
			if err != nil {
				log.Printf("Test stream_end failed: %v", err)
			}
			tuiEvents <- tui.EventEntry("stream_end", text)
		},
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
	model := tui.NewModel(tuiEvents, tuiState, tuiLogs, testTriggers)
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
