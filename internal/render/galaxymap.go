package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

// BuildGalaxyMap renders the star system map.
func BuildGalaxyMap(snap simulation.Snapshot, selected, factionID, width, height int, renderer *lipgloss.Renderer) string {
	mapH := height - HUDRows - 1
	if mapH < 5 {
		mapH = 5
	}
	mapW := width
	if mapW < 10 {
		mapW = 10
	}

	// Initialize character grid
	grid := make([][]rune, mapH)
	colors := make([][]string, mapH)
	for y := 0; y < mapH; y++ {
		grid[y] = make([]rune, mapW)
		colors[y] = make([]string, mapW)
		for x := 0; x < mapW; x++ {
			grid[y][x] = ' '
		}
	}

	// Helper to map system position to screen
	sysPos := func(s simulation.SystemSnapshot) (int, int) {
		sx := int(s.MapX * float64(mapW-2)) + 1
		sy := int(s.MapY * float64(mapH-2)) + 1
		if sx < 0 {
			sx = 0
		}
		if sx >= mapW {
			sx = mapW - 1
		}
		if sy < 0 {
			sy = 0
		}
		if sy >= mapH {
			sy = mapH - 1
		}
		return sx, sy
	}

	// Draw gate connections
	dimColor := "#333333"
	for _, gate := range snap.Gates {
		if gate[0] >= len(snap.Systems) || gate[1] >= len(snap.Systems) {
			continue
		}
		x1, y1 := sysPos(snap.Systems[gate[0]])
		x2, y2 := sysPos(snap.Systems[gate[1]])
		drawLine(grid, colors, x1, y1, x2, y2, mapW, mapH, dimColor)
	}

	// Draw systems
	for _, sys := range snap.Systems {
		sx, sy := sysPos(sys)
		ch := '.'
		color := simulation.StarTypeColors[sys.StarType]

		if sys.HasPlanet {
			ch = '*'
		}
		if sys.Special == simulation.SpecialDakara {
			ch = 'D'
			color = "#FFD700"
		} else if sys.Special == simulation.SpecialArtifact {
			ch = 'A'
			color = "#88FFFF"
		}

		// Color by owner if colonized
		if sys.Owner >= 0 && sys.Owner < faction.Count {
			color = faction.Factions[sys.Owner].ColorFG
			ch = '\u2666' // diamond
		}

		grid[sy][sx] = ch
		colors[sy][sx] = color
	}

	// Highlight selected system
	if selected >= 0 && selected < len(snap.Systems) {
		sx, sy := sysPos(snap.Systems[selected])
		// Draw brackets around selected
		if sx > 0 {
			grid[sy][sx-1] = '['
			colors[sy][sx-1] = "#FFFFFF"
		}
		if sx < mapW-1 {
			grid[sy][sx+1] = ']'
			colors[sy][sx+1] = "#FFFFFF"
		}
	}

	// Render grid to string
	var sb strings.Builder
	for y := 0; y < mapH; y++ {
		for x := 0; x < mapW; x++ {
			ch := string(grid[y][x])
			if colors[y][x] != "" {
				style := renderer.NewStyle().Foreground(lipgloss.Color(colors[y][x]))
				sb.WriteString(style.Render(ch))
			} else {
				sb.WriteString(ch)
			}
		}
		if y < mapH-1 {
			sb.WriteString("\n")
		}
	}

	// System info line
	if selected >= 0 && selected < len(snap.Systems) {
		sys := snap.Systems[selected]
		infoStyle := renderer.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
		dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

		info := "\n" + infoStyle.Render("  "+sys.Name)
		info += dimStyle.Render(fmt.Sprintf("  (%s star)", simulation.StarTypeNames[sys.StarType]))

		if col, ok := snap.Colonies[sys.ID]; ok {
			fStyle := renderer.NewStyle().Foreground(lipgloss.Color(faction.Factions[col.Faction].ColorFG))
			info += fStyle.Render(fmt.Sprintf("  [%s]", faction.Factions[col.Faction].ShortName))
			info += dimStyle.Render(fmt.Sprintf("  Pop:%.0f  Fact:%d", col.Population, col.Factories))
		} else if sys.HasPlanet {
			info += dimStyle.Render("  (uncolonized)")
		} else {
			info += dimStyle.Render("  (no planet)")
		}

		if sys.Special == simulation.SpecialDakara {
			info += renderer.NewStyle().Foreground(lipgloss.Color("#FFD700")).Render("  [DAKARA]")
		} else if sys.Special == simulation.SpecialArtifact {
			info += renderer.NewStyle().Foreground(lipgloss.Color("#88FFFF")).Render("  [ARTIFACT]")
		}

		sb.WriteString(info)
	}

	return sb.String()
}

// drawLine draws a dim line on the grid between two points.
func drawLine(grid [][]rune, colors [][]string, x1, y1, x2, y2, w, h int, color string) {
	dx := x2 - x1
	dy := y2 - y1
	steps := max(abs(dx), abs(dy))
	if steps == 0 {
		return
	}

	for i := 1; i < steps; i++ {
		x := x1 + dx*i/steps
		y := y1 + dy*i/steps
		if x >= 0 && x < w && y >= 0 && y < h && grid[y][x] == ' ' {
			grid[y][x] = '\u00B7' // middle dot
			colors[y][x] = color
		}
	}
}

// NavigateSystems returns the index of the nearest system in the given direction.
func NavigateSystems(systems []simulation.SystemSnapshot, current int, direction string) int {
	if len(systems) == 0 || current < 0 || current >= len(systems) {
		return 0
	}

	cur := systems[current]
	bestIdx := current
	bestScore := math.MaxFloat64

	for i, sys := range systems {
		if i == current {
			continue
		}
		dx := sys.MapX - cur.MapX
		dy := sys.MapY - cur.MapY
		dist := math.Sqrt(dx*dx + dy*dy)

		var valid bool
		switch direction {
		case "right":
			valid = dx > 0.01
		case "left":
			valid = dx < -0.01
		case "down":
			valid = dy > 0.01
		case "up":
			valid = dy < -0.01
		}

		if valid && dist < bestScore {
			bestScore = dist
			bestIdx = i
		}
	}

	return bestIdx
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
