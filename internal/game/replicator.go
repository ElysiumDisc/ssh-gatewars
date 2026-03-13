package game

import "ssh-gatewars/internal/core"

// ReplicatorType classifies enemy variants.
type ReplicatorType int

const (
	ReplicatorBasic   ReplicatorType = iota // Small spider, 1 HP
	ReplicatorArmored                       // Tougher, 3 HP
	ReplicatorQueen                         // Large, spawns basics, 10 HP
)

// Replicator is an enemy entity approaching the chairs.
type Replicator struct {
	ID    int
	Type  ReplicatorType
	Pos   core.Vec2
	Vel   core.Vec2
	HP    int
	MaxHP int
	Alive bool
}

// ReplicatorDef holds static stats per type.
type ReplicatorDef struct {
	Type    ReplicatorType
	Name    string
	HP      int
	Speed   float64 // multiplier on base speed
	Damage  int     // damage to chair on contact
	ZPMDrop int     // ZPM awarded on kill
	Symbols [3]string // small, medium, large ASCII representations
}

// ReplicatorDefs defines stats for each type.
var ReplicatorDefs = map[ReplicatorType]ReplicatorDef{
	ReplicatorBasic: {
		Type: ReplicatorBasic, Name: "Bug",
		HP: 1, Speed: 1.0, Damage: 1, ZPMDrop: 5,
		Symbols: [3]string{"/╲●╲", " ╲║╱", ""},
	},
	ReplicatorArmored: {
		Type: ReplicatorArmored, Name: "Sentinel",
		HP: 3, Speed: 0.7, Damage: 2, ZPMDrop: 15,
		Symbols: [3]string{"/╲■╲", "═╧║╧═", " ╲║╱"},
	},
	ReplicatorQueen: {
		Type: ReplicatorQueen, Name: "Queen",
		HP: 10, Speed: 0.4, Damage: 5, ZPMDrop: 50,
		Symbols: [3]string{"╱╲◉╱╲", "║╬╬╬║", " ╲╩╩╱"},
	},
}
