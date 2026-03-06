package simulation

// Spawner configuration constants.
const (
	BaseSpawnInterval = 3.0  // seconds between spawns per faction (base)
	MinSpawnInterval  = 0.2  // minimum spawn interval (cap)
	PlayerShipCap     = 16   // max ships-per-cycle boost from players
	PlayerHPBonus     = 0.02 // HP bonus per player beyond cap
)
