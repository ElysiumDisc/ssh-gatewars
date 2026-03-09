package faction

const (
	Tauri  = 0
	Goauld = 1
	Asgard = 2
	Tokra  = 3
	Jaffa  = 4
	Count  = 5
)

// Faction defines a faction's identity and MOO-style racial traits.
type Faction struct {
	ID        int
	Name      string
	ShortName string
	ColorFG   string
	ColorBG   string

	// Trait multipliers (1.0 = baseline)
	DiplomacyMod  float64
	ResearchMod   float64
	ProductionMod float64
	AttackMod     float64
	DefenseMod    float64
	GroundMod     float64
	PopGrowthMod  float64
	EspionageMod  float64
	FactoryCapMod int     // bonus max factories per pop
	Special       string
}

var Factions = [Count]Faction{
	{
		ID: Tauri, Name: "Tau'ri", ShortName: "TAU'RI",
		ColorFG: "#4A90D9", ColorBG: "#0A1520",
		DiplomacyMod: 1.30, ResearchMod: 1.10, ProductionMod: 1.00,
		AttackMod: 1.00, DefenseMod: 1.00, GroundMod: 1.00,
		PopGrowthMod: 1.00, EspionageMod: 1.00, FactoryCapMod: 0,
		Special: "Allies of convenience (cheaper treaties)",
	},
	{
		ID: Goauld, Name: "Goa'uld", ShortName: "GOA'ULD",
		ColorFG: "#CC2222", ColorBG: "#180000",
		DiplomacyMod: 1.00, ResearchMod: 0.85, ProductionMod: 1.20,
		AttackMod: 1.20, DefenseMod: 1.00, GroundMod: 1.00,
		PopGrowthMod: 1.00, EspionageMod: 1.00, FactoryCapMod: 3,
		Special: "Slave labor (higher factory cap)",
	},
	{
		ID: Asgard, Name: "Asgard", ShortName: "ASGARD",
		ColorFG: "#40E0D0", ColorBG: "#001414",
		DiplomacyMod: 1.00, ResearchMod: 1.50, ProductionMod: 1.00,
		AttackMod: 1.00, DefenseMod: 1.00, GroundMod: 0.70,
		PopGrowthMod: 0.85, EspionageMod: 1.00, FactoryCapMod: 0,
		Special: "Tech sharing (cheaper miniaturization)",
	},
	{
		ID: Tokra, Name: "Tok'ra", ShortName: "TOK'RA",
		ColorFG: "#C850C0", ColorBG: "#100014",
		DiplomacyMod: 1.00, ResearchMod: 1.00, ProductionMod: 0.80,
		AttackMod: 1.00, DefenseMod: 1.00, GroundMod: 1.00,
		PopGrowthMod: 1.00, EspionageMod: 1.50, FactoryCapMod: 0,
		Special: "Infiltration (sabotage, intel)",
	},
	{
		ID: Jaffa, Name: "Jaffa Free Nation", ShortName: "JAFFA",
		ColorFG: "#E8C820", ColorBG: "#151400",
		DiplomacyMod: 1.00, ResearchMod: 0.80, ProductionMod: 1.00,
		AttackMod: 1.00, DefenseMod: 1.10, GroundMod: 1.50,
		PopGrowthMod: 1.00, EspionageMod: 1.00, FactoryCapMod: 0,
		Special: "Warriors (cheaper ground invasions)",
	},
}
