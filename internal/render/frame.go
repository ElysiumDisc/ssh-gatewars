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

	// 2. Overlay ships
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

		for i, ch := range sym {
			px := sx + i
			if px >= 0 && px < w && sy >= 0 && sy < h {
				cells[sy][px] = Cell{Char: string(ch), FG: fg, BG: cells[sy][px].BG, Bold: true}
			}
		}
	}

	// 3. Overlay explosion effects
	for _, ex := range snap.Explosions {
		sx, sy, visible := fb.Viewport.WorldToScreen(ex.X, ex.Y)
		if !visible {
			continue
		}
		fb.renderExplosion(cells, sx, sy, ex.Frame, ex.Faction)
	}

	// 4. Serialize to string
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
