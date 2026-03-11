package gamedata

// AIBehavior describes how an enemy acts.
type AIBehavior int

const (
	AIPatrol   AIBehavior = iota // wander randomly
	AIGuard                      // stay put, aggro on proximity
	AIHunt                       // actively seek players
	AISentry                     // ranged guard, fires from position
	AISwarm                      // boid-like swarm movement
	AISupport                    // heals/buffs nearby allies
	AIAmbush                     // cloaked until close range
)

// EnemyDef defines a type of enemy.
type EnemyDef struct {
	ID          string
	Name        string
	Glyph       rune
	Color       string
	HP          int
	Attack      int
	Defense     int
	AI          AIBehavior
	AggroRange  int    // tiles before they notice you
	Speed       int    // ticks between moves (lower = faster)
	XP          int    // XP awarded on kill
	LootTable   string
	Faction     string // "goauld", "replicator", "ori", "wildlife", "neutral"
	IsBoss      bool

	// Ranged enemy stats
	IsRanged    bool
	Range       int    // attack range in tiles (0 = melee only)
	Accuracy    int    // base hit chance (percentage)
	ProjGlyph   rune   // projectile character
	ProjColor   string // projectile lipgloss color
}

// Enemies is the catalog of all enemy definitions.
var Enemies = map[string]EnemyDef{
	// ─── Goa'uld Forces ─────────────────────────────────────
	"jaffa_warrior": {
		ID: "jaffa_warrior", Name: "Jaffa Warrior",
		Glyph: 'j', Color: "#DDCC00",
		HP: 30, Attack: 8, Defense: 2,
		AI: AIPatrol, AggroRange: 6, Speed: 3,
		XP: 10, LootTable: "jaffa_common", Faction: "goauld",
		IsRanged: true, Range: 8, Accuracy: 65,
		ProjGlyph: '*', ProjColor: "#FFAA00",
	},
	"serpent_guard": {
		ID: "serpent_guard", Name: "Serpent Guard",
		Glyph: 'J', Color: "#DD2222",
		HP: 50, Attack: 12, Defense: 4,
		AI: AIGuard, AggroRange: 7, Speed: 3,
		XP: 25, LootTable: "serpent_guard", Faction: "goauld",
		IsRanged: true, Range: 10, Accuracy: 70,
		ProjGlyph: '*', ProjColor: "#FFAA00",
	},
	"horus_guard": {
		ID: "horus_guard", Name: "Horus Guard",
		Glyph: 'H', Color: "#DD8800",
		HP: 50, Attack: 14, Defense: 4,
		AI: AIPatrol, AggroRange: 7, Speed: 2,
		XP: 30, LootTable: "serpent_guard", Faction: "goauld",
		IsRanged: true, Range: 10, Accuracy: 70,
		ProjGlyph: '*', ProjColor: "#FFAA00",
	},
	"kull_warrior": {
		ID: "kull_warrior", Name: "Kull Warrior",
		Glyph: 'K', Color: "#444444",
		HP: 120, Attack: 20, Defense: 10,
		AI: AIHunt, AggroRange: 10, Speed: 4,
		XP: 60, LootTable: "kull_warrior", Faction: "goauld",
		IsRanged: true, Range: 8, Accuracy: 75,
		ProjGlyph: '*', ProjColor: "#FF4400",
	},
	"ashrak": {
		ID: "ashrak", Name: "Ashrak",
		Glyph: '?', Color: "#AA44AA",
		HP: 40, Attack: 25, Defense: 3,
		AI: AIAmbush, AggroRange: 3, Speed: 1,
		XP: 45, LootTable: "serpent_guard", Faction: "goauld",
	},
	"jaffa_commander": {
		ID: "jaffa_commander", Name: "Jaffa Commander",
		Glyph: '!', Color: "#FFCC00",
		HP: 70, Attack: 15, Defense: 5,
		AI: AIGuard, AggroRange: 8, Speed: 3,
		XP: 40, LootTable: "serpent_guard", Faction: "goauld",
		IsRanged: true, Range: 10, Accuracy: 75,
		ProjGlyph: '*', ProjColor: "#FFAA00",
	},
	"system_lord": {
		ID: "system_lord", Name: "System Lord",
		Glyph: 'Q', Color: "#FFD700",
		HP: 200, Attack: 30, Defense: 8,
		AI: AIGuard, AggroRange: 10, Speed: 3,
		XP: 200, LootTable: "boss_goauld", Faction: "goauld",
		IsBoss: true,
		IsRanged: true, Range: 5, Accuracy: 85,
		ProjGlyph: '☼', ProjColor: "#FFDD00",
	},

	// ─── Replicators ────────────────────────────────────────
	"replicator_bug": {
		ID: "replicator_bug", Name: "Replicator Bug",
		Glyph: 'x', Color: "#BBBBBB",
		HP: 10, Attack: 5, Defense: 1,
		AI: AISwarm, AggroRange: 8, Speed: 1,
		XP: 5, LootTable: "replicator", Faction: "replicator",
	},
	"replicator_soldier": {
		ID: "replicator_soldier", Name: "Replicator Soldier",
		Glyph: 'X', Color: "#AAAAAA",
		HP: 80, Attack: 18, Defense: 6,
		AI: AIHunt, AggroRange: 10, Speed: 2,
		XP: 50, LootTable: "replicator", Faction: "replicator",
		IsRanged: true, Range: 8, Accuracy: 80,
		ProjGlyph: '-', ProjColor: "#CCCCCC",
	},
	"human_replicator": {
		ID: "human_replicator", Name: "Human-Form Replicator",
		Glyph: 'R', Color: "#CCCCCC",
		HP: 150, Attack: 25, Defense: 8,
		AI: AIHunt, AggroRange: 12, Speed: 2,
		XP: 100, LootTable: "replicator", Faction: "replicator",
		IsBoss: true,
		IsRanged: true, Range: 6, Accuracy: 85,
		ProjGlyph: '~', ProjColor: "#DDDDDD",
	},

	// ─── Ori Forces ─────────────────────────────────────────
	"ori_soldier": {
		ID: "ori_soldier", Name: "Ori Soldier",
		Glyph: 'o', Color: "#FF8800",
		HP: 45, Attack: 14, Defense: 4,
		AI: AIPatrol, AggroRange: 7, Speed: 3,
		XP: 20, LootTable: "ori_common", Faction: "ori",
		IsRanged: true, Range: 10, Accuracy: 75,
		ProjGlyph: '*', ProjColor: "#FF8800",
	},
	"ori_prior": {
		ID: "ori_prior", Name: "Ori Prior",
		Glyph: 'P', Color: "#FF6600",
		HP: 100, Attack: 20, Defense: 5,
		AI: AISupport, AggroRange: 10, Speed: 4,
		XP: 60, LootTable: "ori_rare", Faction: "ori",
		IsRanged: true, Range: 8, Accuracy: 80,
		ProjGlyph: '~', ProjColor: "#FFAA44",
	},
	"ori_commander": {
		ID: "ori_commander", Name: "Ori Commander",
		Glyph: 'O', Color: "#FF4400",
		HP: 150, Attack: 22, Defense: 6,
		AI: AIGuard, AggroRange: 10, Speed: 3,
		XP: 80, LootTable: "ori_rare", Faction: "ori",
		IsBoss: true,
		IsRanged: true, Range: 12, Accuracy: 80,
		ProjGlyph: '*', ProjColor: "#FF6600",
	},

	// ─── Wildlife & Neutral ─────────────────────────────────
	"unas": {
		ID: "unas", Name: "Unas",
		Glyph: 'U', Color: "#886644",
		HP: 60, Attack: 18, Defense: 4,
		AI: AIGuard, AggroRange: 4, Speed: 3,
		XP: 25, LootTable: "wildlife", Faction: "wildlife",
	},
	"giant_scarab": {
		ID: "giant_scarab", Name: "Giant Scarab",
		Glyph: 's', Color: "#666622",
		HP: 15, Attack: 6, Defense: 1,
		AI: AISwarm, AggroRange: 5, Speed: 1,
		XP: 5, LootTable: "wildlife", Faction: "wildlife",
	},
	"crystal_entity": {
		ID: "crystal_entity", Name: "Crystal Entity",
		Glyph: '◇', Color: "#88CCFF",
		HP: 40, Attack: 10, Defense: 3,
		AI: AIPatrol, AggroRange: 5, Speed: 3,
		XP: 20, LootTable: "wildlife", Faction: "wildlife",
	},
	"sodan_warrior": {
		ID: "sodan_warrior", Name: "Sodan Warrior",
		Glyph: '?', Color: "#44AA44",
		HP: 55, Attack: 16, Defense: 4,
		AI: AIAmbush, AggroRange: 3, Speed: 2,
		XP: 30, LootTable: "jaffa_common", Faction: "neutral",
	},
}
