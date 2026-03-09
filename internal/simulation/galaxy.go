package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"ssh-gatewars/internal/gamedata"
)

const (
	StarYellow = iota
	StarRed
	StarBlue
	StarWhite
	StarBinary
	StarTypeCount
)

const (
	SpecialNone = iota
	SpecialArtifact
	SpecialDakara
)

var StarTypeNames = [StarTypeCount]string{"Yellow", "Red", "Blue", "White", "Binary"}
var StarTypeColors = [StarTypeCount]string{"#FFE030", "#FF6030", "#6090FF", "#FFFFFF", "#FFCC00"}

type StarSystem struct {
	ID          int
	Name        string
	GateAddress string
	StarType    int
	MapX, MapY  float64
	Planet      *Planet
	Special     int
}

type Planet struct {
	Type     int
	Size     int
	Minerals int
}

type StargateLink struct {
	From, To int
}

type Galaxy struct {
	Systems   []*StarSystem
	Gates     []StargateLink
	Adjacency map[int][]int
	Seed      int64
}

var homeworldNames = [5]string{"Earth", "Chulak", "Othala", "Vorash", "Cal Mah"}
var homeworldPositions = [5][2]float64{
	{0.50, 0.12},
	{0.85, 0.38},
	{0.73, 0.82},
	{0.27, 0.82},
	{0.15, 0.38},
}

var sgPlanetNames = []string{
	"Abydos", "Langara", "Tollana", "Heliopolis", "Proclarush",
	"Vis Uban", "Tartarus", "Erebus", "Hala", "Orilla",
	"Edora", "Simarka", "Argos", "Cartago", "Velona",
	"Pangar", "Hebridan", "Latona", "Madrona", "Tagrea",
	"Revanna", "Aschen", "Vagonbrei", "Bedrosia", "Vyus",
	"Hadante", "Altair", "Netu", "Kelowna", "Oannes",
}

// NewGalaxy procedurally generates a galaxy.
func NewGalaxy(seed int64, systemCount int) *Galaxy {
	rng := rand.New(rand.NewSource(seed))
	g := &Galaxy{
		Seed:      seed,
		Adjacency: make(map[int][]int),
	}

	if systemCount < 20 {
		systemCount = 20
	}
	if systemCount > 100 {
		systemCount = 100
	}

	// 1. Place 5 homeworlds
	for i := 0; i < 5; i++ {
		jx := (rng.Float64() - 0.5) * 0.04
		jy := (rng.Float64() - 0.5) * 0.04
		sys := &StarSystem{
			ID:          i,
			Name:        homeworldNames[i],
			GateAddress: gateAddr(rng),
			StarType:    StarYellow,
			MapX:        clamp01(homeworldPositions[i][0] + jx),
			MapY:        clamp01(homeworldPositions[i][1] + jy),
			Planet: &Planet{
				Type:     gamedata.PlanetTerran,
				Size:     gamedata.SizeMedium,
				Minerals: gamedata.MineralNormal,
			},
		}
		g.Systems = append(g.Systems, sys)
	}

	// 2. Place Dakara near center
	g.Systems = append(g.Systems, &StarSystem{
		ID:          5,
		Name:        "Dakara",
		GateAddress: gateAddr(rng),
		StarType:    StarWhite,
		MapX:        0.50 + (rng.Float64()-0.5)*0.06,
		MapY:        0.50 + (rng.Float64()-0.5)*0.06,
		Planet: &Planet{
			Type:     gamedata.PlanetTerran,
			Size:     gamedata.SizeHuge,
			Minerals: gamedata.MineralUltraRich,
		},
		Special: SpecialDakara,
	})

	// Shuffle named pool
	names := make([]string, len(sgPlanetNames))
	copy(names, sgPlanetNames)
	rng.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })
	nameIdx := 0

	// 3. Fill remaining systems
	maxAttempts := systemCount * 200
	for len(g.Systems) < systemCount && maxAttempts > 0 {
		maxAttempts--
		x := rng.Float64()*0.88 + 0.06
		y := rng.Float64()*0.88 + 0.06

		tooClose := false
		for _, s := range g.Systems {
			dx := x - s.MapX
			dy := y - s.MapY
			if math.Sqrt(dx*dx+dy*dy) < 0.07 {
				tooClose = true
				break
			}
		}
		if tooClose {
			continue
		}

		id := len(g.Systems)
		var name string
		if nameIdx < len(names) {
			name = names[nameIdx]
			nameIdx++
		} else {
			name = procName(rng)
		}

		sys := &StarSystem{
			ID:          id,
			Name:        name,
			GateAddress: gateAddr(rng),
			StarType:    randStarType(rng),
			MapX:        x,
			MapY:        y,
		}

		if rng.Float64() < 0.90 {
			sys.Planet = randPlanet(rng)
		}

		if rng.Float64() < 0.05 {
			sys.Special = SpecialArtifact
		}

		g.Systems = append(g.Systems, sys)
	}

	// 4. Build gate network
	g.buildGateNetwork()

	return g
}

