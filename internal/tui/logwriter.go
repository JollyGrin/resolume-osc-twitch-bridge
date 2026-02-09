package tui

import (
	"strings"
	"time"
)

// LogWriter implements io.Writer and sends log messages to a channel.
type LogWriter struct {
	ch chan<- logEntry
}

// NewLogWriter creates a new LogWriter that sends to the given channel.
func NewLogWriter(ch chan<- logEntry) *LogWriter {
	return &LogWriter{ch: ch}
}

// Write implements io.Writer. It parses log lines and sends them to the channel.
func (w *LogWriter) Write(p []byte) (n int, err error) {
	message := strings.TrimSpace(string(p))
	if message == "" {
		return len(p), nil
	}

	// Strip the timestamp prefix that log.Printf adds (e.g., "2006/01/02 15:04:05 ")
	// We'll add our own timestamp
	if len(message) > 20 && message[4] == '/' && message[7] == '/' {
		// Looks like it starts with a date, skip past "YYYY/MM/DD HH:MM:SS "
		if idx := strings.Index(message, " "); idx > 0 {
			if idx2 := strings.Index(message[idx+1:], " "); idx2 > 0 {
				message = message[idx+1+idx2+1:]
			}
		}
	}

	select {
	case w.ch <- logEntry{timestamp: time.Now(), message: message}:
	default:
		// Channel full, drop message
	}

	return len(p), nil
}
