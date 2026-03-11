package gamedata

// TileType identifies the kind of terrain at a map position.
type TileType uint8

const (
	TileFloor    TileType = iota // '.'
	TileWall                     // '#'
	TileDoor                     // '+'
	TileStargate                 // '◎'
	TileWater                    // '~'
	TileLava                     // '~' red
	TileSand                     // '.'
	TileIce                      // '.'
	TileTree                     // '♣'
	TileRubble                   // ','
	TileCrate                    // '□'
	TileConsole                  // '▣'
	TileStairsDown               // '>'
	TileHalfWall                 // '▄' — low wall, provides 50% cover
	TilePillar                   // '║' — pillar, opaque
	TileAltar                    // '╬' — altar / sarcophagus
	TileInscription              // '∆' — Ancient inscription (interactable)
	TileVent                     // '░' — steam vent (hazard)
	TileGlyph                    // '∆' — Ancient glyph
)

// TileInfo describes the properties of a tile type.
type TileInfo struct {
	Glyph    rune
	Name     string
	Walkable bool
	Opaque   bool   // blocks line of sight
	Color    string // lipgloss color
	Cover    int    // cover percentage (0-75) against ranged attacks
}

// Tiles maps each TileType to its rendering and gameplay info.
var Tiles = map[TileType]TileInfo{
	TileFloor:       {'.', "floor", true, false, "#666666", 0},
	TileWall:        {'#', "wall", false, true, "#888888", 75},
	TileDoor:        {'+', "door", true, false, "#AA8844", 0},
	TileStargate:    {'◎', "stargate", true, false, "#44AAFF", 0},
	TileWater:       {'~', "water", false, false, "#4488FF", 0},
	TileLava:        {'~', "lava", false, false, "#FF4400", 0},
	TileSand:        {'.', "sand", true, false, "#CCAA44", 0},
	TileIce:         {'.', "ice", true, false, "#AADDFF", 0},
	TileTree:        {'♣', "tree", false, true, "#22AA22", 30},
	TileRubble:      {',', "rubble", true, false, "#777755", 25},
	TileCrate:       {'□', "crate", false, false, "#AA7744", 40},
	TileConsole:     {'▣', "console", false, false, "#44FFCC", 40},
	TileStairsDown:  {'>', "stairs", true, false, "#FFFFFF", 0},
	TileHalfWall:    {'▄', "low wall", false, false, "#777777", 50},
	TilePillar:      {'║', "pillar", false, true, "#999999", 75},
	TileAltar:       {'╬', "altar", false, false, "#DDAA44", 60},
	TileInscription: {'∆', "inscription", true, false, "#7799CC", 0},
	TileVent:        {'░', "vent", true, false, "#CC4400", 0},
	TileGlyph:       {'∆', "glyph", true, false, "#88AADD", 0},
}
