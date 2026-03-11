package combat

import (
	"ssh-gatewars/internal/core"
)

// Projectile represents a projectile in flight on the map.
type Projectile struct {
	ID         uint32
	OwnerKey   string    // SSH key of the player who fired, or enemy ID string
	IsPlayer   bool      // true if fired by a player
	Origin     core.Pos  // where it was fired from
	Target     core.Pos  // where it's aimed at
	Pos        core.Pos  // current position
	Damage     int       // damage on hit
	Glyph      rune      // rendering character
	Color      string    // lipgloss color
	Speed      int       // tiles per tick
	TicksLeft  int       // ticks before it expires
	Path       []core.Pos // precomputed path from origin to target
	PathIndex  int       // current index along path
}

// NewProjectile creates a projectile along a Bresenham path.
func NewProjectile(id uint32, ownerKey string, isPlayer bool, from, to core.Pos, damage int, glyph rune, color string, speed int) *Projectile {
	// Import would be circular; we use a simple line here
	path := bresenhamLine(from, to)
	ttl := len(path) + 2 // give it time to reach + a little extra

	return &Projectile{
		ID:        id,
		OwnerKey:  ownerKey,
		IsPlayer:  isPlayer,
		Origin:    from,
		Target:    to,
		Pos:       from,
		Damage:    damage,
		Glyph:     glyph,
		Color:     color,
		Speed:     speed,
		TicksLeft: ttl,
		Path:      path,
		PathIndex: 0,
	}
}

// Advance moves the projectile along its path. Returns true if it has expired.
func (p *Projectile) Advance() bool {
	p.TicksLeft--
	if p.TicksLeft <= 0 {
		return true
	}

	// Move Speed tiles along the path
	for i := 0; i < p.Speed; i++ {
		p.PathIndex++
		if p.PathIndex >= len(p.Path) {
			return true // reached end of path
		}
		p.Pos = p.Path[p.PathIndex]
	}
	return false
}

// bresenhamLine is a local copy to avoid circular import with world package.
func bresenhamLine(p0, p1 core.Pos) []core.Pos {
	dx := intAbs(p1.X - p0.X)
	dy := -intAbs(p1.Y - p0.Y)
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

func intAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
