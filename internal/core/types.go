package core

import "math"

// Vec2 is a 2D floating-point vector used for galaxy coordinates.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(o Vec2) Vec2      { return Vec2{v.X + o.X, v.Y + o.Y} }
func (v Vec2) Sub(o Vec2) Vec2      { return Vec2{v.X - o.X, v.Y - o.Y} }
func (v Vec2) Scale(s float64) Vec2 { return Vec2{v.X * s, v.Y * s} }

func (v Vec2) Dist(o Vec2) float64 {
	dx := v.X - o.X
	dy := v.Y - o.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (v Vec2) LenSq() float64 { return v.X*v.X + v.Y*v.Y }

// Rect is an integer rectangle for screen regions.
type Rect struct {
	X, Y, W, H int
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

func (r Rect) Center() (int, int) {
	return r.X + r.W/2, r.Y + r.H/2
}

// Pos is an integer 2D position on a tile map.
type Pos struct {
	X, Y int
}

func (p Pos) Add(d Pos) Pos { return Pos{p.X + d.X, p.Y + d.Y} }

func (p Pos) ManhattanDist(o Pos) int {
	dx := p.X - o.X
	dy := p.Y - o.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// Direction offsets for movement.
var (
	DirUp    = Pos{0, -1}
	DirDown  = Pos{0, 1}
	DirLeft  = Pos{-1, 0}
	DirRight = Pos{1, 0}
)
