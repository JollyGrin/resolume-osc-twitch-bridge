package events

import (
	"log"
	"sync"
	"time"

	"github.com/waterhouse/resolume-twitch-osc/internal/osc"
)

// clipKey identifies a specific clip by layer and clip number.
type clipKey struct {
	layer int
	clip  int
}

// DebounceManager handles debounced clip disconnections.
type DebounceManager struct {
	osc    *osc.Client
	timers map[clipKey]*time.Timer
	mu     sync.Mutex
}

// NewDebounceManager creates a new debounce manager.
func NewDebounceManager(oscClient *osc.Client) *DebounceManager {
	return &DebounceManager{
		osc:    oscClient,
		timers: make(map[clipKey]*time.Timer),
	}
}

// Schedule schedules a clip disconnection after the given duration.
// If there's already a timer for this clip, it will be canceled and replaced.
func (dm *DebounceManager) Schedule(layer, clip int, duration time.Duration) {
	key := clipKey{layer: layer, clip: clip}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Cancel existing timer if any
	if existing, ok := dm.timers[key]; ok {
		existing.Stop()
	}

	// Schedule new timer
	dm.timers[key] = time.AfterFunc(duration, func() {
		if err := dm.osc.DisconnectClip(layer, clip); err != nil {
			log.Printf("debounce disconnect failed: %v", err)
		} else {
			log.Printf("debounce: disconnected layer %d clip %d", layer, clip)
		}

		dm.mu.Lock()
		delete(dm.timers, key)
		dm.mu.Unlock()
	})
}

// CancelAll cancels all pending timers. Call on shutdown.
func (dm *DebounceManager) CancelAll() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	for key, timer := range dm.timers {
		timer.Stop()
		delete(dm.timers, key)
	}
}
