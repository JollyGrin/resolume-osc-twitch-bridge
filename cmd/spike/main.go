// OSC Spike - Test connectivity to Resolume
//
// Prerequisites:
// - Resolume running with OSC enabled on port 7000
// - Text clip on Layer 1, Clip 1
//
// Run: go run ./cmd/spike

package main

import (
	"fmt"
	"log"

	"github.com/hypebeast/go-osc/osc"
)

func main() {
	host := "127.0.0.1"
	port := 7000

	fmt.Printf("Connecting to Resolume OSC at %s:%d\n", host, port)

	client := osc.NewClient(host, port)

	// Set text on Layer 1, Clip 1
	textAddr := "/composition/layers/1/clips/1/video/effects/textblock/effect/text/params/lines"
	textMsg := osc.NewMessage(textAddr)
	textMsg.Append("Test from Go!")

	fmt.Printf("Sending text: %s\n", textAddr)
	if err := client.Send(textMsg); err != nil {
		log.Fatalf("Failed to send text: %v", err)
	}

	// Trigger the clip
	triggerAddr := "/composition/layers/1/clips/1/connect"
	triggerMsg := osc.NewMessage(triggerAddr)
	triggerMsg.Append(int32(1))

	fmt.Printf("Triggering clip: %s\n", triggerAddr)
	if err := client.Send(triggerMsg); err != nil {
		log.Fatalf("Failed to trigger clip: %v", err)
	}

	fmt.Println("Done! Check Resolume for the text overlay.")
}
