package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/chat"
	"ssh-gatewars/internal/engine"
	"ssh-gatewars/internal/game"
)

// DefenseModel holds rendering state for the defense view.
type DefenseModel struct{}

func NewDefenseModel() DefenseModel {
	return DefenseModel{}
}

// RenderDefense draws the radial defense view — the showcase.
func RenderDefense(snap *engine.DefenseSnapshot, msgs []chat.ChatMessage, chatInput string, chatMode, chatVisible bool, frameCount int, playerFP string, w, h int) string {
	if snap == nil {
		return Center(StyleDim.Render("Deploying chair... stand by."), w, h)
	}

	// Layout: defense field on left, chat on right
	chatW := 0
	if chatVisible {
		chatW = 30
		if w > 120 {
			chatW = 38
		}
	}
	fieldW := w - chatW - 4 // border + gap
	fieldH := h - 6         // top bar + bottom bar + borders

	// Render components
	field := renderDefenseField(snap, frameCount, fieldW, fieldH)
	topBar := renderDefenseTopBar(snap, w-4)

	// Find current player's tactic from snapshot
	playerTactic := game.TacticSpread
	for _, c := range snap.Chairs {
		if c.PlayerFP == playerFP {
			playerTactic = c.Tactic
			break
		}
	}

	bottomBar := renderDefenseBottomBar(snap, chatMode, playerTactic, w-4)

	var middle string
	if chatVisible {
		chatPanel := renderChatPanel(msgs, chatInput, chatMode, chatW, fieldH+2)
		fieldStyled := lipgloss.NewStyle().Width(fieldW).Render(field)
		middle = SideBySide(fieldStyled, chatPanel)
	} else {
		middle = field
	}

	content := topBar + "\n" + middle + "\n" + bottomBar

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2).
		Height(h - 2).
		Render(content)
}

func renderDefenseTopBar(snap *engine.DefenseSnapshot, w int) string {
	// Planet name
	planet := StyleCyan.Render(" " + snap.PlanetName)

	// Wave
	wave := StyleDim.Render(" Wave ") + StyleGold.Render(fmt.Sprintf("%d", snap.WaveNum))

	// Hold progress bar
	holdPct := 0.0
	if snap.HoldRequired > 0 {
		holdPct = float64(snap.HoldTicks) / float64(snap.HoldRequired)
	}
	barLen := 20
	bar := ProgressBar(holdPct, barLen, ColorSuccess, ColorDim)

	// Timer countdown
	holdSec := 0
	if snap.HoldRequired > 0 {
		remaining := snap.HoldRequired - snap.HoldTicks
		holdSec = remaining / 10
		if holdSec < 0 {
			holdSec = 0
		}
	}
	timer := StyleMid.Render(fmt.Sprintf("%d:%02d", holdSec/60, holdSec%60))

	// ZPM earned
	zpm := StyleDim.Render(" ⚡") + StyleGold.Render(fmt.Sprintf("%d", snap.ZPMEarned))

	// Status
	status := ""
	if snap.Liberated {
		status = "  " + StyleSuccess.Render("★ LIBERATED ★")
	} else if snap.Failed {
		status = "  " + StyleDanger.Render("✖ DEFENSE FAILED")
	}

	// Surge indicator
	surge := ""
	if snap.Surging {
		surge = "  " + StyleDanger.Render("⚠ SURGE")
	}

	// Bounty
	bounty := ""
	if snap.BountyZPM > 0 {
		bounty = StyleDim.Render(" Bounty:") + StyleGold.Render(fmt.Sprintf("⚡%d", snap.BountyZPM))
	}

	return planet + surge + StyleDim.Render(" │ ") + wave +
		StyleDim.Render(" │ Hold ") + bar + " " + timer +
		StyleDim.Render(" │ ") + zpm + bounty + status
}

