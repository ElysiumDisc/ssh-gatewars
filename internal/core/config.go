package core

// GameConfig holds all tunable parameters for the server and game.
type GameConfig struct {
	// Network
	Port        int
	Host        string
	KeyPath     string
	MaxSessions int
	MaxPerKey   int
	ConnectRate float64

	// Persistence
	DBPath string

	// Simulation
	TickRate  int // ticks per second (default 10)
	RenderFPS int // TUI render rate (default 15)

	// Galaxy
	Seed       int64 // 0 = random
	NumPlanets int   // total planets in the galaxy

	// Defense
	HoldTimeSec    int     // base hold time per player in seconds
	DroneSpeed     float64 // drone movement per tick
	ReplicatorSpeed float64 // replicator movement per tick
	SpawnRadius    float64 // distance from center where replicators spawn
	DefenseRadius  float64 // defense perimeter radius
	DroneFireRate  int     // ticks between drone shots
	StartDrones    int     // starting drone count
	StartZPM       int     // starting ZPM balance

	// Events
	SurgeIntervalSec int // seconds between surge rotations

	// Chat
	ChatMaxMessageLen int
	ChatBacklogOps    int
	ChatBacklogLocal  int
	ChatBacklogTeam   int
	ChatMaxTeamSize   int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() GameConfig {
	return GameConfig{
		Port:        2222,
		Host:        "0.0.0.0",
		KeyPath:     "gatewars_host_key",
		MaxSessions: 200,
		MaxPerKey:   3,
		ConnectRate: 10.0,

		DBPath: "gatewars.db",

		TickRate:  10,
		RenderFPS: 15,

		Seed:       0,
		NumPlanets: 50,

		HoldTimeSec:     300, // 5 minutes
		DroneSpeed:       0.8,
		ReplicatorSpeed:  0.15,
		SpawnRadius:      20.0,
		DefenseRadius:    10.0,
		DroneFireRate:    10, // every 1.0 sec at 10Hz (chair level reduces further)
		StartDrones:      5,
		StartZPM:         0,

		SurgeIntervalSec: 300, // surge rotates every 5 min

		ChatMaxMessageLen: 500,
		ChatBacklogOps:    100,
		ChatBacklogLocal:  50,
		ChatBacklogTeam:   100,
		ChatMaxTeamSize:   4,
	}
}
