package simulation

import "time"

const (
	CampaignActive = 0
	CampaignWon    = 1
)

// CampaignState tracks campaign lifecycle.
type CampaignState struct {
	State       int
	Seed        int64
	SystemCount int
	StartedAt   time.Time
	Winner      int
}

func NewCampaign(seed int64, systemCount int) *CampaignState {
	return &CampaignState{
		State:       CampaignActive,
		Seed:        seed,
		SystemCount: systemCount,
		StartedAt:   time.Now(),
		Winner:      -1,
	}
}

func (c *CampaignState) End(winner int) {
	c.State = CampaignWon
	c.Winner = winner
}
