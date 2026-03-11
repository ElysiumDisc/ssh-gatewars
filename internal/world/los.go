package world

import "ssh-gatewars/internal/core"

// HasLOS returns true if there is a clear line of sight between from and to.
// Uses Bresenham's line algorithm, checking each tile for opacity.
func (m *TileMap) HasLOS(from, to core.Pos) bool {
	// Same tile always has LOS
	if from == to {
		return true
	}

	points := BresenhamLine(from, to)
	// Skip the first (source) and last (target) tiles — only check intermediaries
	for i := 1; i < len(points)-1; i++ {
		if m.IsOpaque(points[i]) {
			return false
		}
	}
	return true
}

// LOSPoints returns all tiles along the line of sight, stopping at the first opaque tile.
func (m *TileMap) LOSPoints(from, to core.Pos) []core.Pos {
	points := BresenhamLine(from, to)
	var result []core.Pos
	for i, p := range points {
		result = append(result, p)
		if i > 0 && m.IsOpaque(p) {
			break
		}
	}
	return result
}

// BresenhamLine returns all tile positions along a line from p0 to p1.
func BresenhamLine(p0, p1 core.Pos) []core.Pos {
	dx := abs(p1.X - p0.X)
	dy := -abs(p1.Y - p0.Y)
	sx := 1
	if p0.X > p1.X {
		sx = -1
	}
	sy := 1
	if p0.Y > p1.Y {
		sy = -1
	}
	err := dx + dy

	var points []core.Pos
	x, y := p0.X, p0.Y
	for {
		points = append(points, core.Pos{X: x, Y: y})
		if x == p1.X && y == p1.Y {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
	}
	return points
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
