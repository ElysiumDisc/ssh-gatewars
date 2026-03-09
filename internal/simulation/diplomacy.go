package simulation

import "ssh-gatewars/internal/faction"

const (
	DiplNone     = 0
	DiplWar      = 1
	DiplNAP      = 2
	DiplTrade    = 3
	DiplAlliance = 4
)

// DiplomacyState tracks faction relations. Stub for Layer 5.
type DiplomacyState struct {
	Relations [faction.Count][faction.Count]int
	Proposals []DiplomacyProposal
}

// DiplomacyProposal represents a treaty proposal.
type DiplomacyProposal struct {
	From, To int
	Status   int
	Votes    map[string]bool
}

// NewDiplomacyState creates default (none) relations.
func NewDiplomacyState() *DiplomacyState {
	return &DiplomacyState{}
}
