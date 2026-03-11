package world

import (
	"math/rand"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// GeneratePlanet creates a procedural planet map from a seed and biome.
func GeneratePlanet(seed int64, biome gamedata.Biome, width, height int) *TileMap {
	rng := rand.New(rand.NewSource(seed))
	m := NewTileMap(width, height)

	// BSP room generation
	rooms := generateBSP(rng, 2, 2, width-4, height-4, 5)
	m.Rooms = rooms

	// Carve rooms
	for _, r := range rooms {
		carveRoom(m, r, biome.FloorTile)
	}

	// Connect rooms with corridors
	for i := 1; i < len(rooms); i++ {
		connectRooms(m, rooms[i-1], rooms[i], biome.FloorTile, rng)
	}

	// Place stargate in first room
	gateRoom := rooms[0]
	m.GatePos = gateRoom.Center()
	m.Set(m.GatePos, gamedata.TileStargate)

	// Spawn point adjacent to gate
	m.SpawnPos = core.Pos{X: m.GatePos.X + 1, Y: m.GatePos.Y}
	if !m.IsWalkable(m.SpawnPos) {
		m.SpawnPos = core.Pos{X: m.GatePos.X, Y: m.GatePos.Y + 1}
	}
	if !m.IsWalkable(m.SpawnPos) {
		m.SpawnPos = core.Pos{X: m.GatePos.X - 1, Y: m.GatePos.Y}
	}
	// Ensure spawn is walkable
	m.Set(m.SpawnPos, biome.FloorTile)

	// Scatter accent tiles
	scatterAccents(m, rng, biome, rooms)

	// Place crates in some rooms
	for i := 1; i < len(rooms); i++ {
		if rng.Float64() < 0.3 {
			pos := randomFloorInRoom(m, rng, rooms[i])
			if pos.X != 0 || pos.Y != 0 {
				m.Set(pos, gamedata.TileCrate)
			}
		}
	}

	return m
}

// generateBSP recursively partitions space into rooms.
func generateBSP(rng *rand.Rand, x, y, w, h, depth int) []Room {
	minSize := 5
	if depth <= 0 || w < minSize*2+1 || h < minSize*2+1 {
		// Leaf: create a room within this partition
		rw := minSize + rng.Intn(min(w-minSize, 6)+1)
		rh := minSize + rng.Intn(min(h-minSize, 4)+1)
		if rw > w {
			rw = w
		}
		if rh > h {
			rh = h
		}
		rx := x + rng.Intn(max(w-rw, 1)+1)
		ry := y + rng.Intn(max(h-rh, 1)+1)
		return []Room{{X: rx, Y: ry, W: rw, H: rh}}
	}

	// Split horizontally or vertically
	var rooms []Room
	if rng.Intn(2) == 0 && w >= minSize*2+1 {
		// Vertical split
		split := minSize + rng.Intn(w-minSize*2)
		rooms = append(rooms, generateBSP(rng, x, y, split, h, depth-1)...)
		rooms = append(rooms, generateBSP(rng, x+split+1, y, w-split-1, h, depth-1)...)
	} else if h >= minSize*2+1 {
		// Horizontal split
		split := minSize + rng.Intn(h-minSize*2)
		rooms = append(rooms, generateBSP(rng, x, y, w, split, depth-1)...)
		rooms = append(rooms, generateBSP(rng, x, y+split+1, w, h-split-1, depth-1)...)
	} else {
		rw := minSize + rng.Intn(min(w-minSize, 6)+1)
		rh := minSize + rng.Intn(min(h-minSize, 4)+1)
		if rw > w {
			rw = w
		}
		if rh > h {
			rh = h
		}
		rooms = append(rooms, Room{X: x, Y: y, W: rw, H: rh})
	}
	return rooms
}

func carveRoom(m *TileMap, r Room, floorTile gamedata.TileType) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			m.Set(core.Pos{X: x, Y: y}, floorTile)
		}
	}
}

func connectRooms(m *TileMap, a, b Room, floorTile gamedata.TileType, rng *rand.Rand) {
	ac := a.Center()
	bc := b.Center()

	// L-shaped corridor
	if rng.Intn(2) == 0 {
		carveHCorridor(m, ac.X, bc.X, ac.Y, floorTile)
		carveVCorridor(m, ac.Y, bc.Y, bc.X, floorTile)
	} else {
		carveVCorridor(m, ac.Y, bc.Y, ac.X, floorTile)
		carveHCorridor(m, ac.X, bc.X, bc.Y, floorTile)
	}
}

func carveHCorridor(m *TileMap, x1, x2, y int, floorTile gamedata.TileType) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		m.Set(core.Pos{X: x, Y: y}, floorTile)
	}
}

func carveVCorridor(m *TileMap, y1, y2, x int, floorTile gamedata.TileType) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		m.Set(core.Pos{X: x, Y: y}, floorTile)
	}
}

func scatterAccents(m *TileMap, rng *rand.Rand, biome gamedata.Biome, rooms []Room) {
	if len(biome.AccentTiles) == 0 {
		return
	}
	for _, r := range rooms {
		count := rng.Intn(3)
		for c := 0; c < count; c++ {
			pos := randomFloorInRoom(m, rng, r)
			if pos == m.GatePos || pos == m.SpawnPos {
				continue
			}
			tile := biome.AccentTiles[rng.Intn(len(biome.AccentTiles))]
			// Don't place non-walkable accents on corridors (keep paths clear)
			info := gamedata.Tiles[tile]
			if !info.Walkable {
				// Only place in rooms, not on edges
				inner := Room{X: r.X + 1, Y: r.Y + 1, W: r.W - 2, H: r.H - 2}
				if inner.W > 0 && inner.H > 0 && inner.Contains(pos) {
					m.Set(pos, tile)
				}
			} else {
				m.Set(pos, tile)
			}
		}
	}
}

func randomFloorInRoom(m *TileMap, rng *rand.Rand, r Room) core.Pos {
	for attempts := 0; attempts < 20; attempts++ {
		x := r.X + 1 + rng.Intn(max(r.W-2, 1))
		y := r.Y + 1 + rng.Intn(max(r.H-2, 1))
		p := core.Pos{X: x, Y: y}
		if m.IsWalkable(p) && p != m.GatePos && p != m.SpawnPos {
			return p
		}
	}
	return core.Pos{}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
