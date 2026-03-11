package combat

import (
	"math/rand"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/world"
)

// RangedAttackResult describes the outcome of a ranged attack.
type RangedAttackResult struct {
	Hit       bool
	Damage    int
	Blocked   bool // LOS blocked
	OutOfRange bool
	CoverPct  int // cover percentage that was applied
}

// ResolveRangedAttack calculates whether a ranged shot hits, accounting for
// range, accuracy, cover, and line of sight.
func ResolveRangedAttack(
	attacker, target core.Pos,
	weapon gamedata.ItemDef,
	targetDefense int,
	tm *world.TileMap,
	rng *rand.Rand,
) RangedAttackResult {
	// Check range
	dist := attacker.ManhattanDist(target)
	if dist > weapon.Range {
		return RangedAttackResult{OutOfRange: true}
	}

	// Check LOS
	if !tm.HasLOS(attacker, target) {
		return RangedAttackResult{Blocked: true}
	}

	// Base accuracy: 85% at close (1-3), drops 5% per tile beyond that
	accuracy := weapon.Accuracy
	if accuracy == 0 {
		accuracy = 85
	}
	if dist > 3 {
		accuracy -= (dist - 3) * 5
	}

	// Apply cover penalty
	cover := CoverValue(attacker, target, tm)
	accuracy -= cover

	// Clamp
	if accuracy < 5 {
		accuracy = 5
	}
	if accuracy > 95 {
		accuracy = 95
	}

	// Roll hit
	roll := rng.Intn(100)
	if roll >= accuracy {
		return RangedAttackResult{
			Hit:      false,
			CoverPct: cover,
		}
	}

	// Calculate damage
	dmg := CalcDamage(weapon.Attack, targetDefense)

	return RangedAttackResult{
		Hit:      true,
		Damage:   dmg,
		CoverPct: cover,
	}
}
