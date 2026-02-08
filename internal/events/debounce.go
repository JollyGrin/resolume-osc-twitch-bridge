package events

import (
	"log"
	"sync"
	"time"

	"github.com/waterhouse/resolume-twitch-osc/internal/osc"
)

// DebounceManager handles debounced group bypass and solo state.
type DebounceManager struct {
	osc           *osc.Client
	timer         *time.Timer
	generation    uint64 // Incremented on each Schedule to invalidate stale callbacks
	groupSoloed   bool   // Whether the group is currently solo'd
	groupBypassed bool   // Whether the group is currently bypassed
	group         int    // Resolume group number
	soloEnabled   bool   // Whether solo behavior is enabled
	mu            sync.Mutex
}

// NewDebounceManager creates a new debounce manager.
func NewDebounceManager(oscClient *osc.Client, group int, soloEnabled bool) *DebounceManager {
	return &DebounceManager{
		osc:         oscClient,
		group:       group,
		soloEnabled: soloEnabled,
	}
}

// SoloGroupIfNeeded solos and unbypasses the group.
// Call this before triggering clips for non-chat events.
func (dm *DebounceManager) SoloGroupIfNeeded() {
	if !dm.soloEnabled {
		return
	}

	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Invalidate any pending timer callbacks (race condition protection)
	dm.generation++

	// Set desired state: solo ON, bypass OFF
	// Send twice for UDP reliability (idempotent, so safe)
	dm.setGroupActive()
	time.Sleep(20 * time.Millisecond)
	dm.setGroupActive() // Redundant send for reliability

	dm.groupSoloed = true
	dm.groupBypassed = false
}

// setGroupActive sends bypass=OFF then solo=ON with delay between.
// Must be called with mutex held.
func (dm *DebounceManager) setGroupActive() {
	// Unbypass first, then solo (with delay so Resolume processes each)
	dm.osc.BypassGroup(dm.group, false)
	time.Sleep(10 * time.Millisecond)
	dm.osc.SoloGroup(dm.group, true)
}

// setGroupInactive sends solo=OFF then bypass=ON with delay between.
// Must be called with mutex held.
func (dm *DebounceManager) setGroupInactive() {
	// Unsolo first, then bypass
	dm.osc.SoloGroup(dm.group, false)
	time.Sleep(10 * time.Millisecond)
	dm.osc.BypassGroup(dm.group, true)
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

	// Increment generation - this is the authoritative generation for this timer
	dm.generation++
	gen := dm.generation
	log.Printf("solo: Schedule called, new gen=%d, duration=%v", gen, duration)

	// Schedule new timer
	dm.timer = time.AfterFunc(duration, func() {
		dm.mu.Lock()
		defer dm.mu.Unlock()

		// Check if this callback is stale (a newer Schedule superseded us)
		if gen != dm.generation {
			log.Printf("debounce: stale callback ignored (gen %d != current %d)", gen, dm.generation)
			return
		}

		// Set desired state: solo OFF, bypass ON
		// Send twice for UDP reliability (idempotent, so safe)
		if dm.soloEnabled {
			dm.setGroupInactive()
			time.Sleep(20 * time.Millisecond)
			dm.setGroupInactive() // Redundant send for reliability
			log.Printf("debounce: group %d set inactive (solo=OFF, bypass=ON)", dm.group)
		} else {
			// Just bypass if solo is disabled
			dm.osc.BypassGroup(dm.group, true)
			time.Sleep(20 * time.Millisecond)
			dm.osc.BypassGroup(dm.group, true)
		}

		dm.groupSoloed = false
		dm.groupBypassed = true
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

	// Set inactive state with redundancy for reliability
	if dm.soloEnabled {
		dm.setGroupInactive()
		time.Sleep(20 * time.Millisecond)
		dm.setGroupInactive()
	} else {
		dm.osc.BypassGroup(dm.group, true)
		time.Sleep(20 * time.Millisecond)
		dm.osc.BypassGroup(dm.group, true)
	}

	dm.groupSoloed = false
	dm.groupBypassed = true
}
