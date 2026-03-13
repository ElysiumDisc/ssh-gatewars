package views

import "github.com/charmbracelet/lipgloss"

// Pos represents a screen coordinate.
type Pos struct {
	X, Y int
}

// BresenhamLine returns screen coordinates for a line between two points.
func BresenhamLine(x0, y0, x1, y1 int) []Pos {
	var points []Pos

	dx := x1 - x0
	dy := y1 - y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	for {
		points = append(points, Pos{x0, y0})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
	return points
}

// LineChar returns an appropriate line-drawing character based on direction.
func LineChar(dx, dy int) rune {
	switch {
	case dy == 0:
		return '─'
	case dx == 0:
		return '│'
	case (dx > 0 && dy > 0) || (dx < 0 && dy < 0):
		return '╲'
	default:
		return '╱'
	}
}

// DrawLineOnGrid draws a line between two points on a rune/style grid.
func DrawLineOnGrid(grid [][]rune, colors [][]lipgloss.Style, x0, y0, x1, y1 int, style lipgloss.Style) {
	h := len(grid)
	if h == 0 {
		return
	}
	w := len(grid[0])

	points := BresenhamLine(x0, y0, x1, y1)
	dx := x1 - x0
	dy := y1 - y0
	ch := LineChar(dx, dy)

	for _, p := range points {
		if p.X >= 0 && p.X < w && p.Y >= 0 && p.Y < h {
			if grid[p.Y][p.X] == ' ' || grid[p.Y][p.X] == '·' {
				grid[p.Y][p.X] = ch
				colors[p.Y][p.X] = style
			}
		}
	}
}

// DrawStringOnGrid draws a string at a position, respecting bounds.
func DrawStringOnGrid(grid [][]rune, colors [][]lipgloss.Style, x, y int, s string, style lipgloss.Style) {
	h := len(grid)
	if h == 0 {
		return
	}
	w := len(grid[0])
	if y < 0 || y >= h {
		return
	}
	for i, r := range []rune(s) {
		px := x + i
		if px >= 0 && px < w {
			grid[y][px] = r
			colors[y][px] = style
		}
	}
}
