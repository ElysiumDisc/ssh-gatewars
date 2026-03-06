package faction

const (
	Tauri  = 0
	Goauld = 1
	Jaffa  = 2
	Lucian = 3
	Asgard = 4
	Count  = 5
)

// DirectionalSymbols holds ship characters for each movement direction.
type DirectionalSymbols struct {
	Right string
	Left  string
	Up    string
	Down  string
}

// Faction defines a fleet's visual identity and base stats.
type Faction struct {
	ID        int
	Name      string
	ShortName string
	ColorFG   string // ANSI 256-color code
	ColorBG   string // dim background tint
	Symbols   DirectionalSymbols
	BaseHP    float32
	BaseDamage float32
	BaseSpeed  float32
	SpawnMult  float32 // spawn rate multiplier (1.0 = normal, 0.33 = Asgard)
}

var Factions = [Count]Faction{
	{
		ID: Tauri, Name: "Tau'ri", ShortName: "TAU'RI",
		ColorFG: "#4A90D9", ColorBG: "#0A1520",
		Symbols: DirectionalSymbols{Right: "->", Left: "<-", Up: "/\\", Down: "\\/"},
		BaseHP: 100, BaseDamage: 10, BaseSpeed: 2.0, SpawnMult: 1.0,
	},
	{
		ID: Goauld, Name: "Goa'uld", ShortName: "GOA'ULD",
		ColorFG: "#D4A017", ColorBG: "#151000",
		Symbols: DirectionalSymbols{Right: "{=>", Left: "<=}", Up: "/^\\", Down: "\\v/"},
		BaseHP: 120, BaseDamage: 8, BaseSpeed: 1.2, SpawnMult: 1.0,
	},
	{
		ID: Jaffa, Name: "Jaffa Rebellion", ShortName: "JAFFA",
		ColorFG: "#E8C820", ColorBG: "#151400",
		Symbols: DirectionalSymbols{Right: ">>", Left: "<<", Up: "^^", Down: "vv"},
		BaseHP: 80, BaseDamage: 9, BaseSpeed: 2.5, SpawnMult: 1.0,
	},
	{
		ID: Lucian, Name: "Lucian Alliance", ShortName: "LUCIAN",
		ColorFG: "#C850C0", ColorBG: "#100014",
		Symbols: DirectionalSymbols{Right: "~>", Left: "<~", Up: "~^", Down: "~v"},
		BaseHP: 90, BaseDamage: 10, BaseSpeed: 2.2, SpawnMult: 1.0,
	},
	{
		ID: Asgard, Name: "Asgard", ShortName: "ASGARD",
		ColorFG: "#40E0D0", ColorBG: "#001414",
		Symbols: DirectionalSymbols{Right: "*->", Left: "<-*", Up: "*|*", Down: "*|*"},
		BaseHP: 300, BaseDamage: 30, BaseSpeed: 1.5, SpawnMult: 0.33,
	},
}

// Symbol returns the directional symbol for a ship's velocity.
func Symbol(factionID int, vx, vy float64) string {
	s := Factions[factionID].Symbols
	if abs(vx) >= abs(vy) {
		if vx >= 0 {
			return s.Right
		}
		return s.Left
	}
	if vy >= 0 {
		return s.Down
	}
	return s.Up
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
