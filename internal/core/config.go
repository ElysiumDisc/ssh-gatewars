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
	TickRate int // ticks per second (default 10)
	RenderFPS int // TUI render rate (default 15)

	// World generation
	Seed     int64 // 0 = random
	MapWidth  int  // planet map width in tiles
	MapHeight int  // planet map height in tiles
	SGCWidth  int
	SGCHeight int

	// Combat
	StartHP       int
	StartMaxHP    int
	BaseXPPerKill int
	XPPerLevel    int // XP needed = level * XPPerLevel
	DeathPenalty  float64 // fraction of XP lost on death (0.0-1.0)

	// Enemies
	EnemyDensity   float64 // enemies per 100 tiles
	EnemyTickRate  int     // enemy AI ticks per engine tick (1 = every tick, 3 = every 3rd)
	AggroRange     int     // tiles
	ChaseRange     int     // tiles (give up chase beyond this)

	// Items
	LootChance     float64 // probability of enemy dropping loot (0.0-1.0)
	CrateChance    float64 // probability of crate per room

	// Planet lifecycle
	UnloadTimeout int // seconds before unloading empty planet

	// Chat
	ChatMaxMessageLen int
	ChatBacklogOps    int
	ChatBacklogLocal  int
	ChatBacklogTeam   int
	ChatCompactRows   int
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

		Seed:      0,
		MapWidth:  80,
		MapHeight: 50,
		SGCWidth:  40,
		SGCHeight: 20,

		StartHP:       20,
		StartMaxHP:    20,
		BaseXPPerKill: 10,
		XPPerLevel:    100,
		DeathPenalty:  0.1,

		EnemyDensity:  1.5,
		EnemyTickRate: 3,
		AggroRange:    8,
		ChaseRange:    15,

		LootChance:  0.4,
		CrateChance: 0.3,

		UnloadTimeout: 60,

		ChatMaxMessageLen: 500,
		ChatBacklogOps:    100,
		ChatBacklogLocal:  50,
		ChatBacklogTeam:   100,
		ChatCompactRows:   6,
		ChatMaxTeamSize:   4,
	}
}
