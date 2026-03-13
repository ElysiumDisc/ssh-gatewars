package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/engine"
	"ssh-gatewars/internal/game"
)

// NetworkMode defines the interaction mode in the network view.
type NetworkMode int

const (
	NetworkBrowse   NetworkMode = iota // browsing stations
	NetworkUpgrade                     // selecting a link to upgrade
	NetworkTransfer                    // selecting a resource to send
)

// NetworkModel holds state for the Stargate network tube map view.
type NetworkModel struct {
	Snapshot       *engine.GalaxySnapshot
	Selected       int // index into Snapshot.Planets
	Mode           NetworkMode
	UpgradeLink    int // index into Snapshot.Links being considered (-1 = none)
	TransferTarget int // planet ID for transfer (-1 = none)

	// Layout cache (computed on Reset)
	stationPos map[int]Pos // planet ID → screen position in the tube map space
	mapW, mapH int         // total tube map dimensions

	// Scroll
	ScrollX, ScrollY int

	// Status message
	StatusMsg   string
	StatusStyle lipgloss.Style
	StatusTTL   int

	Frame int
}

func NewNetworkModel() NetworkModel {
	return NetworkModel{Selected: 0, UpgradeLink: -1, TransferTarget: -1}
}

func (n *NetworkModel) Reset(snap *engine.GalaxySnapshot) {
	n.Snapshot = snap
	n.Selected = 0
	n.Mode = NetworkBrowse
	n.UpgradeLink = -1
	n.TransferTarget = -1
	n.ScrollX = 0
	n.ScrollY = 0
	n.StatusMsg = ""
	n.StatusTTL = 0
	n.Frame = 0
	n.computeLayout()
}

func (n *NetworkModel) MoveSelection(delta int) {
	if n.Snapshot == nil || len(n.Snapshot.Planets) == 0 {
		return
	}
	n.Selected += delta
	if n.Selected < 0 {
		n.Selected = 0
	}
	max := len(n.Snapshot.Planets) - 1
	if n.Selected > max {
		n.Selected = max
	}
}

func (n *NetworkModel) SelectedPlanetID() int {
	if n.Snapshot == nil || n.Selected < 0 || n.Selected >= len(n.Snapshot.Planets) {
		return -1
	}
	return n.Snapshot.Planets[n.Selected].ID
}

func (n *NetworkModel) SetStatus(msg string, style lipgloss.Style) {
	n.StatusMsg = msg
	n.StatusStyle = style
	n.StatusTTL = 45 // ~3 seconds at 15fps
}

func (n *NetworkModel) Tick() {
	n.Frame++
	if n.StatusTTL > 0 {
		n.StatusTTL--
		if n.StatusTTL == 0 {
			n.StatusMsg = ""
		}
	}
}

// CycleUpgradeLink cycles through links connected to the selected planet.
func (n *NetworkModel) CycleUpgradeLink(delta int) {
	if n.Snapshot == nil {
		return
	}
	pid := n.SelectedPlanetID()
	if pid < 0 {
		return
	}
	// Find all links connected to this planet
	var connLinks []int
	for i, l := range n.Snapshot.Links {
		if l.FromID == pid || l.ToID == pid {
			connLinks = append(connLinks, i)
		}
	}
	if len(connLinks) == 0 {
		return
	}
	// Find current position
	cur := 0
	for i, li := range connLinks {
		if li == n.UpgradeLink {
			cur = i
			break
		}
	}
	cur += delta
	if cur < 0 {
		cur = len(connLinks) - 1
	}
	if cur >= len(connLinks) {
		cur = 0
	}
	n.UpgradeLink = connLinks[cur]
}

