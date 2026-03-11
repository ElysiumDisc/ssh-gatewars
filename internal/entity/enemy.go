package entity

import (
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// EntityID is a unique identifier for entities on a planet.
type EntityID uint32

// AIState describes an enemy's current behavior.
type AIState int

const (
	AIStateIdle    AIState = iota // stationary guard
	AIStatePatrol                 // wander along path
	AIStateAlert                  // player detected at edge, "looking around"
	AIStateChase                  // pursuing player
	AIStateAttack                 // within weapon range, firing
	AIStateFlee                   // retreating (low HP)
	AIStateRegroup                // moving toward commander
	AIStateStunned                // temporarily disabled
)

// Enemy is an NPC enemy instance on a planet.
type Enemy struct {
	ID        EntityID
	DefID     string // references gamedata.Enemies key
	HP        int
	MaxHP     int
	Pos       core.Pos
	State     AIState
	TickCD    int      // ticks until next action
	Target    core.Pos // chase/attack target (player position)
	PatrolDir core.Pos // current patrol direction
	LastKnown core.Pos // last known player position (for ALERT → CHASE)
	AlertTicks int     // ticks remaining in ALERT state
	StunTicks  int     // ticks remaining stunned
	ShotCD     int     // ticks until next ranged shot
}

// Def returns the enemy definition.
func (e *Enemy) Def() gamedata.EnemyDef {
	return gamedata.Enemies[e.DefID]
}

// NewEnemy creates an enemy instance from a definition.
func NewEnemy(id EntityID, defID string, pos core.Pos) *Enemy {
	def := gamedata.Enemies[defID]
	initialState := AIStatePatrol
	switch def.AI {
	case gamedata.AIGuard, gamedata.AISentry:
		initialState = AIStateIdle
	}
	return &Enemy{
		ID:        id,
		DefID:     defID,
		HP:        def.HP,
		MaxHP:     def.HP,
		Pos:       pos,
		State:     initialState,
		TickCD:    def.Speed,
		PatrolDir: core.DirRight,
	}
}

// IsDead returns true if HP is 0.
func (e *Enemy) IsDead() bool {
	return e.HP <= 0
}

// TakeDamage reduces HP and returns true if the enemy dies.
func (e *Enemy) TakeDamage(dmg int) bool {
	e.HP -= dmg
	if e.HP <= 0 {
		e.HP = 0
		return true
	}
	return false
}

// AttackPower returns the enemy's attack stat.
func (e *Enemy) AttackPower() int {
	return e.Def().Attack
}

// DefensePower returns the enemy's defense stat.
func (e *Enemy) DefensePower() int {
	return e.Def().Defense
}

// IsStunned returns true if the enemy is currently stunned.
func (e *Enemy) IsStunned() bool {
	return e.StunTicks > 0
}

// Stun applies a stun for the given number of ticks.
func (e *Enemy) Stun(ticks int) {
	e.StunTicks = ticks
	e.State = AIStateStunned
}

// TickStun decrements stun counter. Returns true when stun expires.
func (e *Enemy) TickStun() bool {
	if e.StunTicks > 0 {
		e.StunTicks--
		if e.StunTicks == 0 {
			e.State = AIStatePatrol
			return true
		}
	}
	return false
}

// ShouldFlee returns true if enemy HP is below 25%.
func (e *Enemy) ShouldFlee() bool {
	return e.HP > 0 && e.HP*4 <= e.MaxHP
}

// CanShoot returns true if ranged enemy is ready to fire.
func (e *Enemy) CanShoot() bool {
	def := e.Def()
	return def.IsRanged && e.ShotCD <= 0
}
