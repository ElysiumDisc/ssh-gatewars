package simulation

import (
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// ActionType identifies a player action.
type ActionType int

const (
	ActionMove      ActionType = iota // move in a direction (bump-attack if enemy)
	ActionInteract                     // interact with adjacent tile (loot crate, console)
	ActionDialGate                     // dial a gate address
	ActionPickup                       // pick up ground item
	ActionUseItem                      // use consumable from inventory
	ActionEquip                        // equip an inventory item
	ActionUnequip                      // unequip to inventory
	ActionFire                         // fire ranged weapon at target position
	ActionReload                       // reload current weapon
)

// PlayerAction is a command from a player session.
type PlayerAction struct {
	Type      ActionType
	PlayerKey string // SSH fingerprint

	// ActionMove
	Dir core.Pos

	// ActionDialGate
	Address gamedata.GateAddress

	// ActionPickup
	TargetPos core.Pos

	// ActionUseItem / ActionEquip / ActionUnequip
	ItemIndex int

	// ActionFire
	FireTarget core.Pos
}
