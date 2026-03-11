package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/gamedata"
)

// RenderDHD renders the Dial Home Device interface.
// Symbols that are part of the dialed address light up.
func RenderDHD(symbols []int, cursor int, width, height int) string {
	// Styles
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44AAFF")).Bold(true)
	chevronStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA44")).Bold(true)
	litStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")).Bold(true).Background(lipgloss.Color("#332200"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Background(lipgloss.Color("#224488"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555566"))
	ringStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333344"))
	centerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600")).Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	chevronLitStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8800")).Bold(true)
	chevronDimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	// Track which glyphs are "locked" (part of dialed address)
	locked := make(map[int]bool)
	for _, s := range symbols {
		locked[s] = true
	}

	var b strings.Builder
	b.WriteString("\n")

	// Title
	title := "◎ D I A L   H O M E   D E V I C E ◎"
	pad := (width - lipgloss.Width(title)) / 2
	if pad < 0 {
		pad = 0
	}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Chevron display — 7 slots
	chevronLine := "  ╔══╗ "
	for i := 0; i < 7; i++ {
		if i < len(symbols) {
			g := gamedata.Glyphs[symbols[i]]
			chevronLine += chevronLitStyle.Render(fmt.Sprintf("◆%s◆", string(g)))
		} else if i == len(symbols) {
			chevronLine += chevronStyle.Render("◇_◇")
		} else {
			chevronLine += chevronDimStyle.Render("◇·◇")
		}
		if i < 6 {
			chevronLine += chevronDimStyle.Render("─")
		}
	}
	chevronLine += " ╚══╝"
	chevPad := (width - lipgloss.Width(chevronLine)) / 2
	if chevPad < 0 {
		chevPad = 0
	}
	b.WriteString(strings.Repeat(" ", chevPad))
	b.WriteString(chevronLine)
	b.WriteString("\n\n")

	// DHD circular layout
	// 39 glyphs arranged in 3 concentric rings:
	// Ring 1 (outer): 19 glyphs (indices 0-18)
	// Ring 2 (middle): 13 glyphs (indices 19-31)
	// Ring 3 (inner): 7 glyphs (indices 32-38)
	// Center: big activate button

	// Render the circular DHD on a 2D character grid
	dhdW := 58
	dhdH := 23
	grid := make([][]rune, dhdH)
	colors := make([][]string, dhdH)
	for y := range grid {
		grid[y] = make([]rune, dhdW)
		colors[y] = make([]string, dhdW)
		for x := range grid[y] {
			grid[y][x] = ' '
			colors[y][x] = ""
		}
	}

	cx := dhdW / 2 // center x
	cy := dhdH / 2 // center y

	// Draw ring borders (decorative ellipses)
	drawEllipse(grid, colors, cx, cy, 27, 10, ringStyle.Render("·")[0:1], "#333344")
	drawEllipse(grid, colors, cx, cy, 19, 7, ringStyle.Render("·")[0:1], "#333344")
	drawEllipse(grid, colors, cx, cy, 11, 4, ringStyle.Render("·")[0:1], "#333344")

	// Place glyphs on rings
	// Outer ring: 19 glyphs at radius ~27x10 (elliptical)
	placeGlyphsOnRing(grid, colors, cx, cy, 25, 9, 0, 19, symbols, cursor, locked, litStyle, cursorStyle, dimStyle)
	// Middle ring: 13 glyphs at radius ~18x6
	placeGlyphsOnRing(grid, colors, cx, cy, 17, 6, 19, 32, symbols, cursor, locked, litStyle, cursorStyle, dimStyle)
	// Inner ring: 7 glyphs at radius ~10x3.5
	placeGlyphsOnRing(grid, colors, cx, cy, 10, 3, 32, 39, symbols, cursor, locked, litStyle, cursorStyle, dimStyle)

	// Center button
	centerChars := []string{"╔═══╗", "║ ⊛ ║", "╚═══╝"}
	for i, line := range centerChars {
		y := cy - 1 + i
		x := cx - 2
		for j, ch := range line {
			if x+j >= 0 && x+j < dhdW && y >= 0 && y < dhdH {
				grid[y][x+j] = ch
				colors[y][x+j] = "center"
			}
		}
	}

	// Render grid to string
	dhdPad := (width - dhdW) / 2
	if dhdPad < 0 {
		dhdPad = 0
	}
	for y := 0; y < dhdH; y++ {
		b.WriteString(strings.Repeat(" ", dhdPad))
		for x := 0; x < dhdW; x++ {
			ch := string(grid[y][x])
			c := colors[y][x]
			switch c {
			case "lit":
				b.WriteString(litStyle.Render(ch))
			case "cursor":
				b.WriteString(cursorStyle.Render(ch))
			case "dim":
				b.WriteString(dimStyle.Render(ch))
			case "center":
				b.WriteString(centerStyle.Render(ch))
			case "ring":
				b.WriteString(ringStyle.Render(ch))
			default:
				b.WriteString(ch)
			}
		}
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  ←→↑↓ select glyph | Enter lock chevron | Backspace undo | Esc cancel"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  Quick dial: type address (e.g. 26-6-14-31-11-29-0)"))
	b.WriteString("\n")

	// Current glyph info
	if cursor >= 0 && cursor < gamedata.GlyphCount {
		g := gamedata.Glyphs[cursor]
		info := fmt.Sprintf("  Selected: %d → %s", cursor, string(g))
		b.WriteString(chevronStyle.Render(info))
	}

	return b.String()
}

// placeGlyphsOnRing places glyphs along an elliptical ring.
func placeGlyphsOnRing(
	grid [][]rune, colors [][]string,
	cx, cy, rx, ry int,
	startIdx, endIdx int,
	symbols []int, cursor int, locked map[int]bool,
	litStyle, cursorStyle, dimStyle lipgloss.Style,
) {
	count := endIdx - startIdx
	for i := 0; i < count; i++ {
		idx := startIdx + i
		if idx >= gamedata.GlyphCount {
			break
		}

		// Angle: distribute evenly, start from top (-π/2)
		angle := -math.Pi/2 + 2*math.Pi*float64(i)/float64(count)
		x := cx + int(math.Round(float64(rx)*math.Cos(angle)))
		y := cy + int(math.Round(float64(ry)*math.Sin(angle)))

		if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
			continue
		}

		g := gamedata.Glyphs[idx]
		grid[y][x] = g

		if idx == cursor {
			colors[y][x] = "cursor"
		} else if locked[idx] {
			colors[y][x] = "lit"
		} else {
			colors[y][x] = "dim"
		}
	}
}

// drawEllipse draws a decorative ellipse border on the grid.
func drawEllipse(grid [][]rune, colors [][]string, cx, cy, rx, ry int, _ string, _ string) {
	steps := 120
	for i := 0; i < steps; i++ {
		angle := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + int(math.Round(float64(rx)*math.Cos(angle)))
		y := cy + int(math.Round(float64(ry)*math.Sin(angle)))

		if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[0]) {
			if grid[y][x] == ' ' {
				grid[y][x] = '·'
				colors[y][x] = "ring"
			}
		}
	}
}
