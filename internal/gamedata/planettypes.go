package gamedata

// Planet type constants.
const (
	PlanetTerran = iota
	PlanetOcean
	PlanetJungle
	PlanetArid
	PlanetSteppe
	PlanetTundra
	PlanetDesert
	PlanetBarren
	PlanetVolcanic
	PlanetToxic
	PlanetInferno
	PlanetDead
	PlanetRadiated
	PlanetTypeCount
)

// Planet size constants.
const (
	SizeTiny = iota
	SizeSmall
	SizeMedium
	SizeLarge
	SizeHuge
	SizeCount
)

// Mineral richness constants.
const (
	MineralUltraPoor = iota
	MineralPoor
	MineralNormal
	MineralRich
	MineralUltraRich
	MineralCount
)

// SizeMaxPop returns the base max population for each planet size.
var SizeMaxPop = [SizeCount]int{2, 4, 7, 10, 14}

// SizeNames for display.
var SizeNames = [SizeCount]string{"Tiny", "Small", "Medium", "Large", "Huge"}

// MineralMultiplier is the production multiplier per mineral richness.
var MineralMultiplier = [MineralCount]float64{0.33, 0.50, 1.00, 1.33, 1.67}

// MineralNames for display.
var MineralNames = [MineralCount]string{"Ultra-Poor", "Poor", "Normal", "Rich", "Ultra-Rich"}

// PlanetTypeInfo describes a planet type.
type PlanetTypeInfo struct {
	Name         string
	Habitable    bool
	PopModifier  float64 // multiplier on SizeMaxPop
	TechRequired string  // empty = colonizable now
	Color        string  // hex color for rendering
}

// PlanetTypes defines all 13 planet types.
var PlanetTypes = [PlanetTypeCount]PlanetTypeInfo{
	{Name: "Terran", Habitable: true, PopModifier: 1.0, Color: "#22AA44"},
	{Name: "Ocean", Habitable: true, PopModifier: 0.9, Color: "#2266CC"},
	{Name: "Jungle", Habitable: true, PopModifier: 0.9, Color: "#118833"},
	{Name: "Arid", Habitable: true, PopModifier: 0.8, Color: "#CC9944"},
	{Name: "Steppe", Habitable: true, PopModifier: 0.8, Color: "#88AA44"},
	{Name: "Tundra", Habitable: true, PopModifier: 0.7, Color: "#88BBDD"},
	{Name: "Desert", Habitable: true, PopModifier: 0.6, Color: "#DDAA44"},
	{Name: "Barren", Habitable: false, PopModifier: 0.4, TechRequired: "controlled_env", Color: "#886644"},
	{Name: "Volcanic", Habitable: false, PopModifier: 0.3, TechRequired: "controlled_env", Color: "#CC4400"},
	{Name: "Toxic", Habitable: false, PopModifier: 0.3, TechRequired: "controlled_env", Color: "#44CC44"},
	{Name: "Inferno", Habitable: false, PopModifier: 0.2, TechRequired: "advanced_env", Color: "#FF4400"},
	{Name: "Dead", Habitable: false, PopModifier: 0.3, TechRequired: "controlled_env", Color: "#666666"},
	{Name: "Radiated", Habitable: false, PopModifier: 0.2, TechRequired: "advanced_env", Color: "#AAFF00"},
}

// MaxPop calculates the maximum population for a planet.
func MaxPop(planetType, size int) int {
	base := SizeMaxPop[size]
	mod := PlanetTypes[planetType].PopModifier
	pop := int(float64(base) * mod)
	if pop < 1 {
		pop = 1
	}
	return pop
}

// BaseFactoryCap returns the max factories per unit of population.
// In MOO1 this is 10 per pop, with robotics tech increasing it.
func BaseFactoryCap() int {
	return 10
}
