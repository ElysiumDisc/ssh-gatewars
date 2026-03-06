package simulation

import "ssh-gatewars/internal/faction"

const (
	TerritoryGridW = 40 // 400 / 10
	TerritoryGridH = 20 // 200 / 10
	TerritoryCellW = WorldW / TerritoryGridW
	TerritoryCellH = WorldH / TerritoryGridH
)

// TerritoryMap holds zone ownership data.
type TerritoryMap struct {
	Zones    [TerritoryGridW][TerritoryGridH]int // faction ID or -1
	Percents [faction.Count]float64
}

// NewTerritoryMap creates an empty territory map.
func NewTerritoryMap() *TerritoryMap {
	t := &TerritoryMap{}
	for x := range t.Zones {
		for y := range t.Zones[x] {
			t.Zones[x][y] = -1
		}
	}
	// Start with equal distribution
	for i := range t.Percents {
		t.Percents[i] = 20.0
	}
	return t
}

// CalculateTerritory computes zone ownership from ship positions.
func CalculateTerritory(ships []*Ship) *TerritoryMap {
	t := &TerritoryMap{}
	var counts [TerritoryGridW][TerritoryGridH][faction.Count]int

	for _, s := range ships {
		if s.State != Alive {
			continue
		}
		zx := int(s.X) / TerritoryCellW
		zy := int(s.Y) / TerritoryCellH
		if zx < 0 || zx >= TerritoryGridW || zy < 0 || zy >= TerritoryGridH {
			continue
		}
		counts[zx][zy][s.Faction]++
	}

	totalZones := float64(TerritoryGridW * TerritoryGridH)
	var factionZones [faction.Count]int

	for x := 0; x < TerritoryGridW; x++ {
		for y := 0; y < TerritoryGridH; y++ {
			best := -1
			bestCount := 0
			tied := false
			for f := 0; f < faction.Count; f++ {
				if counts[x][y][f] > bestCount {
					bestCount = counts[x][y][f]
					best = f
					tied = false
				} else if counts[x][y][f] == bestCount && counts[x][y][f] > 0 {
					tied = true
				}
			}
			if tied || bestCount == 0 {
				t.Zones[x][y] = -1 // contested or empty
			} else {
				t.Zones[x][y] = best
				factionZones[best]++
			}
		}
	}

	for f := 0; f < faction.Count; f++ {
		t.Percents[f] = float64(factionZones[f]) / totalZones * 100
	}

	return t
}

// ZoneAt returns the owning faction of the zone containing world coords, or -1.
func (t *TerritoryMap) ZoneAt(wx, wy int) int {
	zx := wx / TerritoryCellW
	zy := wy / TerritoryCellH
	if zx < 0 || zx >= TerritoryGridW || zy < 0 || zy >= TerritoryGridH {
		return -1
	}
	return t.Zones[zx][zy]
}