func renderDefenseBottomBar(snap *engine.DefenseSnapshot, chatMode bool, tactic game.DroneTactic, w int) string {
	droneCount := len(snap.Drones)
	repCount := len(snap.Replicators)

	// Tactic display
	tacticName := game.TacticNames[tactic]
	tacticStyle := StyleCyan
	if tactic == game.TacticFocus {
		tacticStyle = StyleGold
	} else if tactic == game.TacticPerimeter {
		tacticStyle = StyleSuccess
	}

	left := " " +
		StyleGold.Render("✸") + " " + fmtInt(droneCount, StyleGold) +
		StyleDim.Render("  ") +
		StyleDanger.Render("●") + " " + fmtInt(repCount, StyleDanger) +
		StyleDim.Render("  Kills: ") + fmtInt(snap.TotalKills, StyleBright) +
		StyleDim.Render("  Tactic: ") + tacticStyle.Render(tacticName)

	var keys string
	if chatMode {
		keys = StyleMid.Render("type message") + StyleDim.Render(", ") +
			FormatKeyHint("esc", "cancel") + StyleDim.Render(", ") +
			FormatKeyHint("⏎", "send")
	} else {
		keys = FormatKeyHint("1", "spread") + " " +
			FormatKeyHint("2", "focus") + " " +
			FormatKeyHint("3", "perim") + "  " +
			FormatKeyHint("q", "retreat") + " " +
			FormatKeyHint("c", "chat") + " " +
			FormatKeyHint("tab", "chat±")
	}

	return left + StyleDim.Render("  │  ") + keys
}

