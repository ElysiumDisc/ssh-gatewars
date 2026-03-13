package game

import (
	"fmt"
	"math"
	"math/rand"

	"ssh-gatewars/internal/core"
)

// Galaxy holds the shared state of all planets.
type Galaxy struct {
	Planets []*Planet
	Network *GalaxyNetwork
	Seed    int64
}

// NewGalaxy generates a galaxy with the given number of planets.
func NewGalaxy(seed int64, numPlanets int) *Galaxy {
	rng := rand.New(rand.NewSource(seed))
	planets := make([]*Planet, numPlanets)

	prefixes := []string{
		"P7X", "P3X", "P4X", "P2X", "P9X", "M4R", "M7G", "P5C", "P3Y", "P8X",
	}
	for i := range planets {
		prefix := prefixes[rng.Intn(len(prefixes))]
		name := fmt.Sprintf("%s-%03d", prefix, rng.Intn(999))

		// Place planets in a spiral pattern for visual interest
		angle := float64(i) * 2.4 // golden angle in radians
		radius := 2.0 + math.Sqrt(float64(i))*3.0
		pos := core.Vec2{
			X: math.Cos(angle) * radius,
			Y: math.Sin(angle) * radius,
		}

		planets[i] = &Planet{
			ID:            i,
			Name:          name,
			Seed:          seed + int64(i*7919), // deterministic per-planet seed
			Pos:           pos,
			Status:        PlanetInvaded,
			InvasionLevel: 1 + (i % 10), // difficulty scales
		}
	}

	g := &Galaxy{
		Planets: planets,
		Seed:    seed,
	}
	g.Network = GenerateNetwork(g.Planets)
	return g
}

// FreePlanet marks a planet as liberated.
func (g *Galaxy) FreePlanet(planetID int) {
	if planetID >= 0 && planetID < len(g.Planets) {
		g.Planets[planetID].Status = PlanetFree
	}
}

// InvadedCount returns the number of invaded planets.
func (g *Galaxy) InvadedCount() int {
	count := 0
	for _, p := range g.Planets {
		if p.Status == PlanetInvaded {
			count++
		}
	}
	return count
}

// FreeCount returns the number of free planets.
func (g *Galaxy) FreeCount() int {
	count := 0
	for _, p := range g.Planets {
		if p.Status == PlanetFree {
			count++
		}
	}
	return count
}
