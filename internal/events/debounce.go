package events

import (
	"log"
	"sync"
	"time"

	"github.com/waterhouse/resolume-twitch-osc/internal/osc"
)

// DebounceManager handles debounced group bypass and solo state.
type DebounceManager struct {
	osc          *osc.Client
	timer        *time.Timer
	groupSoloed  bool // Whether the group is currently solo'd
	groupBypassed bool // Whether the group is currently bypassed
	group        int  // Resolume group number
	soloEnabled  bool // Whether solo behavior is enabled
	mu           sync.Mutex
}

// NewDebounceManager creates a new debounce manager.
func NewDebounceManager(oscClient *osc.Client, group int, soloEnabled bool) *DebounceManager {
	return &DebounceManager{
		osc:         oscClient,
		group:       group,
		soloEnabled: soloEnabled,
	}
}

// SoloGroupIfNeeded solos and unbypasses the group if not already done.
// Call this before triggering clips for non-chat events.
func (dm *DebounceManager) SoloGroupIfNeeded() {
	if !dm.soloEnabled {
		return
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Unbypass the group if it was bypassed
	if dm.groupBypassed {
		if err := dm.osc.BypassGroup(dm.group, false); err != nil {
			log.Printf("failed to unbypass group: %v", err)
		} else {
			log.Printf("group %d unbypassed", dm.group)
			dm.groupBypassed = false
		}
	}

	// Solo the group if not already solo'd
	if !dm.groupSoloed {
		if err := dm.osc.SoloGroup(dm.group, true); err != nil {
			log.Printf("failed to solo group: %v", err)
		} else {
			log.Printf("solo: group %d solo'd", dm.group)
			dm.groupSoloed = true
		}
	}
}

// Schedule schedules a group bypass after the given duration.
// If there's already a timer, it will be canceled and replaced.
func (dm *DebounceManager) Schedule(layer, clip int, duration time.Duration) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Cancel existing timer if any
	if dm.timer != nil {
		dm.timer.Stop()
	}

	// Schedule new timer
	dm.timer = time.AfterFunc(duration, func() {
		dm.mu.Lock()
		defer dm.mu.Unlock()

		// Bypass the group
		if err := dm.osc.BypassGroup(dm.group, true); err != nil {
			log.Printf("debounce bypass group failed: %v", err)
		} else {
			log.Printf("debounce: bypassed group %d", dm.group)
			dm.groupBypassed = true
		}

		// Unsolo the group
		if dm.groupSoloed && dm.soloEnabled {
			if err := dm.osc.SoloGroup(dm.group, false); err != nil {
				log.Printf("failed to unsolo group: %v", err)
			} else {
				log.Printf("solo: group %d unsolo'd", dm.group)
				dm.groupSoloed = false
			}
		}

		dm.timer = nil
	})
}

// CancelAll cancels any pending timer, bypasses the group, and unsolos.
// Call on shutdown.
func (dm *DebounceManager) CancelAll() {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Cancel timer
	if dm.timer != nil {
		dm.timer.Stop()
		dm.timer = nil
	}

	// Bypass group
	if err := dm.osc.BypassGroup(dm.group, true); err != nil {
		log.Printf("shutdown bypass group failed: %v", err)
	}

	// Unsolo group if it was solo'd
	if dm.groupSoloed && dm.soloEnabled {
		if err := dm.osc.SoloGroup(dm.group, false); err != nil {
			log.Printf("shutdown unsolo failed: %v", err)
		}
		dm.groupSoloed = false
	}
}
