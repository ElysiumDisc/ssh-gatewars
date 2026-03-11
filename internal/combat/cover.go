package combat

import (
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/world"
)

// CoverValue returns the cover percentage (0-75) a defender gets from
// their position relative to the attacker. Checks the tile between
// attacker and defender (the tile the defender is "behind").
func CoverValue(attacker, defender core.Pos, tm *world.TileMap) int {
	// Direction from defender toward attacker
	dx := sign(attacker.X - defender.X)
	dy := sign(attacker.Y - defender.Y)

	// The cover tile is the tile adjacent to the defender in the direction of the attacker
	coverPos := core.Pos{X: defender.X + dx, Y: defender.Y + dy}

	// If that tile IS the attacker (adjacent melee), no cover
	if coverPos == attacker {
		return 0
	}

	tile := tm.At(coverPos)
	return TileCover(tile)
}

// TileCover returns the cover value for a tile type.
func TileCover(t gamedata.TileType) int {
	switch t {
	case gamedata.TileWall:
		return 75 // full wall corner
	case gamedata.TileCrate, gamedata.TileConsole:
		return 40 // objects provide partial cover
	case gamedata.TileRubble:
		return 25 // low rubble
	case gamedata.TileTree:
		return 30 // tree cover
	default:
		return 0
	}
}

func sign(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}
