package world

import (
	"math/rand"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// BiomeForAddress picks a biome based on the gate address seed and threat level.
func BiomeForAddress(addr gamedata.GateAddress) (gamedata.Biome, int) {
	seed := addr.Seed()
	rng := rand.New(rand.NewSource(seed))

	// Threat level 1-10 based on address hash
	threat := 1 + rng.Intn(10)

	// Filter biomes that match the threat range
	var candidates []gamedata.Biome
	for _, b := range gamedata.Biomes {
		if threat >= b.MinThreat && threat <= b.MaxThreat {
			candidates = append(candidates, b)
		}
	}
	if len(candidates) == 0 {
		candidates = gamedata.Biomes[:1] // fallback to desert
	}

	biome := candidates[rng.Intn(len(candidates))]
	return biome, threat
}

// EnemiesForPlanet returns enemy spawn positions and types for a planet.
func EnemiesForPlanet(m *TileMap, biome gamedata.Biome, threat int, seed int64) []EnemySpawn {
	rng := rand.New(rand.NewSource(seed + 9999))

	floors := m.FloorPositions()
	if len(floors) == 0 {
		return nil
	}

	// More enemies at higher threat
	count := (len(floors) * threat) / 200
	if count < 2 {
		count = 2
	}
	if count > 20 {
		count = 20
	}

	var spawns []EnemySpawn
	used := make(map[core.Pos]bool)
	used[m.GatePos] = true
	used[m.SpawnPos] = true

	for i := 0; i < count; i++ {
		// Pick random floor position away from gate
		for attempts := 0; attempts < 50; attempts++ {
			pos := floors[rng.Intn(len(floors))]
			if used[pos] {
				continue
			}
			if pos.ManhattanDist(m.GatePos) < 5 {
				continue // don't spawn too close to gate
			}

			// Pick enemy type from biome
			enemyID := biome.Enemies[rng.Intn(len(biome.Enemies))]

			// Higher threat = stronger enemy variants more likely
			if threat >= 7 && rng.Float64() < 0.3 {
				// Try to pick a harder enemy
				for _, e := range biome.Enemies {
					def := gamedata.Enemies[e]
					if def.HP > 12 {
						enemyID = e
						break
					}
				}
			}

			spawns = append(spawns, EnemySpawn{
				Pos:     pos,
				EnemyID: enemyID,
			})
			used[pos] = true
			break
		}
	}

	return spawns
}

// EnemySpawn represents where an enemy should be placed.
type EnemySpawn struct {
	Pos     core.Pos
	EnemyID string
}
