package game

import "ssh-gatewars/internal/core"

// DroneTier defines upgrade levels for drones.
type DroneTier int

const (
	DroneTierBase    DroneTier = iota // Yellow — standard
	DroneTierCyan                     // Cyan — faster fire rate
	DroneTierMagenta                  // Magenta — splash damage
	DroneTierWhite                    // White — piercing
)

// Drone is an active projectile fired from a chair.
type Drone struct {
	ID       int
	OwnerFP  string    // chair owner's fingerprint
	Pos      core.Vec2
	Vel      core.Vec2
	TargetID int       // replicator ID being tracked
	Tier     DroneTier
	Damage   int
	Alive    bool
}

// DroneConfig holds static stats per tier.
type DroneConfig struct {
	Tier      DroneTier
	Name      string
	Damage    int
	Speed     float64 // multiplier on base speed
	Splash    float64 // splash radius (0 = single target)
	Pierce    bool    // passes through targets
	Color     string  // ANSI color name
	Symbol    rune
}

// DroneTiers defines stats for each tier.
var DroneTiers = map[DroneTier]DroneConfig{
	DroneTierBase: {
		Tier: DroneTierBase, Name: "Standard Drone",
		Damage: 1, Speed: 1.0, Color: "yellow", Symbol: '✸',
	},
	DroneTierCyan: {
		Tier: DroneTierCyan, Name: "Swift Drone",
		Damage: 1, Speed: 1.5, Color: "cyan", Symbol: '✸',
	},
	DroneTierMagenta: {
		Tier: DroneTierMagenta, Name: "Blast Drone",
		Damage: 2, Speed: 1.0, Splash: 2.0, Color: "magenta", Symbol: '✸',
	},
	DroneTierWhite: {
		Tier: DroneTierWhite, Name: "Piercing Drone",
		Damage: 3, Speed: 1.2, Pierce: true, Color: "white", Symbol: '✸',
	},
}
