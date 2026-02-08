// Package config handles loading and validation of YAML configuration.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure.
type Config struct {
	WebSocket WebSocketConfig          `yaml:"websocket"`
	OSC       OSCConfig                `yaml:"osc"`
	Defaults  Defaults                 `yaml:"defaults"`
	Mappings  map[string]EventMapping `yaml:"mappings"`
}

// Defaults holds default values for event mappings.
type Defaults struct {
	Debounce    Duration `yaml:"debounce"`
	Group       int      `yaml:"group"`         // Resolume group number for event layers
	SoloOnEvent bool     `yaml:"solo_on_event"` // Enable group solo behavior
}

// Duration wraps time.Duration for YAML unmarshaling.
type Duration time.Duration

// UnmarshalYAML parses duration strings like "8s", "5m", etc.
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// Duration returns the time.Duration value.
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// WebSocketConfig holds WebSocket connection settings.
type WebSocketConfig struct {
	URL string `yaml:"url"`
}

// OSCConfig holds OSC connection settings.
type OSCConfig struct {
	Host  string   `yaml:"host"`
	Port  int      `yaml:"port"`
	Delay Duration `yaml:"delay,omitempty"` // Delay between OSC messages (default 20ms)
}

// EventMapping defines how an event type maps to Resolume OSC.
type EventMapping struct {
	Actions       []Action  `yaml:"actions"`
	Debounce      *Duration `yaml:"debounce,omitempty"`       // Override default debounce
	ReturnToScene *bool     `yaml:"return_to_scene,omitempty"` // Default true, set false to disable
}

// Action defines a single OSC action (trigger a clip, optionally with text).
type Action struct {
	Layer    int    `yaml:"layer"`
	Clip     int    `yaml:"clip"`
	Template string `yaml:"template,omitempty"` // Optional: if empty, just triggers clip
}

// GetDebounce returns the effective debounce duration for this mapping.
func (m *EventMapping) GetDebounce(defaultDebounce Duration) time.Duration {
	if m.Debounce != nil {
		return m.Debounce.Duration()
	}
	return defaultDebounce.Duration()
}

// ShouldReturnToScene returns whether this event should trigger a return after debounce.
func (m *EventMapping) ShouldReturnToScene() bool {
	if m.ReturnToScene != nil {
		return *m.ReturnToScene
	}
	return true // default to true
}

// Load reads configuration from the specified file path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks that required fields are set.
func (c *Config) validate() error {
	if c.WebSocket.URL == "" {
		return fmt.Errorf("websocket.url is required")
	}
	if c.OSC.Host == "" {
		return fmt.Errorf("osc.host is required")
	}
	if c.OSC.Port == 0 {
		return fmt.Errorf("osc.port is required")
	}
	if len(c.Mappings) == 0 {
		return fmt.Errorf("at least one event mapping is required")
	}
	return nil
}