func (g *Galaxy) buildGateNetwork() {
	n := len(g.Systems)
	if n == 0 {
		return
	}

	connected := make(map[[2]int]bool)

	type distPair struct {
		id   int
		dist float64
	}

	for i := 0; i < n; i++ {
		var pairs []distPair
		for j := 0; j < n; j++ {
			if i == j {
				continue
			}
			dx := g.Systems[i].MapX - g.Systems[j].MapX
			dy := g.Systems[i].MapY - g.Systems[j].MapY
			pairs = append(pairs, distPair{j, math.Sqrt(dx*dx + dy*dy)})
		}
		sort.Slice(pairs, func(a, b int) bool { return pairs[a].dist < pairs[b].dist })

		links := 0
		for _, p := range pairs {
			if links >= 3 {
				break
			}
			key := [2]int{min(i, p.id), max(i, p.id)}
			if connected[key] {
				links++
				continue
			}
			connected[key] = true
			g.Gates = append(g.Gates, StargateLink{From: key[0], To: key[1]})
			links++
		}
	}

	for _, gate := range g.Gates {
		g.Adjacency[gate.From] = append(g.Adjacency[gate.From], gate.To)
		g.Adjacency[gate.To] = append(g.Adjacency[gate.To], gate.From)
	}

	// Ensure connectivity
	visited := make([]bool, n)
	queue := []int{0}
	visited[0] = true
	for head := 0; head < len(queue); head++ {
		for _, nb := range g.Adjacency[queue[head]] {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, nb)
			}
		}
	}

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}
		bestDist := math.MaxFloat64
		bestJ := 0
		for j := 0; j < n; j++ {
			if !visited[j] {
				continue
			}
			dx := g.Systems[i].MapX - g.Systems[j].MapX
			dy := g.Systems[i].MapY - g.Systems[j].MapY
			d := math.Sqrt(dx*dx + dy*dy)
			if d < bestDist {
				bestDist = d
				bestJ = j
			}
		}
		key := [2]int{min(i, bestJ), max(i, bestJ)}
		if !connected[key] {
			connected[key] = true
			g.Gates = append(g.Gates, StargateLink{From: key[0], To: key[1]})
			g.Adjacency[key[0]] = append(g.Adjacency[key[0]], key[1])
			g.Adjacency[key[1]] = append(g.Adjacency[key[1]], key[0])
		}
		visited[i] = true
		bfsQ := []int{i}
		for h := 0; h < len(bfsQ); h++ {
			for _, nb := range g.Adjacency[bfsQ[h]] {
				if !visited[nb] {
					visited[nb] = true
					bfsQ = append(bfsQ, nb)
				}
			}
		}
	}
}

func randStarType(rng *rand.Rand) int {
	r := rng.Float64()
	switch {
	case r < 0.35:
		return StarYellow
	case r < 0.60:
		return StarRed
	case r < 0.75:
		return StarBlue
	case r < 0.90:
		return StarWhite
	default:
		return StarBinary
	}
}

func randPlanet(rng *rand.Rand) *Planet {
	var pType int
	if rng.Float64() < 0.60 {
		pType = rng.Intn(7)
	} else {
		pType = 7 + rng.Intn(6)
	}

	var size int
	sr := rng.Float64()
	switch {
	case sr < 0.10:
		size = gamedata.SizeTiny
	case sr < 0.30:
		size = gamedata.SizeSmall
	case sr < 0.60:
		size = gamedata.SizeMedium
	case sr < 0.85:
		size = gamedata.SizeLarge
	default:
		size = gamedata.SizeHuge
	}

	var minerals int
	mr := rng.Float64()
	switch {
	case mr < 0.05:
		minerals = gamedata.MineralUltraPoor
	case mr < 0.20:
		minerals = gamedata.MineralPoor
	case mr < 0.65:
		minerals = gamedata.MineralNormal
	case mr < 0.85:
		minerals = gamedata.MineralRich
	default:
		minerals = gamedata.MineralUltraRich
	}

	return &Planet{Type: pType, Size: size, Minerals: minerals}
}

func gateAddr(rng *rand.Rand) string {
	return fmt.Sprintf("P%d%c-%03d", rng.Intn(9)+1, rune('A'+rng.Intn(26)), rng.Intn(1000))
}

func procName(rng *rand.Rand) string {
	return fmt.Sprintf("P%d%c-%03d", rng.Intn(9)+1, rune('A'+rng.Intn(26)), rng.Intn(1000))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
