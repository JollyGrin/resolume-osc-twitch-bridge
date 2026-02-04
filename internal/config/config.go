// Package config handles loading and validation of YAML configuration.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure.
type Config struct {
	WebSocket WebSocketConfig          `yaml:"websocket"`
	OSC       OSCConfig                `yaml:"osc"`
	Mappings  map[string]EventMapping `yaml:"mappings"`
}

// WebSocketConfig holds WebSocket connection settings.
type WebSocketConfig struct {
	URL string `yaml:"url"`
}

// OSCConfig holds OSC connection settings.
type OSCConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// EventMapping defines how an event type maps to Resolume OSC.
type EventMapping struct {
	Layer    int    `yaml:"layer"`
	Clip     int    `yaml:"clip"`
	Template string `yaml:"template"`
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
