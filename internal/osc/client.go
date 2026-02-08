// Package osc provides a wrapper around hypebeast/go-osc for Resolume communication.
package osc

import (
	"fmt"
	"log"

	"github.com/hypebeast/go-osc/osc"
)

// Client wraps the go-osc client for sending messages to Resolume.
type Client struct {
	client *osc.Client
}

// NewClient creates a new OSC client connected to the specified host and port.
func NewClient(host string, port int) *Client {
	return &Client{
		client: osc.NewClient(host, port),
	}
}

// SendText sends a text value to the specified OSC address.
func (c *Client) SendText(address, text string) error {
	msg := osc.NewMessage(address)
	msg.Append(text)

	if err := c.client.Send(msg); err != nil {
		log.Printf("OSC send text failed: %v", err)
		return fmt.Errorf("sending text to %s: %w", address, err)
	}
	return nil
}

// SendTrigger sends an integer trigger (value 1) to the specified OSC address.
func (c *Client) SendTrigger(address string) error {
	msg := osc.NewMessage(address)
	msg.Append(int32(1))

	if err := c.client.Send(msg); err != nil {
		log.Printf("OSC send trigger failed: %v", err)
		return fmt.Errorf("sending trigger to %s: %w", address, err)
	}
	return nil
}

// TriggerClipWithText sets text on a clip and triggers it.
// This is the main method for displaying text overlays in Resolume.
func (c *Client) TriggerClipWithText(layer, clip int, text string) error {
	// Set the text
	textAddr := fmt.Sprintf("/composition/layers/%d/clips/%d/video/effects/textblock/effect/text/params/lines", layer, clip)
	if err := c.SendText(textAddr, text); err != nil {
		return err
	}

	// Trigger the clip
	triggerAddr := fmt.Sprintf("/composition/layers/%d/clips/%d/connect", layer, clip)
	return c.SendTrigger(triggerAddr)
}

// TriggerClip triggers a clip without setting text.
func (c *Client) TriggerClip(layer, clip int) error {
	triggerAddr := fmt.Sprintf("/composition/layers/%d/clips/%d/connect", layer, clip)
	return c.SendTrigger(triggerAddr)
}

// DisconnectClip disconnects a clip by sending value 0 to its connect address.
func (c *Client) DisconnectClip(layer, clip int) error {
	addr := fmt.Sprintf("/composition/layers/%d/clips/%d/connect", layer, clip)
	msg := osc.NewMessage(addr)
	msg.Append(int32(0))

	if err := c.client.Send(msg); err != nil {
		log.Printf("OSC disconnect clip failed: %v", err)
		return fmt.Errorf("disconnecting clip at %s: %w", addr, err)
	}
	return nil
}

// SoloGroup toggles solo for a Resolume group.
func (c *Client) SoloGroup(group int, on bool) error {
	addr := fmt.Sprintf("/composition/groups/%d/solo", group)
	msg := osc.NewMessage(addr)
	val := int32(0)
	if on {
		val = 1
	}
	msg.Append(val)

	if err := c.client.Send(msg); err != nil {
		log.Printf("OSC solo group failed: %v", err)
		return fmt.Errorf("setting solo on group %d: %w", group, err)
	}
	return nil
}

// BypassGroup toggles bypass for a Resolume group.
func (c *Client) BypassGroup(group int, bypass bool) error {
	addr := fmt.Sprintf("/composition/groups/%d/bypassed", group)
	msg := osc.NewMessage(addr)
	val := int32(0)
	if bypass {
		val = 1
	}
	msg.Append(val)

	if err := c.client.Send(msg); err != nil {
		log.Printf("OSC bypass group failed: %v", err)
		return fmt.Errorf("setting bypass on group %d: %w", group, err)
	}
	return nil
}
