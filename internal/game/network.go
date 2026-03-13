package game

import (
	"math"
	"sort"
)

// GateLink represents a Stargate connection between two planets.
type GateLink struct {
	FromID   int
	ToID     int
	Distance float64
	Level    int // 0=base, 1-3=upgraded
	RouteID  int // which route this belongs to
}

// GateRoute is a named sequence of planets forming a "line" on the tube map.
type GateRoute struct {
	ID      int
	Name    string
	Color   string // lipgloss color hex
	Planets []int  // ordered planet IDs along this route
}

// GalaxyNetwork holds the Stargate network graph.
type GalaxyNetwork struct {
	Links  []GateLink
	Routes []GateRoute
}

// TransferBonus identifies a type of resource transfer.
type TransferBonus int

const (
	TransferShieldBoost TransferBonus = iota // +20 shield HP to all chairs
	TransferDroneBoost                       // +2 temporary bonus drones for 60s
	TransferZPMDrop                          // add ZPM to planet bounty pool
)

// TransferCosts maps each bonus type to its ZPM cost.
var TransferCosts = map[TransferBonus]int{
	TransferShieldBoost: 30,
	TransferDroneBoost:  50,
	TransferZPMDrop:     25,
}

// GateLinkUpgradeCosts is the ZPM cost to reach each level (0 = free/base).
var GateLinkUpgradeCosts = [4]int{0, 50, 150, 400}

// GateLinkBonus describes defense bonuses for a given link level.
type GateLinkBonus struct {
	ShieldRegen    float64 // fraction of maxShield healed per tick
	DamageBoost    float64 // multiplier added to drone damage (0.05 = +5%)
	SpawnReduction float64 // fraction reduction in replicator spawns (0.10 = -10%)
}

// GateLinkBonuses defines bonuses per upgrade level.
var GateLinkBonuses = [4]GateLinkBonus{
	{0, 0, 0},             // level 0: no bonus
	{0.001, 0.05, 0},      // level 1: tiny regen, +5% damage
	{0.002, 0.10, 0.10},   // level 2: more regen, +10% damage, -10% spawns
	{0.005, 0.15, 0.20},   // level 3: strong regen, +15% damage, -20% spawns
}

// Route definitions — Stargate lore themed.
var routeDefs = []struct {
	Name  string
	Color string
}{
	{"Milky Way Core", "#00D9FF"},   // cyan
	{"Pegasus Rim", "#FF00FF"},      // magenta
	{"Ori Frontier", "#FF3333"},     // red
	{"Asgard Reach", "#00FF88"},     // green
	{"Nox Passage", "#FFD700"},      // gold
	{"Tok'ra Circuit", "#8899AA"},   // silver
}

