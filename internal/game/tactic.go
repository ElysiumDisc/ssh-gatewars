package game

// DroneTactic controls how a chair's drones select targets.
type DroneTactic int

const (
	TacticSpread    DroneTactic = iota // Target nearest — default spread fire
	TacticFocus                        // All drones focus the strongest enemy
	TacticPerimeter                    // Prioritize enemies closest to center
)

// TacticNames maps tactic to display name.
var TacticNames = map[DroneTactic]string{
	TacticSpread:    "SPREAD",
	TacticFocus:     "FOCUS",
	TacticPerimeter: "PERIMETER",
}

// TacticDescs maps tactic to description.
var TacticDescs = map[DroneTactic]string{
	TacticSpread:    "Drones target nearest enemy",
	TacticFocus:     "All drones focus strongest threat",
	TacticPerimeter: "Defend the inner perimeter",
}