func renderDefenseField(snap *engine.DefenseSnapshot, frameCount, w, h int) string {
	// Create character grid
	grid := make([][]rune, h)
	colors := make([][]lipgloss.Style, h)

	for y := range grid {
		grid[y] = make([]rune, w)
		colors[y] = make([]lipgloss.Style, w)
		for x := range grid[y] {
			grid[y][x] = ' '
			colors[y][x] = StyleDim
		}
	}

	centerX := w / 2
	centerY := h / 2

	// Scale: game coords → screen coords
	scaleX := float64(w) / 50.0
	scaleY := float64(h) / 50.0

	toScreenX := func(gx float64) int { return centerX + int(gx*scaleX) }
	toScreenY := func(gy float64) int { return centerY + int(gy*scaleY) }
	inBounds := func(x, y int) bool { return x >= 0 && x < w && y >= 0 && y < h }

	// ── Background star dots (sparse) ──────────────────────────
	// Deterministic based on position, not random
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if (x*7+y*13)%37 == 0 {
				grid[y][x] = '·'
				colors[y][x] = lipgloss.NewStyle().Foreground(ColorDim)
			}
		}
	}

	// ── Crosshair at center ────────────────────────────────────
	veryDim := lipgloss.NewStyle().Foreground(ColorDim)
	if inBounds(centerX, centerY) {
		grid[centerY][centerX] = '╋'
		colors[centerY][centerX] = veryDim
	}

	// ── 3 concentric defense rings ─────────────────────────────
	ringConfigs := []struct {
		radius float64
		style  lipgloss.Style
		char   rune
	}{
		{18.0, lipgloss.NewStyle().Foreground(ColorDim), '·'},
		{13.0, lipgloss.NewStyle().Foreground(ColorCyanDim), '·'},
		{8.0, lipgloss.NewStyle().Foreground(ColorCyan), '·'},
	}
	for _, ring := range ringConfigs {
		steps := int(ring.radius * 8)
		for i := 0; i < steps; i++ {
			angle := float64(i) * 2 * math.Pi / float64(steps)
			px := toScreenX(math.Cos(angle) * ring.radius)
			py := toScreenY(math.Sin(angle) * ring.radius)
			if inBounds(px, py) && grid[py][px] == ' ' || (inBounds(px, py) && grid[py][px] == '·') {
				grid[py][px] = ring.char
				colors[py][px] = ring.style
			}
		}
	}

	// ── Replicators ────────────────────────────────────────────
	for _, r := range snap.Replicators {
		sx := toScreenX(r.Pos.X)
		sy := toScreenY(r.Pos.Y)

		switch r.Type {
		case game.ReplicatorBasic:
			// ● grey
			if inBounds(sx, sy) {
				grid[sy][sx] = '●'
				colors[sy][sx] = RepStyle(int(game.ReplicatorBasic))
			}
		case game.ReplicatorArmored:
			// ■ bright grey
			if inBounds(sx, sy) {
				grid[sy][sx] = '■'
				colors[sy][sx] = RepStyle(int(game.ReplicatorArmored))
			}
		case game.ReplicatorQueen:
			// ◉ pulsing red
			style := RepStyle(int(game.ReplicatorQueen))
			if (frameCount/6)%2 == 0 {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6060")).Bold(true)
			}
			if inBounds(sx, sy) {
				grid[sy][sx] = '◉'
				colors[sy][sx] = style
			}
			// Queen is bigger — draw surrounding marks
			queenMarks := []struct {
				dx, dy int
				ch     rune
			}{
				{-1, 0, '╬'}, {1, 0, '╬'},
				{0, -1, '╬'}, {0, 1, '╬'},
			}
			for _, m := range queenMarks {
				mx, my := sx+m.dx, sy+m.dy
				if inBounds(mx, my) {
					grid[my][mx] = m.ch
					colors[my][mx] = style
				}
			}
		}
	}

	// ── Drones ─────────────────────────────────────────────────
	droneGlyphs := map[game.DroneTier]rune{
		game.DroneTierBase:    '✸',
		game.DroneTierCyan:    '✦',
		game.DroneTierMagenta: '✸',
		game.DroneTierWhite:   '►',
	}
	for _, d := range snap.Drones {
		sx := toScreenX(d.Pos.X)
		sy := toScreenY(d.Pos.Y)
		if inBounds(sx, sy) {
			glyph := droneGlyphs[d.Tier]
			if glyph == 0 {
				glyph = '✸'
			}
			grid[sy][sx] = glyph
			colors[sy][sx] = DroneStyle(int(d.Tier))
		}
	}

	// ── Chairs with shield bar ─────────────────────────────────
	chairStyle := lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
	chairGold := lipgloss.NewStyle().Foreground(ColorGold).Bold(true)
	for _, c := range snap.Chairs {
		sx := toScreenX(c.Pos.X)
		sy := toScreenY(c.Pos.Y)

		// Chair body (3 lines)
		drawStr(grid, colors, sx-2, sy-1, "╔═══╗", chairStyle, w, h)
		// Callsign or symbol inside
		label := "║ ⌂ ║"
		if len(c.Callsign) <= 3 {
			label = fmt.Sprintf("║⌂%s║", c.Callsign)
		}
		drawStr(grid, colors, sx-2, sy, label, chairGold, w, h)
		drawStr(grid, colors, sx-2, sy+1, "╚═══╝", chairStyle, w, h)

		// Shield bar underneath (5 wide)
		shieldBar := ShieldBar(c.ShieldHP, c.MaxShield, 5)
		// Draw shield bar character by character
		barRunes := []rune(shieldBar)
		// Since ShieldBar returns styled text with ANSI, draw raw bar instead
		barW := 5
		barPct := 0.0
		if c.MaxShield > 0 {
			barPct = float64(c.ShieldHP) / float64(c.MaxShield)
		}
		filledCells := int(barPct * float64(barW))
		var barColor lipgloss.Color
		switch {
		case barPct > 0.6:
			barColor = ColorSuccess
		case barPct > 0.3:
			barColor = lipgloss.Color("#FFAA00")
		default:
			barColor = ColorDanger
		}
		barFilledStyle := lipgloss.NewStyle().Foreground(barColor)
		barEmptyStyle := lipgloss.NewStyle().Foreground(ColorDim)
		_ = barRunes // not used — draw directly
		for bx := 0; bx < barW; bx++ {
			px := sx - 2 + bx
			py := sy + 2
			if inBounds(px, py) {
				if bx < filledCells {
					grid[py][px] = '▓'
					colors[py][px] = barFilledStyle
				} else {
					grid[py][px] = '░'
					colors[py][px] = barEmptyStyle
				}
			}
		}
	}

	// ── Render grid to string ──────────────────────────────────
	var sb strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sb.WriteString(colors[y][x].Render(string(grid[y][x])))
		}
		if y < h-1 {
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

// drawStr draws a string onto the grid at the given position.
func drawStr(grid [][]rune, colors [][]lipgloss.Style, x, y int, s string, style lipgloss.Style, w, h int) {
	runes := []rune(s)
	for i, r := range runes {
		px := x + i
		if px >= 0 && px < w && y >= 0 && y < h {
			grid[y][px] = r
			colors[y][px] = style
		}
	}
}
