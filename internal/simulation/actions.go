package simulation

import "sync"

type ActionType int

const (
	ActionSetSliders ActionType = iota
	ActionSetTechAlloc
	ActionAddBuildOrder
	ActionClearBuildQueue
	ActionDesignShip
	ActionScrapDesign
	ActionMoveFleet
	ActionColonize
	ActionDiplomacyPropose
	ActionDiplomacyVote
)

type PlayerAction struct {
	Type      ActionType
	PlayerKey string
	Faction   int
	SystemID  int

	Sliders   [5]int
	TechAlloc [6]int

	BuildType int
	BuildName string

	FleetID      uint64
	TargetSystem int

	DiplTarget int
	DiplStatus int
}

type ActionQueue struct {
	mu      sync.Mutex
	actions []PlayerAction
}

func NewActionQueue() *ActionQueue {
	return &ActionQueue{}
}

func (q *ActionQueue) Enqueue(a PlayerAction) {
	q.mu.Lock()
	q.actions = append(q.actions, a)
	q.mu.Unlock()
}

func (q *ActionQueue) Drain() []PlayerAction {
	q.mu.Lock()
	a := q.actions
	q.actions = nil
	q.mu.Unlock()
	return a
}
