package entity

import (
	"math/rand"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// GroundItem is an item lying on the floor of a planet.
type GroundItem struct {
	ID    EntityID
	DefID string
	Qty   int
	Pos   core.Pos
}

// RollLoot picks a random item from a loot table. Returns "" if no drop.
func RollLoot(tableID string, rng *rand.Rand) string {
	table, ok := gamedata.LootTables[tableID]
	if !ok || len(table) == 0 {
		return ""
	}

	totalWeight := 0
	for _, entry := range table {
		totalWeight += entry.Weight
	}
	if totalWeight <= 0 {
		return ""
	}

	roll := rng.Intn(totalWeight)
	for _, entry := range table {
		roll -= entry.Weight
		if roll < 0 {
			return entry.ItemID
		}
	}
	return table[len(table)-1].ItemID
}
