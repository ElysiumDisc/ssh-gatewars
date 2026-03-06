package simulation

import "math"

// ShipState represents the lifecycle state of a ship.
type ShipState int

const (
	Alive     ShipState = iota
	Exploding           // playing death animation
	Dead                // ready for removal
)

// Ship is a single entity in the simulation.
type Ship struct {
	ID        uint64
	Faction   int
	X, Y      float64
	VX, VY    float64
	HP        float32
	MaxHP     float32
	Damage    float32
	Speed     float32
	State     ShipState
	SpawnTick uint64
	Boosted   bool

	// Explosion animation
	ExplodeFrame int // 0-3
	ExplodeTicks int // ticks remaining on current frame

	// Trail: last 3 positions for fading trail effect
	Trail     [3]Vec2
	TrailLen  int // how many trail entries are valid (0-3)
}

// Vec2 is a simple 2D vector.
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(o Vec2) Vec2     { return Vec2{v.X + o.X, v.Y + o.Y} }
func (v Vec2) Sub(o Vec2) Vec2     { return Vec2{v.X - o.X, v.Y - o.Y} }
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

// Dist returns the distance between two ships.
func Dist(a, b *Ship) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}
