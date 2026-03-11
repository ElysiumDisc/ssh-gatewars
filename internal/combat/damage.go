package combat

// CalcDamage computes damage dealt. Minimum 1 if attack > 0.
func CalcDamage(attack, defense int) int {
	dmg := attack - defense
	if dmg < 1 && attack > 0 {
		dmg = 1
	}
	if dmg < 0 {
		dmg = 0
	}
	return dmg
}

// XPForKill returns XP awarded for killing an enemy.
// Scales with threat level.
func XPForKill(baseXP, threat int) int {
	return baseXP + (threat * 2)
}
