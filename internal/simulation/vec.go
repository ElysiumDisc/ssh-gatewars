package simulation

import "math"

// Vec2 is a simple 2D vector.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(o Vec2) Vec2      { return Vec2{v.X + o.X, v.Y + o.Y} }
func (v Vec2) Sub(o Vec2) Vec2      { return Vec2{v.X - o.X, v.Y - o.Y} }
func (v Vec2) Scale(s float64) Vec2 { return Vec2{v.X * s, v.Y * s} }

func (v Vec2) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) Normalize() Vec2 {
	l := v.Len()
	if l < 0.0001 {
		return Vec2{}
	}
	return Vec2{v.X / l, v.Y / l}
}

func (v Vec2) DistTo(o Vec2) float64 {
	dx := v.X - o.X
	dy := v.Y - o.Y
	return math.Sqrt(dx*dx + dy*dy)
}
