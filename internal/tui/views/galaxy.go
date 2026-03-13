package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/engine"
	"ssh-gatewars/internal/game"
)

// GalaxyModel holds state for the galaxy browser.
type GalaxyModel struct {
	Snapshot  *engine.GalaxySnapshot
	Selected  int
	ScrollTop int
}

func NewGalaxyModel() GalaxyModel {
	return GalaxyModel{}
}

func (g *GalaxyModel) Reset(snap *engine.GalaxySnapshot) {
	g.Snapshot = snap
	g.Selected = 0
	g.ScrollTop = 0
}

func (g *GalaxyModel) MoveSelection(delta int) {
	if g.Snapshot == nil {
		return
	}
	g.Selected += delta
	if g.Selected < 0 {
		g.Selected = 0
	}
	max := len(g.Snapshot.Planets) - 1
	if g.Selected > max {
		g.Selected = max
	}
}

func (g *GalaxyModel) SelectedPlanetID() int {
	if g.Snapshot == nil || len(g.Snapshot.Planets) == 0 {
		return -1
	}
	return g.Snapshot.Planets[g.Selected].ID
}

// RenderGalaxy draws the galaxy browser — sensor display aesthetic.
func RenderGalaxy(g GalaxyModel, w, h int) string {
	innerH := h - 4 // outer border

	// ── Title bar ──────────────────────────────────────────────
	cycleStr := ""
	if g.Snapshot.Cycle > 1 {
		cycleStr = StyleDanger.Render(fmt.Sprintf("  CYCLE %d", g.Snapshot.Cycle))
	}
	freePct := StyleDim.Render(fmt.Sprintf("  [%d%% FREE]", g.Snapshot.FreePct))
	title := StyleGold.Render("◆ LONG RANGE SENSORS") + StyleDim.Render(" — ") + StyleCyan.Render("GALAXY MAP") + freePct + cycleStr

	// ── Content ────────────────────────────────────────────────
	if g.Snapshot == nil {
		content := title + "\n\n  " + StyleDim.Render("Scanning galaxy...")
		return lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorCyan).
			Width(w - 2).
			Height(h - 2).
			Render(content)
	}

	// Show detail panel on wide terminals
	showDetail := w > 90
	listW := w - 6
	detailW := 0
	if showDetail {
		detailW = 30
		listW = w - detailW - 8
	}

	// ── Planet list ────────────────────────────────────────────
	// Column header
	header := fmt.Sprintf("  %-3s %-14s %-11s %-6s %-10s",
		"", "PLANET", "STATUS", "DIFF", "DEFENDERS")
	colHeader := StyleDim.Render(header)
	separator := StyleCyanDim.Render("  " + strings.Repeat("─", minInt(listW, 55)))

	// Visible rows
	visibleRows := innerH - 6 // title + header + separator + footer stats + footer keys + spacer
	if visibleRows < 3 {
		visibleRows = 3
	}

	// Adjust scroll
	if g.Selected < g.ScrollTop {
		g.ScrollTop = g.Selected
	}
	if g.Selected >= g.ScrollTop+visibleRows {
		g.ScrollTop = g.Selected - visibleRows + 1
	}

	end := g.ScrollTop + visibleRows
	if end > len(g.Snapshot.Planets) {
		end = len(g.Snapshot.Planets)
	}

	var rows []string
	for i := g.ScrollTop; i < end; i++ {
		p := g.Snapshot.Planets[i]
		rows = append(rows, renderPlanetRow(p, i == g.Selected, listW))
	}

	// Pad rows to fill
	for len(rows) < visibleRows {
		rows = append(rows, "")
	}

	// ── Aggregate stats ────────────────────────────────────────
	invaded, contested, freed := 0, 0, 0
	for _, p := range g.Snapshot.Planets {
		switch p.Status {
		case game.PlanetInvaded:
			invaded++
		case game.PlanetContested:
			contested++
		case game.PlanetFree:
			freed++
		}
	}
	stats := "  " +
		StyleDanger.Render(fmt.Sprintf("● %d invaded", invaded)) + "  " +
		StyleGold.Render(fmt.Sprintf("◆ %d contested", contested)) + "  " +
		StyleSuccess.Render(fmt.Sprintf("✧ %d free", freed)) + "  " +
		StyleDim.Render(fmt.Sprintf("(%d total)", len(g.Snapshot.Planets)))

	// ── Key hints ──────────────────────────────────────────────
	keys := "  " +
		FormatKeyHint("↑↓", "navigate") + "  " +
		FormatKeyHint("⏎", "deploy") + "  " +
		FormatKeyHint("q", "back")

	// ── Assemble list panel ────────────────────────────────────
	var listLines []string
	listLines = append(listLines, " "+title)
	listLines = append(listLines, "")
	listLines = append(listLines, colHeader)
	listLines = append(listLines, separator)
	listLines = append(listLines, rows...)
	listLines = append(listLines, "")
	listLines = append(listLines, stats)
	listLines = append(listLines, keys)

	listContent := strings.Join(listLines, "\n")

	// ── Detail panel (wide terminals) ──────────────────────────
	if showDetail && g.Selected >= 0 && g.Selected < len(g.Snapshot.Planets) {
		detail := renderPlanetDetail(g.Snapshot.Planets[g.Selected], detailW, innerH)
		listStyled := lipgloss.NewStyle().Width(listW + 4).Render(listContent)
		combined := SideBySide(listStyled, detail)

		return lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorCyan).
			Width(w - 2).
			Height(h - 2).
			Render(combined)
	}

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2).
		Height(h - 2).
		Render(listContent)
}

