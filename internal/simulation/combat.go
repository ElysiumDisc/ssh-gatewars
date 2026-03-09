package simulation

// Combat represents an active battle. Stub for Layer 3.
type Combat struct {
	SystemID  int
	Attackers []CombatGroup
	Defenders []CombatGroup
	Round     int
	Resolved  bool
	Winner    int
}

// CombatGroup is a stack of identical ships.
type CombatGroup struct {
	Faction  int
	DesignID int
	Count    int
	HP       float64
	ShieldHP float64
}