// computeLayout builds the subway map layout from route data.
func (n *NetworkModel) computeLayout() {
	n.stationPos = make(map[int]Pos)
	if n.Snapshot == nil {
		return
	}

	routes := n.Snapshot.Routes
	if len(routes) == 0 {
		// Fallback: position planets in a grid
		cols := 10
		for i, p := range n.Snapshot.Planets {
			n.stationPos[p.ID] = Pos{
				X: (i % cols) * 8 + 4,
				Y: (i / cols) * 4 + 2,
			}
		}
		n.mapW = cols*8 + 8
		n.mapH = (len(n.Snapshot.Planets)/cols+1)*4 + 4
		return
	}

	// Layout: each route is a horizontal band
	bandSpacing := 6
	stationSpacing := 10
	maxW := 0

	for ri, route := range routes {
		baseY := ri*bandSpacing + 3
		for si, pid := range route.Planets {
			x := si*stationSpacing + 6
			pos := Pos{X: x, Y: baseY}
			// If planet already placed (interchange), keep first position
			if _, exists := n.stationPos[pid]; !exists {
				n.stationPos[pid] = pos
			}
			if x > maxW {
				maxW = x
			}
		}
	}

	// Place any unplaced planets (not on any route) at the bottom
	bottomY := len(routes)*bandSpacing + 5
	nextX := 6
	for _, p := range n.Snapshot.Planets {
		if _, exists := n.stationPos[p.ID]; !exists {
			n.stationPos[p.ID] = Pos{X: nextX, Y: bottomY}
			nextX += stationSpacing
			if nextX > maxW {
				maxW = nextX
			}
		}
	}

	n.mapW = maxW + 12
	n.mapH = bottomY + 4
}

// Tube map route display colors.
var tubeColors = []lipgloss.Color{
	lipgloss.Color("#00D9FF"), // cyan
	lipgloss.Color("#FF00FF"), // magenta
	lipgloss.Color("#FF3333"), // red
	lipgloss.Color("#00FF88"), // green
	lipgloss.Color("#FFD700"), // gold
	lipgloss.Color("#8899AA"), // silver
}

