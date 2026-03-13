package game

import "ssh-gatewars/internal/core"

// Chair represents a player's Ancient Control Chair on a planet.
type Chair struct {
	PlayerFP  string      // SSH fingerprint of owner
	Callsign  string      // display name
	Pos       core.Vec2   // position on the defense field (near center)
	Level     int         // chair level (affects drone slots, shield)
	MaxDrones int         // max simultaneous drones
	FireTimer int         // ticks until next drone fires
	ShieldHP  int         // remaining shield health
	MaxShield int         // max shield health
	DroneTier DroneTier   // current drone upgrade tier
	Tactic    DroneTactic // current firing tactic
	Faction   Faction     // Ancient or Ori
}

// NewChair creates a chair for a player at the given position.
func NewChair(fp, callsign string, pos core.Vec2, level int, tier DroneTier, faction Faction) *Chair {
	maxDrones := CalcMaxDronesFaction(faction, level)
	maxShield := CalcMaxShieldFaction(faction, level)
	return &Chair{
		PlayerFP:  fp,
		Callsign:  callsign,
		Pos:       pos,
		Level:     level,
		MaxDrones: maxDrones,
		FireTimer: 0,
		ShieldHP:  maxShield,
		MaxShield: maxShield,
		DroneTier: tier,
		Tactic:    TacticSpread,
		Faction:   faction,
	}
}

// CalcMaxDrones returns max drones for level (legacy, uses Ancient defaults).
func CalcMaxDrones(level int) int {
	return CalcMaxDronesFaction(FactionAncient, level)
}

// CalcMaxShield returns max shield for level (legacy, uses Ancient defaults).
func CalcMaxShield(level int) int {
	return CalcMaxShieldFaction(FactionAncient, level)
}

// EffectiveFireRate returns the fire rate (in ticks) adjusted for faction and level.
func (c *Chair) EffectiveFireRate(baseFR int) int {
	return CalcEffectiveFireRate(c.Faction, c.Level, baseFR)
}

// SalvoCount returns how many drones fire per shot based on faction and level.
func (c *Chair) SalvoCount() int {
	return CalcSalvoCount(c.Faction, c.Level)
}

// DroneDamage returns effective damage for this chair's faction and tier.
func (c *Chair) DroneDamage() int {
	tierCfg := DroneTiers[c.DroneTier]
	return CalcDroneDamage(c.Faction, tierCfg.Damage)
}

// Alive returns true if the chair's shield is above zero.
func (c *Chair) Alive() bool {
	return c.ShieldHP > 0
}

// TakeDamage reduces shield HP. Returns remaining HP.
func (c *Chair) TakeDamage(dmg int) int {
	c.ShieldHP -= dmg
	if c.ShieldHP < 0 {
		c.ShieldHP = 0
	}
	return c.ShieldHP
}
