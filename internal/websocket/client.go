// Package websocket provides a WebSocket client with automatic reconnection.
package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionState represents the current connection status.
type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
)

func (s ConnectionState) String() string {
	switch s {
	case Disconnected:
		return "disconnected"
	case Connecting:
		return "connecting"
	case Connected:
		return "connected"
	default:
		return "unknown"
	}
}

// Client is a WebSocket client with automatic reconnection.
type Client struct {
	url      string
	conn     *websocket.Conn
	mu       sync.Mutex
	done     chan struct{}
	messages chan []byte
	state    chan ConnectionState
}

// NewClient creates a new WebSocket client for the given URL.
func NewClient(url string) *Client {
	return &Client{
		url:      url,
		done:     make(chan struct{}),
		messages: make(chan []byte, 100),
		state:    make(chan ConnectionState, 10),
	}
}

// Messages returns the channel for receiving messages.
func (c *Client) Messages() <-chan []byte {
	return c.messages
}

// State returns the channel for connection state changes.
func (c *Client) State() <-chan ConnectionState {
	return c.state
}

// Connect starts the connection loop with automatic reconnection.
func (c *Client) Connect() {
	go c.connectLoop()
}

// Close stops the client and closes the connection.
func (c *Client) Close() {
	close(c.done)
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
	}
	c.mu.Unlock()
}

func (c *Client) connectLoop() {
	delay := time.Second

	for {
		select {
		case <-c.done:
			return
		default:
		}

		c.setState(Connecting)
		log.Printf("WebSocket connecting to %s", c.url)

		conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
		if err != nil {
			log.Printf("WebSocket connection failed: %v", err)
			c.setState(Disconnected)
			c.backoff(&delay)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()
		delay = time.Second // Reset backoff on successful connect

		c.setState(Connected)
		log.Printf("WebSocket connected")

		// Set up ping/pong handling
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		// Read messages until error
		c.readLoop(conn)

		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		c.setState(Disconnected)
	}
}

func (c *Client) readLoop(conn *websocket.Conn) {
	defer conn.Close()

	// Set initial read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		select {
		case <-c.done:
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}

		// Reset deadline on any message (server sends pings)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		select {
		case c.messages <- message:
		default:
			log.Printf("WebSocket message buffer full, dropping message")
		}
	}
}

func (c *Client) setState(state ConnectionState) {
	select {
	case c.state <- state:
	default:
	}
}

func (c *Client) backoff(delay *time.Duration) {
	select {
	case <-c.done:
		return
	case <-time.After(*delay):
	}

	*delay *= 2
	if *delay > 30*time.Second {
		*delay = 30 * time.Second
	}
}
