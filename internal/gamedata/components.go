package gamedata

// Component categories.
const (
	CompWeapon  = iota
	CompShield
	CompArmor
	CompEngine
	CompComputer
	CompSpecial
)

// Component defines a ship component.
type Component struct {
	ID       int
	Name     string
	Category int
	TechTier int // required tech tier to unlock (0 = always available)
	TechTree int // which tech tree (0-5)
	Size     int // space used on hull
	Cost     float64
	// Weapon stats
	MinDamage int
	MaxDamage int
	// Shield stats
	ShieldHP int
	// Armor stats
	ArmorHP int
	// Engine stats
	Speed   float64
	Warp    int
	// Computer stats
	AttackBonus int
}

// Starting components (always available, tech tier 0).
var StartingComponents = []Component{
	// Weapons
	{ID: 1, Name: "Staff Cannon", Category: CompWeapon, Size: 5, Cost: 5, MinDamage: 1, MaxDamage: 4},
	{ID: 2, Name: "Railgun", Category: CompWeapon, Size: 10, Cost: 10, MinDamage: 2, MaxDamage: 8},

	// Shields
	{ID: 10, Name: "Naquadah Shield Mk I", Category: CompShield, Size: 8, Cost: 8, ShieldHP: 3},

	// Armor
	{ID: 20, Name: "Trinium Plating", Category: CompArmor, Size: 0, Cost: 3, ArmorHP: 3},

	// Engines
	{ID: 30, Name: "Sublight Drive", Category: CompEngine, Size: 10, Cost: 5, Speed: 1.0, Warp: 1},
	{ID: 31, Name: "Hyperdrive Mk I", Category: CompEngine, Size: 15, Cost: 15, Speed: 2.0, Warp: 2},

	// Computers
	{ID: 40, Name: "Targeting Computer Mk I", Category: CompComputer, Size: 5, Cost: 5, AttackBonus: 1},
}

// DefaultScoutDesign returns the default scout ship design components.
func DefaultScoutDesign() []int {
	return []int{1, 30, 40} // Staff Cannon + Sublight + Targeting Mk I
}

// DefaultFighterDesign returns the default fighter ship design components.
func DefaultFighterDesign() []int {
	return []int{1, 1, 10, 20, 30, 40} // 2x Staff + Shield + Armor + Sublight + Targeting
}
