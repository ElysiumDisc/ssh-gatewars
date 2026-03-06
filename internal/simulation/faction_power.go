package simulation

import (
	"fmt"
	"sync"
	"time"

	"ssh-gatewars/internal/faction"
)

// PowerState represents the lifecycle of a faction power.
type PowerState int

const (
	PowerReady    PowerState = iota
	PowerActive              // effect is running
	PowerCooldown            // waiting to recharge
)

// FactionPower defines a faction's active ability.
type FactionPower struct {
	Name     string
	Duration time.Duration
	Cooldown time.Duration

	State       PowerState
	ActiveUntil time.Time
	ReadyAt     time.Time
	ActivatedBy string // SSH key of the player who triggered
}

// PowerStatus is a snapshot of power state for rendering.
type PowerStatus struct {
	Name        string
	State       PowerState
	Remaining   time.Duration // time left in current state
	Total       time.Duration // total duration of current state
	ActivatedBy string
}

// PowerManager manages all 5 faction powers.
type PowerManager struct {
	mu     sync.Mutex
	powers [faction.Count]*FactionPower
}

// NewPowerManager creates powers for all factions.
func NewPowerManager() *PowerManager {
	pm := &PowerManager{}
	pm.powers[faction.Tauri] = &FactionPower{
		Name: "COORDINATED STRIKE", Duration: 5 * time.Second, Cooldown: 45 * time.Second,
	}
	pm.powers[faction.Goauld] = &FactionPower{
		Name: "BOMBARDMENT", Duration: 5 * time.Second, Cooldown: 45 * time.Second,
	}
	pm.powers[faction.Jaffa] = &FactionPower{
		Name: "KREE!", Duration: 8 * time.Second, Cooldown: 45 * time.Second,
	}
	pm.powers[faction.Lucian] = &FactionPower{
		Name: "KASSA RUSH", Duration: 10 * time.Second, Cooldown: 45 * time.Second,
	}
	pm.powers[faction.Asgard] = &FactionPower{
		Name: "ION CANNON", Duration: 1 * time.Second, Cooldown: 30 * time.Second,
	}
	return pm
}

// TryActivate attempts to fire a faction power. Returns success and a notification message.
func (pm *PowerManager) TryActivate(factionID int, playerKey string) (bool, string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.powers[factionID]
	pm.updateState(p)

	if p.State != PowerReady {
		return false, ""
	}

	now := time.Now()
	p.State = PowerActive
	p.ActiveUntil = now.Add(p.Duration)
	p.ReadyAt = now.Add(p.Duration + p.Cooldown)
	p.ActivatedBy = playerKey

	f := faction.Factions[factionID]
	msg := fmt.Sprintf("A %s soldier has activated %s!", f.ShortName, p.Name)
	return true, msg
}

// IsActive returns whether a faction's power is currently active.
func (pm *PowerManager) IsActive(factionID int) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.powers[factionID]
	pm.updateState(p)
	return p.State == PowerActive
}

// Status returns a snapshot of a faction's power state for rendering.
func (pm *PowerManager) Status(factionID int) PowerStatus {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p := pm.powers[factionID]
	pm.updateState(p)

	now := time.Now()
	status := PowerStatus{
		Name:        p.Name,
		State:       p.State,
		ActivatedBy: p.ActivatedBy,
	}

	switch p.State {
	case PowerReady:
		// no timing info needed
	case PowerActive:
		status.Remaining = p.ActiveUntil.Sub(now)
		status.Total = p.Duration
	case PowerCooldown:
		status.Remaining = p.ReadyAt.Sub(now)
		status.Total = p.Cooldown
	}

	return status
}

// updateState transitions power state based on time. Must be called with lock held.
func (pm *PowerManager) updateState(p *FactionPower) {
	now := time.Now()
	switch p.State {
	case PowerActive:
		if now.After(p.ActiveUntil) {
			p.State = PowerCooldown
		}
	case PowerCooldown:
		if now.After(p.ReadyAt) {
			p.State = PowerReady
			p.ActivatedBy = ""
		}
	}
}
