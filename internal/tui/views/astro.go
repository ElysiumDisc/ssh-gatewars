package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/engine"
	"ssh-gatewars/internal/game"
)

// AstroModel holds state for the astrologic star map view.
type AstroModel struct {
	Snapshot *engine.GalaxySnapshot
	CameraX  float64 // pan offset in game-space
	CameraY  float64
	Zoom     float64 // 1.0 = default
	Selected int     // index into Snapshot.Planets (-1 = none)
	Frame    int
}

func NewAstroModel() AstroModel {
	return AstroModel{Zoom: 1.0, Selected: 0}
}

func (a *AstroModel) Reset(snap *engine.GalaxySnapshot) {
	a.Snapshot = snap
	a.CameraX = 0
	a.CameraY = 0
	a.Zoom = 1.0
	a.Selected = 0
	a.Frame = 0
}

func (a *AstroModel) Pan(dx, dy float64) {
	a.CameraX += dx / a.Zoom
	a.CameraY += dy / a.Zoom
}

func (a *AstroModel) ZoomIn() {
	a.Zoom *= 1.3
	if a.Zoom > 3.0 {
		a.Zoom = 3.0
	}
}

func (a *AstroModel) ZoomOut() {
	a.Zoom /= 1.3
	if a.Zoom < 0.3 {
		a.Zoom = 0.3
	}
}

func (a *AstroModel) CycleSelection(delta int) {
	if a.Snapshot == nil || len(a.Snapshot.Planets) == 0 {
		return
	}
	a.Selected += delta
	if a.Selected < 0 {
		a.Selected = len(a.Snapshot.Planets) - 1
	}
	if a.Selected >= len(a.Snapshot.Planets) {
		a.Selected = 0
	}
}

func (a *AstroModel) SelectedPlanetID() int {
	if a.Snapshot == nil || a.Selected < 0 || a.Selected >= len(a.Snapshot.Planets) {
		return -1
	}
	return a.Snapshot.Planets[a.Selected].ID
}

func (a *AstroModel) Tick() {
	a.Frame++
}

// Region labels positioned in galaxy quadrants.
var regionLabels = []struct {
	Name      string
	AngleMin  float64 // radians
	AngleMax  float64
}{
	{"ALPHA QUADRANT", -math.Pi / 4, math.Pi / 4},
	{"PEGASUS SECTOR", math.Pi / 4, 3 * math.Pi / 4},
	{"ORI TERRITORY", 3 * math.Pi / 4, math.Pi},
	{"ORI TERRITORY", -math.Pi, -3 * math.Pi / 4},
	{"ASGARD REACH", -3 * math.Pi / 4, -math.Pi / 4},
}

// Route colors for constellation lines.
var routeColors = []lipgloss.Color{
	lipgloss.Color("#00D9FF"), // cyan
	lipgloss.Color("#FF00FF"), // magenta
	lipgloss.Color("#FF3333"), // red
	lipgloss.Color("#00FF88"), // green
	lipgloss.Color("#FFD700"), // gold
	lipgloss.Color("#8899AA"), // silver
}

