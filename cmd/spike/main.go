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
	"time"

	"github.com/hypebeast/go-osc/osc"
)

func main() {
	host := "127.0.0.1"
	port := 7000
	soloAddr := "/composition/groups/1/solo"

	fmt.Printf("Testing OSC solo toggle at %s:%d\n", host, port)

	client := osc.NewClient(host, port)

	// Theory: Resolume ignores repeated 1s, but 0 resets it

	// Step 1: Send 1 (should turn ON)
	fmt.Println("Step 1: Sending 1 (should turn solo ON)")
	msg1 := osc.NewMessage(soloAddr)
	msg1.Append(int32(1))
	client.Send(msg1)

	fmt.Println("Waiting 3 seconds... (solo should be ON)")
	time.Sleep(3 * time.Second)

	// Step 2: Send 0 (reset/release - might toggle OFF)
	fmt.Println("Step 2: Sending 0 (release/reset)")
	msg2 := osc.NewMessage(soloAddr)
	msg2.Append(int32(0))
	client.Send(msg2)

	fmt.Println("Waiting 1 second...")
	time.Sleep(1 * time.Second)

	// Step 3: Send 1 again (should this work now?)
	fmt.Println("Step 3: Sending 1 again (should turn solo ON again)")
	msg3 := osc.NewMessage(soloAddr)
	msg3.Append(int32(1))
	client.Send(msg3)

	fmt.Println("Done! Watch what happens at each step.")

	fmt.Println("Done! Solo should have toggled ON then OFF.")
}
