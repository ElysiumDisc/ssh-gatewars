package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

// Cell represents one terminal character with styling.
type Cell struct {
	Char  string
	FG    string // color hex
	BG    string // color hex, empty = default
	Bold  bool
}

// FrameBuilder composites starfield, territory, ships, and effects.
type FrameBuilder struct {
	Viewport  *Viewport
	Starfield *Starfield
	Renderer  *lipgloss.Renderer
}

// Build creates a rendered frame string from a simulation snapshot.
func (fb *FrameBuilder) Build(snap simulation.Snapshot) string {
	w, h := fb.Viewport.Width, fb.Viewport.Height
	if w < 1 || h < 1 {
		return ""
	}

	// 1. Build cell buffer
	cells := make([][]Cell, h)
	for y := 0; y < h; y++ {
		cells[y] = make([]Cell, w)
		for x := 0; x < w; x++ {
			wx, wy := fb.Viewport.ScreenToWorld(x, y)
			ch := fb.Starfield.Get(wx, wy)

			cell := Cell{Char: string(ch), FG: "#555555"}
			if ch == '*' {
				cell.FG = "#AAAAAA"
			} else if ch == '+' {
				cell.FG = "#FFFFFF"
				cell.Bold = true
			} else if ch == ' ' {
				cell.FG = "#333333"
			}

			// Territory tinting
			if snap.Territory != nil {
				owner := snap.Territory.ZoneAt(wx, wy)
				if owner >= 0 && owner < faction.Count {
					cell.BG = faction.Factions[owner].ColorBG
				}
			}

			cells[y][x] = cell
		}
	}

	// 2. Overlay ship trails (dim fading dots behind ships)
	trailColors := []string{"#555555", "#444444", "#333333"}
	for _, s := range snap.Ships {
		if s.State != simulation.Alive || s.TrailLen == 0 {
			continue
		}
		for ti := 0; ti < s.TrailLen; ti++ {
			tx, ty, vis := fb.Viewport.WorldToScreen(s.Trail[ti].X, s.Trail[ti].Y)
			if vis && tx >= 0 && tx < w && ty >= 0 && ty < h {
				cells[ty][tx] = Cell{Char: "·", FG: trailColors[ti], BG: cells[ty][tx].BG}
			}
		}
	}

	// 3. Overlay ships
	for _, s := range snap.Ships {
		if s.State == simulation.Dead {
			continue
		}

		sx, sy, visible := fb.Viewport.WorldToScreen(s.X, s.Y)
		if !visible {
			continue
		}

		if s.State == simulation.Exploding {
			fb.renderExplosion(cells, sx, sy, s.ExplodeFrame, s.Faction)
			continue
		}

		sym := faction.Symbol(s.Faction, s.VX, s.VY)
		fg := faction.Factions[s.Faction].ColorFG

		// Jaffa Kree! boost: highlight with bright yellow
		if s.Boosted {
			fg = "#FFFF00"
		}

		for i, ch := range sym {
			px := sx + i
			if px >= 0 && px < w && sy >= 0 && sy < h {
				bg := cells[sy][px].BG
				// Kassa Rush: alternating purple/red background tint for Lucian ships
				if s.Faction == faction.Lucian && snap.PowerStatuses[faction.Lucian].State == simulation.PowerActive {
					if (sx+sy)%2 == 0 {
						bg = "#3A0040"
					} else {
						bg = "#400010"
					}
				}
				cells[sy][px] = Cell{Char: string(ch), FG: fg, BG: bg, Bold: true}
			}
		}
	}

	// 4. Overlay explosion effects
	for _, ex := range snap.Explosions {
		sx, sy, visible := fb.Viewport.WorldToScreen(ex.X, ex.Y)
		if !visible {
			continue
		}
		fb.renderExplosion(cells, sx, sy, ex.Frame, ex.Faction)
	}

	// 5. Overlay beam effects (Asgard Ion Cannon, etc.)
	for _, beam := range snap.Beams {
		fb.renderBeam(cells, beam)
	}

	// 6. Serialize to string
	return fb.serialize(cells)
}

func (fb *FrameBuilder) renderExplosion(cells [][]Cell, sx, sy, frame, factionID int) {
	w, h := fb.Viewport.Width, fb.Viewport.Height
	colors := []string{"#FF4444", "#FF8844", "#FFAA44", "#666666"}
	color := colors[0]
	if frame < len(colors) {
		color = colors[frame]
	}

	type offset struct {
		dx, dy int
		ch     string
	}

	var pattern []offset
	switch frame {
	case 0:
		pattern = []offset{{0, 0, "*"}}
	case 1:
		pattern = []offset{
			{0, -1, "|"}, {-1, 0, "-"}, {0, 0, "X"}, {1, 0, "-"}, {0, 1, "|"},
		}
	case 2:
		pattern = []offset{
			{-1, 0, ":"}, {0, 0, "."}, {1, 0, ":"}, {0, 1, "."},
		}
	case 3:
		pattern = []offset{
			{-1, 0, "."}, {1, 0, "."},
		}
	}

	for _, p := range pattern {
		px, py := sx+p.dx, sy+p.dy
		if px >= 0 && px < w && py >= 0 && py < h {
			cells[py][px] = Cell{Char: p.ch, FG: color, Bold: frame < 2}
		}
	}
}

func (fb *FrameBuilder) renderBeam(cells [][]Cell, beam simulation.BeamEffect) {
	w, h := fb.Viewport.Width, fb.Viewport.Height

	// Map beam endpoints to screen coords
	sx1, sy1, _ := fb.Viewport.WorldToScreen(beam.X1, beam.Y1)
	sx2, sy2, _ := fb.Viewport.WorldToScreen(beam.X2, beam.Y2)

	// Bresenham line to draw beam across screen
	dx := sx2 - sx1
	dy := sy2 - sy1
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		return
	}

	beamColor := "#00FFFF" // cyan for Asgard
	if beam.Faction >= 0 && beam.Faction < faction.Count {
		beamColor = faction.Factions[beam.Faction].ColorFG
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		px := sx1 + int(t*float64(dx))
		py := sy1 + int(t*float64(dy))
		if px >= 0 && px < w && py >= 0 && py < h {
			ch := "═"
			if abs(dy) > abs(dx) {
				ch = "║"
			}
			cells[py][px] = Cell{Char: ch, FG: beamColor, Bold: true}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (fb *FrameBuilder) serialize(cells [][]Cell) string {
	var sb strings.Builder
	sb.Grow(len(cells) * len(cells[0]) * 4) // rough estimate

	for y, row := range cells {
		if y > 0 {
			sb.WriteByte('\n')
		}
		// Group consecutive cells with the same style for efficiency
		i := 0
		for i < len(row) {
			j := i + 1
			for j < len(row) && row[j].FG == row[i].FG && row[j].BG == row[i].BG && row[j].Bold == row[i].Bold {
				j++
			}

			// Build the text for this run
			var text strings.Builder
			for k := i; k < j; k++ {
				text.WriteString(row[k].Char)
			}

			style := fb.Renderer.NewStyle().Foreground(lipgloss.Color(row[i].FG))
			if row[i].BG != "" {
				style = style.Background(lipgloss.Color(row[i].BG))
			}
			if row[i].Bold {
				style = style.Bold(true)
			}

			sb.WriteString(style.Render(text.String()))
			i = j
		}
	}

	return sb.String()
}
