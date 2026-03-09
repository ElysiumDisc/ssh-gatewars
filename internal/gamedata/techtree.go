package gamedata

// Tech tree indices.
const (
	TreeSGCSystems = iota
	TreeGoauldEng
	TreeAsgardShields
	TreeAncientKnowledge
	TreeHyperdrive
	TreeWeapons
	TreeCount
)

// TreeNames for display.
var TreeNames = [TreeCount]string{
	"SGC Systems",
	"Goa'uld Engineering",
	"Asgard Shields",
	"Ancient Knowledge",
	"Hyperdrive Tech",
	"Weapons",
}

// TreeColors for rendering.
var TreeColors = [TreeCount]string{
	"#4A90D9", // SGC - blue
	"#CC2222", // Goa'uld - red
	"#40E0D0", // Asgard - teal
	"#FFD700", // Ancient - gold
	"#C850C0", // Hyperdrive - purple
	"#FF8800", // Weapons - orange
}

// MaxTier is the maximum tech tier per tree.
const MaxTier = 10

// TierCost returns the research points needed to advance to the given tier.
func TierCost(tier int) float64 {
	if tier <= 0 {
		return 0
	}
	// Exponential scaling: 100, 250, 500, 1000, 2000, 4000, 8000, ...
	base := 100.0
	for i := 1; i < tier; i++ {
		base *= 2.0
		if i <= 2 {
			base += 50.0
		}
	}
	return base
}

// MiniaturizationDiscount returns the cost/size reduction for a component
// when the faction's tech tier exceeds the component's required tier.
// Each tier above reduces by 5%, minimum 50% of original.
func MiniaturizationDiscount(factionTier, componentTier int) float64 {
	if factionTier <= componentTier {
		return 1.0
	}
	discount := 1.0 - float64(factionTier-componentTier)*0.05
	if discount < 0.50 {
		discount = 0.50
	}
	return discount
}
