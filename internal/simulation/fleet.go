package simulation

const (
	FleetIdle   = 0
	FleetMoving = 1
	FleetCombat = 2
)

const (
	TravelGate       = 0
	TravelHyperspace = 1
)

// Fleet represents a group of ships. Stub for Layer 2.
type Fleet struct {
	ID         uint64
	Faction    int
	SystemID   int
	Ships      map[int]int
	State      int
	FromSystem int
	ToSystem   int
	Progress   float64
	TravelMode int
	Speed      float64
}
