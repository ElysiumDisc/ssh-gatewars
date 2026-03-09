package simulation

import "ssh-gatewars/internal/gamedata"

// FactionTech tracks research progress for a faction.
type FactionTech struct {
	Allocation  [gamedata.TreeCount]int
	CurrentTier [gamedata.TreeCount]int
	ResearchPts [gamedata.TreeCount]float64
}

// NewFactionTech creates a tech tracker with default equal allocation.
func NewFactionTech() *FactionTech {
	ft := &FactionTech{}
	base := 100 / gamedata.TreeCount
	remainder := 100 - base*gamedata.TreeCount
	for i := range ft.Allocation {
		ft.Allocation[i] = base
	}
	for i := 0; i < remainder; i++ {
		ft.Allocation[i]++
	}
	return ft
}

// AddResearch distributes RP across trees and checks for tier advancement.
func (ft *FactionTech) AddResearch(totalRP float64) {
	allocSum := 0
	for _, a := range ft.Allocation {
		allocSum += a
	}
	if allocSum <= 0 {
		return
	}
	for i := range ft.Allocation {
		share := totalRP * float64(ft.Allocation[i]) / float64(allocSum)
		ft.ResearchPts[i] += share
		nextCost := gamedata.TierCost(ft.CurrentTier[i] + 1)
		if ft.CurrentTier[i] < gamedata.MaxTier && ft.ResearchPts[i] >= nextCost {
			ft.ResearchPts[i] -= nextCost
			ft.CurrentTier[i]++
		}
	}
}
