package simulation

import "math"

const (
	FactoryCostBase   = 10.0
	MissileBaseCost   = 50.0
	PopGrowthRate     = 0.005
	WastePerPop       = 0.001
	BaseFactoryPerPop = 10
)

// Colony represents a colonized planet.
type Colony struct {
	SystemID   int
	Faction    int
	Population float64
	MaxPop     int
	Factories  int
	MaxFactory int
	Waste      float64

	SliderShip     int
	SliderDefense  int
	SliderIndustry int
	SliderEcology  int
	SliderResearch int

	MissileBases int
	ShieldLevel  int
	HasStarbase  bool

	BuildQueue    []BuildOrder
	BuildProgress float64

	IndustryAccum float64
	DefenseAccum  float64
}

// BuildOrder represents an item in the ship build queue.
type BuildOrder struct {
	Name string
	Cost float64
}

// NewColony creates a colony with starting values.
func NewColony(systemID, factionID, maxPop int) *Colony {
	maxFact := 5 * BaseFactoryPerPop
	return &Colony{
		SystemID:       systemID,
		Faction:        factionID,
		Population:     5.0,
		MaxPop:         maxPop,
		Factories:      10,
		MaxFactory:     maxFact,
		SliderShip:     0,
		SliderDefense:  10,
		SliderIndustry: 30,
		SliderEcology:  20,
		SliderResearch: 40,
	}
}

// TotalOutput computes the colony's total production per second.
func (c *Colony) TotalOutput(mineralMult, factionProdMod float64, onlinePlayers int) float64 {
	workers := math.Min(c.Population, float64(c.Factories))
	base := workers * mineralMult * factionProdMod
	online := float64(onlinePlayers)
	if online > 10 {
		online = 10
	}
	bonus := 1.0 + online*0.05
	return base * bonus
}

// UpdateMaxFactory recalculates max factories based on current population.
func (c *Colony) UpdateMaxFactory(factionBonus int) {
	c.MaxFactory = int(c.Population) * (BaseFactoryPerPop + factionBonus)
	if c.MaxFactory < 1 {
		c.MaxFactory = 1
	}
}
