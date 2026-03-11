package world

import (
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// GenerateSGC creates the fixed SGC base layout.
//
// Layout (40x20):
//
//	+--------+-----------+----------+
//	| GATE   |           | BRIEFING |
//	| ROOM   |   HALL    |   ROOM   |
//	|   ◎    |           |    ▣     |
//	+---+----+--+-----+--+----+-----+
//	    |       |      |       |
//	+---+---+ +-+------+-+ +--+------+
//	|ARMORY | | MESS HALL | |INFIRM- |
//	|  □ □  | |           | |  ARY   |
//	+-------+ +-----------+ +--------+
func GenerateSGC(width, height int) *TileMap {
	m := NewTileMap(width, height)

	// Gate Room (top-left): 10x7
	gateRoom := Room{X: 1, Y: 1, W: 10, H: 7}
	carveRoom(m, gateRoom, gamedata.TileFloor)

	// Main Hall (top-center): 14x7
	hall := Room{X: 12, Y: 1, W: 14, H: 7}
	carveRoom(m, hall, gamedata.TileFloor)

	// Briefing Room (top-right): 10x7
	briefing := Room{X: 27, Y: 1, W: 11, H: 7}
	carveRoom(m, briefing, gamedata.TileFloor)

	// Armory (bottom-left): 8x5
	armory := Room{X: 1, Y: 10, W: 8, H: 5}
	carveRoom(m, armory, gamedata.TileFloor)

	// Mess Hall (bottom-center): 14x5
	mess := Room{X: 11, Y: 10, W: 14, H: 5}
	carveRoom(m, mess, gamedata.TileFloor)

	// Infirmary (bottom-right): 10x5
	infirmary := Room{X: 27, Y: 10, W: 11, H: 5}
	carveRoom(m, infirmary, gamedata.TileFloor)

	// Corridors connecting rooms
	// Gate room → Hall
	carveHCorridor(m, 11, 12, 4, gamedata.TileFloor)
	// Hall → Briefing
	carveHCorridor(m, 26, 27, 4, gamedata.TileFloor)
	// Hall → lower corridor
	carveVCorridor(m, 8, 10, 18, gamedata.TileFloor)
	// Gate room → Armory
	carveVCorridor(m, 8, 10, 5, gamedata.TileFloor)
	// Lower corridor → Mess
	carveHCorridor(m, 9, 11, 12, gamedata.TileFloor)
	// Lower corridor → Infirmary
	carveHCorridor(m, 25, 27, 12, gamedata.TileFloor)
	// Briefing → Infirmary
	carveVCorridor(m, 8, 10, 32, gamedata.TileFloor)

	// Place stargate
	gatePos := core.Pos{X: 6, Y: 4}
	m.Set(gatePos, gamedata.TileStargate)
	m.GatePos = gatePos
	m.SpawnPos = core.Pos{X: 7, Y: 4}
	m.Set(m.SpawnPos, gamedata.TileFloor) // ensure walkable

	// Place consoles in briefing room
	m.Set(core.Pos{X: 32, Y: 4}, gamedata.TileConsole)

	// Place crates in armory
	m.Set(core.Pos{X: 3, Y: 12}, gamedata.TileCrate)
	m.Set(core.Pos{X: 5, Y: 12}, gamedata.TileCrate)

	// Place doors
	m.Set(core.Pos{X: 11, Y: 4}, gamedata.TileDoor) // Gate→Hall
	m.Set(core.Pos{X: 26, Y: 4}, gamedata.TileDoor) // Hall→Briefing
	m.Set(core.Pos{X: 5, Y: 8}, gamedata.TileDoor)  // Gate→Armory corridor
	m.Set(core.Pos{X: 18, Y: 8}, gamedata.TileDoor)  // Hall→Lower

	m.Rooms = []Room{gateRoom, hall, briefing, armory, mess, infirmary}

	return m
}
