package game

import "ssh-gatewars/internal/core"

// PlanetStatus tracks the state of a planet in the galaxy.
type PlanetStatus int

const (
	PlanetInvaded   PlanetStatus = iota // Replicators control it
	PlanetContested                     // Players are defending
	PlanetFree                          // Liberated
)

// Planet represents a world in the shared galaxy.
type Planet struct {
	ID             int
	Name           string
	Seed           int64
	Pos            core.Vec2 // position in galaxy space
	Status         PlanetStatus
	InvasionLevel  int  // difficulty 1-10
	DefenderCount  int  // current number of players
	Surging        bool // active replicator surge (2x spawns, 2x ZPM)
	BountyZPM      int  // bonus ZPM awarded on liberation
}