// GenerateNetwork builds a Stargate network from planet positions.
// It creates an MST for connectivity, adds short extra edges, then groups into routes.
func GenerateNetwork(planets []*Planet) *GalaxyNetwork {
	n := len(planets)
	if n < 2 {
		return &GalaxyNetwork{}
	}

	// Build all pairwise edges sorted by distance
	type edge struct {
		from, to int
		dist     float64
	}
	var edges []edge
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			dx := planets[i].Pos.X - planets[j].Pos.X
			dy := planets[i].Pos.Y - planets[j].Pos.Y
			d := math.Sqrt(dx*dx + dy*dy)
			edges = append(edges, edge{i, j, d})
		}
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].dist < edges[j].dist
	})

	// Kruskal's MST using union-find
	parent := make([]int, n)
	rank := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) bool {
		ra, rb := find(a), find(b)
		if ra == rb {
			return false
		}
		if rank[ra] < rank[rb] {
			ra, rb = rb, ra
		}
		parent[rb] = ra
		if rank[ra] == rank[rb] {
			rank[ra]++
		}
		return true
	}

	// Adjacency count per planet
	adjCount := make([]int, n)
	var links []GateLink
	linkSet := make(map[[2]int]bool)

	addLink := func(from, to int, dist float64) {
		key := [2]int{from, to}
		if from > to {
			key = [2]int{to, from}
		}
		if linkSet[key] {
			return
		}
		linkSet[key] = true
		links = append(links, GateLink{
			FromID:   from,
			ToID:     to,
			Distance: dist,
		})
		adjCount[from]++
		adjCount[to]++
	}

	// Phase 1: MST
	for _, e := range edges {
		if union(e.from, e.to) {
			addLink(e.from, e.to, e.dist)
		}
	}

	// Phase 2: Add short extra edges for redundancy (max 4 connections per planet)
	maxDist := 12.0
	maxAdj := 4
	for _, e := range edges {
		if linkSet[[2]int{MinI(e.from, e.to), MaxI(e.from, e.to)}] {
			continue
		}
		if e.dist > maxDist {
			break // edges are sorted by distance
		}
		if adjCount[e.from] >= maxAdj || adjCount[e.to] >= maxAdj {
			continue
		}
		addLink(e.from, e.to, e.dist)
	}

	// Build adjacency list
	adj := make([][]int, n)
	for _, l := range links {
		adj[l.FromID] = append(adj[l.FromID], l.ToID)
		adj[l.ToID] = append(adj[l.ToID], l.FromID)
	}

	// Group links into routes by greedy path tracing
	routes := buildRoutes(planets, links, adj, n)

	// Assign route IDs to links
	linkRouteMap := make(map[[2]int]int)
	for ri, route := range routes {
		for k := 0; k < len(route.Planets)-1; k++ {
			a, b := route.Planets[k], route.Planets[k+1]
			key := [2]int{MinI(a, b), MaxI(a, b)}
			linkRouteMap[key] = ri
		}
	}
	for i := range links {
		key := [2]int{MinI(links[i].FromID, links[i].ToID), MaxI(links[i].FromID, links[i].ToID)}
		if rid, ok := linkRouteMap[key]; ok {
			links[i].RouteID = rid
		}
	}

	return &GalaxyNetwork{
		Links:  links,
		Routes: routes,
	}
}

// buildRoutes groups the network into named routes by greedy path tracing.
func buildRoutes(planets []*Planet, links []GateLink, adj [][]int, n int) []GateRoute {
	used := make(map[[2]int]bool)
	var routes []GateRoute

	// Sort planets by angle from center to seed route starting points
	type planetAngle struct {
		id    int
		angle float64
	}
	var byAngle []planetAngle
	for _, p := range planets {
		a := math.Atan2(p.Pos.Y, p.Pos.X)
		byAngle = append(byAngle, planetAngle{p.ID, a})
	}
	sort.Slice(byAngle, func(i, j int) bool {
		return byAngle[i].angle < byAngle[j].angle
	})

	numRoutes := len(routeDefs)
	if numRoutes > n {
		numRoutes = n
	}

	// Pick starting planets spaced around the galaxy
	startPlanets := make([]int, numRoutes)
	step := len(byAngle) / numRoutes
	if step < 1 {
		step = 1
	}
	for i := 0; i < numRoutes; i++ {
		idx := (i * step) % len(byAngle)
		startPlanets[i] = byAngle[idx].id
	}

	// Trace each route greedily
	for ri := 0; ri < numRoutes; ri++ {
		var path []int
		current := startPlanets[ri]
		visited := make(map[int]bool)

		path = append(path, current)
		visited[current] = true

		// Extend in one direction
		for iterations := 0; iterations < n; iterations++ {
			best := -1
			bestDist := math.MaxFloat64
			for _, nb := range adj[current] {
				if visited[nb] {
					continue
				}
				key := [2]int{MinI(current, nb), MaxI(current, nb)}
				if used[key] {
					continue
				}
				dx := planets[nb].Pos.X - planets[current].Pos.X
				dy := planets[nb].Pos.Y - planets[current].Pos.Y
				d := math.Sqrt(dx*dx + dy*dy)
				if d < bestDist {
					bestDist = d
					best = nb
				}
			}
			if best < 0 {
				break
			}
			key := [2]int{MinI(current, best), MaxI(current, best)}
			used[key] = true
			path = append(path, best)
			visited[best] = true
			current = best
		}

		if len(path) < 2 {
			continue
		}

		rd := routeDefs[ri%len(routeDefs)]
		routes = append(routes, GateRoute{
			ID:      ri,
			Name:    rd.Name,
			Color:   rd.Color,
			Planets: path,
		})
	}

	return routes
}

func MinI(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