// RenderNetwork draws the Stargate network tube map view.
func RenderNetwork(n NetworkModel, w, h int) string {
	if n.Snapshot == nil {
		return Center(StyleDim.Render("Dialing Stargate network..."), w, h)
	}

	innerW := w - 2
	innerH := h - 2

	// ── Title bar ─────────────────────────────────────────
	title := " " + StyleGold.Render("⌬ STARGATE NETWORK") +
		StyleDim.Render(" — ") +
		StyleCyan.Render("GATE TOPOLOGY")

	if n.StatusMsg != "" {
		title += "  " + n.StatusStyle.Render(n.StatusMsg)
	}

	// ── Mode indicator ────────────────────────────────────
	modeStr := ""
	switch n.Mode {
	case NetworkUpgrade:
		modeStr = StyleGold.Render("  [UPGRADE MODE]") +
			StyleDim.Render(" ←→ select link, ⏎ upgrade, esc cancel")
	case NetworkTransfer:
		modeStr = StyleSuccess.Render("  [TRANSFER MODE]") +
			StyleDim.Render(" [1]Shield [2]Drones [3]ZPM, esc cancel")
	}
	title += modeStr

	// ── Route legend ──────────────────────────────────────
	showLegend := innerW > 100
	legendW := 0
	if showLegend {
		legendW = 22
	}

	fieldW := innerW - legendW
	fieldH := innerH - 5 // title + detail + keys + borders

	// ── Build the tube map grid ───────────────────────────
	gridW := fieldW
	gridH := fieldH
	if gridW < 20 {
		gridW = 20
	}
	if gridH < 10 {
		gridH = 10
	}

	grid := make([][]rune, gridH)
	colors := make([][]lipgloss.Style, gridH)
	for y := range grid {
		grid[y] = make([]rune, gridW)
		colors[y] = make([]lipgloss.Style, gridW)
		for x := range grid[y] {
			grid[y][x] = ' '
			colors[y][x] = StyleDim
		}
	}

	// ── Auto-scroll to keep selected visible ──────────────
	if selID := n.SelectedPlanetID(); selID >= 0 {
		if pos, ok := n.stationPos[selID]; ok {
			// Ensure selected station is visible
			if pos.X-n.ScrollX < 4 {
				n.ScrollX = pos.X - 4
			}
			if pos.X-n.ScrollX > gridW-8 {
				n.ScrollX = pos.X - gridW + 8
			}
			if pos.Y-n.ScrollY < 2 {
				n.ScrollY = pos.Y - 2
			}
			if pos.Y-n.ScrollY > gridH-3 {
				n.ScrollY = pos.Y - gridH + 3
			}
		}
	}
	if n.ScrollX < 0 {
		n.ScrollX = 0
	}
	if n.ScrollY < 0 {
		n.ScrollY = 0
	}

	// ── Draw route lines ──────────────────────────────────
	for ri, route := range n.Snapshot.Routes {
		colorIdx := ri % len(tubeColors)
		routeStyle := lipgloss.NewStyle().Foreground(tubeColors[colorIdx])

		for k := 0; k < len(route.Planets)-1; k++ {
			fromID := route.Planets[k]
			toID := route.Planets[k+1]
			fromPos, ok1 := n.stationPos[fromID]
			toPos, ok2 := n.stationPos[toID]
			if !ok1 || !ok2 {
				continue
			}

			// Screen coords with scroll offset
			fx := fromPos.X - n.ScrollX
			fy := fromPos.Y - n.ScrollY
			tx := toPos.X - n.ScrollX
			ty := toPos.Y - n.ScrollY

			// Draw with horizontal-first routing (subway style)
			if fy == ty {
				// Same level — straight horizontal
				DrawLineOnGrid(grid, colors, fx, fy, tx, ty, routeStyle)
			} else {
				// Route: horizontal to midpoint, then vertical, then horizontal
				midX := (fx + tx) / 2
				DrawLineOnGrid(grid, colors, fx, fy, midX, fy, routeStyle)
				DrawLineOnGrid(grid, colors, midX, fy, midX, ty, routeStyle)
				DrawLineOnGrid(grid, colors, midX, ty, tx, ty, routeStyle)
			}
		}
	}

	// ── Highlight upgrade link ────────────────────────────
	if n.Mode == NetworkUpgrade && n.UpgradeLink >= 0 && n.UpgradeLink < len(n.Snapshot.Links) {
		link := n.Snapshot.Links[n.UpgradeLink]
		fromPos, ok1 := n.stationPos[link.FromID]
		toPos, ok2 := n.stationPos[link.ToID]
		if ok1 && ok2 {
			upgradeStyle := lipgloss.NewStyle().Foreground(ColorGold).Bold(true)
			fx := fromPos.X - n.ScrollX
			fy := fromPos.Y - n.ScrollY
			tx := toPos.X - n.ScrollX
			ty := toPos.Y - n.ScrollY
			DrawLineOnGrid(grid, colors, fx, fy, tx, ty, upgradeStyle)
		}
	}

	// ── Draw station dots ─────────────────────────────────
	for i, p := range n.Snapshot.Planets {
		pos, ok := n.stationPos[p.ID]
		if !ok {
			continue
		}
		sx := pos.X - n.ScrollX
		sy := pos.Y - n.ScrollY
		if sx < 0 || sx >= gridW || sy < 0 || sy >= gridH {
			continue
		}

		var glyph rune
		var style lipgloss.Style
		switch p.Status {
		case game.PlanetInvaded:
			glyph = '●'
			style = lipgloss.NewStyle().Foreground(ColorDanger)
		case game.PlanetContested:
			glyph = '◆'
			style = lipgloss.NewStyle().Foreground(ColorGold)
		case game.PlanetFree:
			glyph = '○'
			style = lipgloss.NewStyle().Foreground(ColorSuccess)
		}

		// Selected highlight
		if i == n.Selected {
			glyph = '◉'
			style = lipgloss.NewStyle().Foreground(ColorBright).Bold(true)
			// Brackets
			if sx-1 >= 0 {
				grid[sy][sx-1] = '▸'
				colors[sy][sx-1] = lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)
			}
		}

		grid[sy][sx] = glyph
		colors[sy][sx] = style

		// Station name label (to the right)
		labelX := sx + 2
		if i == n.Selected {
			labelX = sx + 2
		}
		if labelX+len(p.Name) < gridW && sy >= 0 && sy < gridH {
			labelStyle := StyleDim
			if i == n.Selected {
				labelStyle = StyleBright
			}
			DrawStringOnGrid(grid, colors, labelX, sy, p.Name, labelStyle)
		}
	}

	// ── Render grid ───────────────────────────────────────
	var sb strings.Builder
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			sb.WriteString(colors[y][x].Render(string(grid[y][x])))
		}
		if y < gridH-1 {
			sb.WriteRune('\n')
		}
	}
	mapStr := sb.String()

	// ── Route legend ──────────────────────────────────────
	if showLegend {
		var legendLines []string
		legendLines = append(legendLines, "")
		legendLines = append(legendLines, " "+StyleDim.Render("ROUTES"))
		legendLines = append(legendLines, " "+StyleCyanDim.Render(strings.Repeat("─", legendW-4)))
		for ri, route := range n.Snapshot.Routes {
			colorIdx := ri % len(tubeColors)
			swatch := lipgloss.NewStyle().Foreground(tubeColors[colorIdx]).Render("━━")
			name := truncate(route.Name, legendW-6)
			legendLines = append(legendLines, " "+swatch+" "+StyleMid.Render(name))
		}
		// Pad legend
		for len(legendLines) < gridH {
			legendLines = append(legendLines, "")
		}
		legendStr := strings.Join(legendLines[:gridH], "\n")

		legendPanel := lipgloss.NewStyle().Width(legendW).Render(legendStr)
		mapStr = SideBySide(mapStr, legendPanel)
	}

	// ── Detail line for selected planet ───────────────────
	detailLine := ""
	if n.Selected >= 0 && n.Selected < len(n.Snapshot.Planets) {
		p := n.Snapshot.Planets[n.Selected]
		var statusStr string
		switch p.Status {
		case game.PlanetInvaded:
			statusStr = StyleDanger.Render("INVADED")
		case game.PlanetContested:
			statusStr = StyleGold.Render("CONTESTED")
		case game.PlanetFree:
			statusStr = StyleSuccess.Render("FREE")
		}

		// Count connections and show link levels
		connInfo := ""
		connCount := 0
		for _, l := range n.Snapshot.Links {
			if l.FromID == p.ID || l.ToID == p.ID {
				connCount++
			}
		}
		connInfo = StyleDim.Render(fmt.Sprintf("  Gates:%d", connCount))

		// Network bonus summary
		bonus := computeNetworkBonus(n.Snapshot, p.ID)
		bonusStr := ""
		if bonus.DamageBoost > 0 || bonus.ShieldRegen > 0 || bonus.SpawnReduction > 0 {
			bonusStr = StyleCyan.Render(fmt.Sprintf("  Bonuses: +%d%%dmg", int(bonus.DamageBoost*100)))
			if bonus.SpawnReduction > 0 {
				bonusStr += StyleSuccess.Render(fmt.Sprintf(" -%d%%spawns", int(bonus.SpawnReduction*100)))
			}
		}

		detailLine = "  " + StyleBright.Render(p.Name) + " " + statusStr +
			StyleDim.Render(fmt.Sprintf("  Threat:%d", p.InvasionLevel)) +
			connInfo + bonusStr
		if p.DefenderCount > 0 {
			detailLine += StyleCyan.Render(fmt.Sprintf("  %d defending", p.DefenderCount))
		}
	}

	// ── Key hints ─────────────────────────────────────────
	var keys string
	switch n.Mode {
	case NetworkBrowse:
		keys = "  " +
			FormatKeyHint("↑↓", "navigate") + "  " +
			FormatKeyHint("⏎", "deploy") + "  " +
			FormatKeyHint("u", "upgrade") + "  " +
			FormatKeyHint("s", "transfer") + "  " +
			FormatKeyHint("g", "galaxy") + "  " +
			FormatKeyHint("a", "astro") + "  " +
			FormatKeyHint("q", "back")
	case NetworkUpgrade:
		keys = "  " +
			FormatKeyHint("←→", "select link") + "  " +
			FormatKeyHint("⏎", "confirm upgrade") + "  " +
			FormatKeyHint("esc", "cancel")
	case NetworkTransfer:
		keys = "  " +
			FormatKeyHint("1", "shield +20HP (30 ZPM)") + "  " +
			FormatKeyHint("2", "drones +2 (50 ZPM)") + "  " +
			FormatKeyHint("3", "ZPM gift (25)") + "  " +
			FormatKeyHint("esc", "cancel")
	}

	content := title + "\n" + mapStr + "\n" + detailLine + "\n" + keys

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2).
		Height(h - 2).
		Render(content)
}

// computeNetworkBonus sums gate link bonuses for a planet.
func computeNetworkBonus(snap *engine.GalaxySnapshot, planetID int) game.GateLinkBonus {
	var total game.GateLinkBonus
	for _, l := range snap.Links {
		if l.FromID == planetID || l.ToID == planetID {
			if l.Level > 0 && l.Level < len(game.GateLinkBonuses) {
				b := game.GateLinkBonuses[l.Level]
				total.ShieldRegen += b.ShieldRegen
				total.DamageBoost += b.DamageBoost
				total.SpawnReduction += b.SpawnReduction
			}
		}
	}
	// Cap spawn reduction at 50%
	if total.SpawnReduction > 0.5 {
		total.SpawnReduction = 0.5
	}
	return total
}
