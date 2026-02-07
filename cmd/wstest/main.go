package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "wss://twitch-api.waterhousestudios.nl/ws"
	fmt.Println("Connecting to", url)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected! Waiting 15s for messages...")

	done := make(chan struct{})

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}
			fmt.Printf("MSG: %s\n", string(msg))
		}
	}()

	// Exit after 15s or on interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	select {
	case <-sig:
		fmt.Println("Interrupted")
	case <-time.After(15 * time.Second):
		fmt.Println("Timeout")
	case <-done:
	}
}