// RenderAstro draws the astrologic star map view.
func RenderAstro(a AstroModel, w, h int) string {
	if a.Snapshot == nil {
		return Center(StyleDim.Render("Initializing astrologic sensors..."), w, h)
	}

	innerW := w - 2
	innerH := h - 2

	// Create grid
	fieldW := innerW
	fieldH := innerH - 4 // top bar + bottom bar + spacer
	if fieldH < 10 {
		fieldH = 10
	}

	grid := make([][]rune, fieldH)
	colors := make([][]lipgloss.Style, fieldH)
	for y := range grid {
		grid[y] = make([]rune, fieldW)
		colors[y] = make([]lipgloss.Style, fieldW)
		for x := range grid[y] {
			grid[y][x] = ' '
			colors[y][x] = StyleDim
		}
	}

	centerX := fieldW / 2
	centerY := fieldH / 2

	// Coordinate transform: game-space → screen-space
	scaleX := a.Zoom * float64(fieldW) / 80.0  // 80 game units across screen
	scaleY := a.Zoom * float64(fieldH) / 60.0   // 60 game units tall (terminal chars are taller)

	toScreenX := func(gx float64) int {
		return centerX + int((gx-a.CameraX)*scaleX)
	}
	toScreenY := func(gy float64) int {
		return centerY + int((gy-a.CameraY)*scaleY)
	}
	inBounds := func(x, y int) bool {
		return x >= 0 && x < fieldW && y >= 0 && y < fieldH
	}

	// ── Background star field ─────────────────────────────
	starDim := lipgloss.NewStyle().Foreground(lipgloss.Color("#1A2030"))
	starMid := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A3545"))
	for y := 0; y < fieldH; y++ {
		for x := 0; x < fieldW; x++ {
			hash := (x*7 + y*13 + 31) % 97
			if hash < 3 {
				grid[y][x] = '·'
				if hash == 0 {
					colors[y][x] = starMid
				} else {
					colors[y][x] = starDim
				}
			}
		}
	}

	// ── Constellation lines (network edges) ───────────────
	routeStyleCache := make(map[int]lipgloss.Style)
	for _, link := range a.Snapshot.Links {
		fromIdx := link.FromID
		toIdx := link.ToID
		if fromIdx >= len(a.Snapshot.Planets) || toIdx >= len(a.Snapshot.Planets) {
			continue
		}
		fp := a.Snapshot.Planets[fromIdx]
		tp := a.Snapshot.Planets[toIdx]
		sx0 := toScreenX(fp.Pos.X)
		sy0 := toScreenY(fp.Pos.Y)
		sx1 := toScreenX(tp.Pos.X)
		sy1 := toScreenY(tp.Pos.Y)

		// Get route color
		style, ok := routeStyleCache[link.RouteID]
		if !ok {
			colorIdx := link.RouteID % len(routeColors)
			baseColor := routeColors[colorIdx]
			// Dim constellation lines
			style = lipgloss.NewStyle().Foreground(baseColor).Faint(true)
			routeStyleCache[link.RouteID] = style
		}

		// Brighter if upgraded
		if link.Level > 0 {
			colorIdx := link.RouteID % len(routeColors)
			style = lipgloss.NewStyle().Foreground(routeColors[colorIdx])
			if link.Level >= 2 {
				style = style.Bold(true)
			}
		}

		DrawLineOnGrid(grid, colors, sx0, sy0, sx1, sy1, style)
	}

	// ── Region labels ─────────────────────────────────────
	if a.Zoom < 1.5 { // only show when zoomed out enough
		regionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1E2D3D"))
		for _, reg := range regionLabels {
			midAngle := (reg.AngleMin + reg.AngleMax) / 2
			labelR := 20.0
			lx := toScreenX(math.Cos(midAngle) * labelR)
			ly := toScreenY(math.Sin(midAngle) * labelR)
			DrawStringOnGrid(grid, colors, lx-len(reg.Name)/2, ly, reg.Name, regionStyle)
		}
	}

	// ── Planet dots ───────────────────────────────────────
	for i, p := range a.Snapshot.Planets {
		sx := toScreenX(p.Pos.X)
		sy := toScreenY(p.Pos.Y)
		if !inBounds(sx, sy) {
			continue
		}

		var glyph rune
		var style lipgloss.Style

		switch p.Status {
		case game.PlanetInvaded:
			glyph = '●'
			style = lipgloss.NewStyle().Foreground(ColorDanger)
			if p.Surging {
				if (a.Frame/8)%2 == 0 {
					glyph = '⚠'
				}
			}
		case game.PlanetContested:
			glyph = '◆'
			style = lipgloss.NewStyle().Foreground(ColorGold)
		case game.PlanetFree:
			glyph = '✧'
			style = lipgloss.NewStyle().Foreground(ColorSuccess)
			// Twinkling effect
			if (a.Frame+i*7)%25 < 3 {
				glyph = '·'
				style = lipgloss.NewStyle().Foreground(ColorSuccess).Faint(true)
			}
		}

		grid[sy][sx] = glyph
		colors[sy][sx] = style

		// Selected planet highlight
		if i == a.Selected {
			blink := (a.Frame/10)%2 == 0
			bracketStyle := lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
			if blink {
				if sx-1 >= 0 {
					grid[sy][sx-1] = '['
					colors[sy][sx-1] = bracketStyle
				}
				if sx+1 < fieldW {
					grid[sy][sx+1] = ']'
					colors[sy][sx+1] = bracketStyle
				}
			}
			// Name label
			nameX := sx + 3
			if nameX+len(p.Name) >= fieldW {
				nameX = sx - len(p.Name) - 2
			}
			if nameX >= 0 {
				DrawStringOnGrid(grid, colors, nameX, sy, p.Name, StyleBright)
			}
		}
	}

	// ── Render grid to string ─────────────────────────────
	var sb strings.Builder
	for y := 0; y < fieldH; y++ {
		for x := 0; x < fieldW; x++ {
			sb.WriteString(colors[y][x].Render(string(grid[y][x])))
		}
		if y < fieldH-1 {
			sb.WriteRune('\n')
		}
	}
	field := sb.String()

	// ── Title bar ─────────────────────────────────────────
	title := " " + StyleGold.Render("✦ ASTROLOGIC SURVEY") +
		StyleDim.Render(" — ") +
		StyleCyan.Render("ANCIENT STAR MAP") +
		StyleDim.Render(fmt.Sprintf("  [%d%% FREE]", a.Snapshot.FreePct))

	if a.Snapshot.Cycle > 1 {
		title += StyleDanger.Render(fmt.Sprintf("  CYCLE %d", a.Snapshot.Cycle))
	}

	// ── Detail box for selected planet ────────────────────
	detailLine := ""
	if a.Selected >= 0 && a.Selected < len(a.Snapshot.Planets) {
		p := a.Snapshot.Planets[a.Selected]
		var statusStr string
		switch p.Status {
		case game.PlanetInvaded:
			statusStr = StyleDanger.Render("INVADED")
		case game.PlanetContested:
			statusStr = StyleGold.Render("CONTESTED")
		case game.PlanetFree:
			statusStr = StyleSuccess.Render("FREE")
		}
		detailLine = "  " + StyleBright.Render(p.Name) + " " + statusStr +
			StyleDim.Render(fmt.Sprintf("  Threat:%d", p.InvasionLevel))
		if p.DefenderCount > 0 {
			detailLine += StyleCyan.Render(fmt.Sprintf("  %d defending", p.DefenderCount))
		}
		if p.BountyZPM > 0 && p.Status != game.PlanetFree {
			detailLine += StyleGold.Render(fmt.Sprintf("  ⚡%d", p.BountyZPM))
		}
	}

	// ── Key hints ─────────────────────────────────────────
	keys := "  " +
		FormatKeyHint("←→↑↓", "pan") + "  " +
		FormatKeyHint("+/-", "zoom") + "  " +
		FormatKeyHint("tab", "cycle") + "  " +
		FormatKeyHint("⏎", "deploy") + "  " +
		FormatKeyHint("g", "galaxy") + "  " +
		FormatKeyHint("n", "network") + "  " +
		FormatKeyHint("q", "back")

	content := title + "\n" + field + "\n" + detailLine + "\n" + keys

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2).
		Height(h - 2).
		Render(content)
}
