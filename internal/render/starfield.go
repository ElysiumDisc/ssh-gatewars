package render

import "math/rand"

// Starfield holds a static background grid of stars.
type Starfield struct {
	Width  int
	Height int
	cells  []byte // flat array, row-major
}

// NewStarfield generates a procedural starfield.
func NewStarfield(width, height int, seed int64) *Starfield {
	rng := rand.New(rand.NewSource(seed))
	sf := &Starfield{
		Width:  width,
		Height: height,
		cells:  make([]byte, width*height),
	}
	for i := range sf.cells {
		r := rng.Float64()
		switch {
		case r < 0.001:
			sf.cells[i] = '+'
		case r < 0.006:
			sf.cells[i] = '*'
		case r < 0.046:
			sf.cells[i] = '.'
		default:
			sf.cells[i] = ' '
		}
	}
	return sf
}

// Get returns the star character at world coordinates.
func (sf *Starfield) Get(x, y int) byte {
	if x < 0 || x >= sf.Width || y < 0 || y >= sf.Height {
		return ' '
	}
	return sf.cells[y*sf.Width+x]
}

// GetColor returns the hex color for the star at world coordinates.
// Returns "" for empty space (no star).
func (sf *Starfield) GetColor(x, y int) string {
	switch sf.Get(x, y) {
	case '.':
		return "#555555"
	case '*':
		return "#AAAAAA"
	case '+':
		return "#FFFFFF"
	default:
		return ""
	}
}