func renderPlanetRow(p engine.PlanetSnap, selected bool, maxW int) string {
	// Status symbol + text
	var statusSym, statusText string
	switch p.Status {
	case game.PlanetInvaded:
		statusSym = StyleDanger.Render("●")
		statusText = StyleDanger.Render("INVADED")
	case game.PlanetContested:
		statusSym = StyleGold.Render("◆")
		statusText = StyleGold.Render("CONTESTED")
	case game.PlanetFree:
		statusSym = StyleSuccess.Render("✧")
		statusText = StyleSuccess.Render("FREE")
	}

	defenders := ""
	if p.DefenderCount > 0 {
		defenders = StyleCyan.Render(fmt.Sprintf("%d online", p.DefenderCount))
	}

	// Surge + bounty tags
	tags := ""
	if p.Surging {
		tags += " " + StyleDanger.Render("⚠SURGE")
	}
	if p.BountyZPM > 0 && p.Status != game.PlanetFree {
		tags += " " + StyleGold.Render(fmt.Sprintf("⚡%d", p.BountyZPM))
	}

	if selected {
		plainStatus := statusTextPlain(p.Status)
		defText := ""
		if p.DefenderCount > 0 {
			defText = fmt.Sprintf("%d online", p.DefenderCount)
		}
		surgeTag := ""
		if p.Surging {
			surgeTag = " SURGE"
		}
		line := fmt.Sprintf("  ▸ %-14s %-11s %-4d %s%s",
			p.Name, plainStatus, p.InvasionLevel, defText, surgeTag)
		return StyleHighlight.Render(truncate(line, maxW))
	}

	return fmt.Sprintf("  %s %-14s %s %s %s%s",
		statusSym,
		StyleBright.Render(p.Name),
		statusText,
		StyleMid.Render(fmt.Sprintf("%-4d", p.InvasionLevel)),
		defenders,
		tags)
}

func renderPlanetDetail(p engine.PlanetSnap, w, h int) string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, " "+StyleCyan.Render("◆ ")+StyleBright.Render(p.Name))
	lines = append(lines, "")

	// Status
	var statusStr string
	switch p.Status {
	case game.PlanetInvaded:
		statusStr = StyleDanger.Render("● INVADED")
	case game.PlanetContested:
		statusStr = StyleGold.Render("◆ CONTESTED")
	case game.PlanetFree:
		statusStr = StyleSuccess.Render("✧ FREE")
	}
	lines = append(lines, " "+StyleDim.Render("Status: ")+statusStr)
	lines = append(lines, "")

	// Difficulty bar
	diffW := w - 14
	if diffW < 5 {
		diffW = 5
	}
	diffPct := float64(p.InvasionLevel) / 10.0
	diffBar := ProgressBar(diffPct, diffW, ColorDanger, ColorDim)
	lines = append(lines, " "+StyleDim.Render("Threat: ")+diffBar+StyleMid.Render(fmt.Sprintf(" %d", p.InvasionLevel)))
	lines = append(lines, "")

	// Defenders
	defStr := StyleDim.Render("none")
	if p.DefenderCount > 0 {
		defStr = StyleCyan.Render(fmt.Sprintf("%d active", p.DefenderCount))
	}
	lines = append(lines, " "+StyleDim.Render("Defenders: ")+defStr)

	// Bounty
	if p.BountyZPM > 0 && p.Status != game.PlanetFree {
		lines = append(lines, "")
		lines = append(lines, " "+StyleDim.Render("Bounty: ")+StyleGold.Render(fmt.Sprintf("⚡%d ZPM", p.BountyZPM)))
	}

	// Surge
	if p.Surging {
		lines = append(lines, "")
		lines = append(lines, " "+StyleDanger.Render("⚠ REPLICATOR SURGE"))
		lines = append(lines, " "+StyleDim.Render("2x spawns, 2x ZPM"))
	}

	content := strings.Join(lines, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyanDim).
		Width(w - 2).
		Height(h - 2).
		Render(content)
}

func statusTextPlain(s game.PlanetStatus) string {
	switch s {
	case game.PlanetInvaded:
		return "INVADED"
	case game.PlanetContested:
		return "CONTESTED"
	case game.PlanetFree:
		return "FREE"
	}
	return "?"
}
