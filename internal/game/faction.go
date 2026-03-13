package game

// Faction represents the player's chosen path.
type Faction int

const (
	FactionAncient Faction = iota // Ancients — drone swarm, strong shields
	FactionOri                    // Ori — fewer drones, devastating damage, fast fire
)

// FactionInfo holds display and gameplay data for a faction.
type FactionInfo struct {
	Name        string
	Tag         string // short display tag
	Desc        string
	DroneSymbol rune   // primary drone glyph
	// Scaling modifiers
	MaxDronesBase  int     // base drones at level 0
	MaxDronesPerLv int     // drones gained per level
	DamageMult     float64 // multiplier on drone damage
	FireRateBase   int     // base fire rate (ticks)
	FireRateMin    int     // minimum fire rate
	ShieldMult     float64 // multiplier on max shield
	SalvoLevels    [4]int  // level thresholds for salvo 1/2/3/4
}

// FactionDefs maps faction to its gameplay definition.
var FactionDefs = map[Faction]FactionInfo{
	FactionAncient: {
		Name:        "Ancient",
		Tag:         "ANCIENT",
		Desc:        "Drone swarm masters. More drones, stronger shields, balanced damage.",
		DroneSymbol: '✸',
		MaxDronesBase:  5,
		MaxDronesPerLv: 4,  // 5, 9, 13, ... 45 at lv10
		DamageMult:     1.0,
		FireRateBase:   10,
		FireRateMin:    4,
		ShieldMult:     1.25, // +25% shields
		SalvoLevels:    [4]int{0, 2, 5, 8}, // salvo 1 at lv0, 2 at lv2, 3 at lv5, 4 at lv8
	},
	FactionOri: {
		Name:        "Ori",
		Tag:         "ORI",
		Desc:        "Devastating firepower. Fewer drones, but each hits harder and faster.",
		DroneSymbol: '⬥',
		MaxDronesBase:  3,
		MaxDronesPerLv: 2,  // 3, 5, 7, ... 23 at lv10
		DamageMult:     2.0, // double damage
		FireRateBase:   7,   // faster base fire
		FireRateMin:    3,   // faster minimum
		ShieldMult:     0.8, // weaker shields
		SalvoLevels:    [4]int{0, 3, 6, 9}, // slower salvo progression
	},
}

// FactionNames returns display names for all factions.
var FactionNames = map[Faction]string{
	FactionAncient: "Ancient",
	FactionOri:     "Ori",
}

// CalcMaxDronesFaction returns max drones for a faction and level.
func CalcMaxDronesFaction(faction Faction, level int) int {
	f := FactionDefs[faction]
	return f.MaxDronesBase + level*f.MaxDronesPerLv
}

// CalcMaxShieldFaction returns max shield for a faction and level.
func CalcMaxShieldFaction(faction Faction, level int) int {
	f := FactionDefs[faction]
	base := 10 + level*5
	return int(float64(base) * f.ShieldMult)
}

// CalcEffectiveFireRate returns fire rate for a faction, level, and base rate.
func CalcEffectiveFireRate(faction Faction, level, baseFR int) int {
	f := FactionDefs[faction]
	rate := f.FireRateBase - level
	if rate < f.FireRateMin {
		rate = f.FireRateMin
	}
	return rate
}

// CalcSalvoCount returns salvo size for a faction and level.
func CalcSalvoCount(faction Faction, level int) int {
	f := FactionDefs[faction]
	salvo := 1
	for i := 3; i >= 0; i-- {
		if level >= f.SalvoLevels[i] {
			salvo = i + 1
			break
		}
	}
	return salvo
}

// CalcDroneDamage returns effective damage for a faction and base damage.
func CalcDroneDamage(faction Faction, baseDmg int) int {
	f := FactionDefs[faction]
	dmg := int(float64(baseDmg) * f.DamageMult)
	if dmg < 1 {
		dmg = 1
	}
	return dmg
}
