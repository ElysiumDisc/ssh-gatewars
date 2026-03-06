package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

const HUDRows = 3

// BuildHUD creates the bottom status bar.
func BuildHUD(factionID int, snap simulation.Snapshot, width int, renderer *lipgloss.Renderer) string {
	f := faction.Factions[factionID]
	fgColor := lipgloss.Color(f.ColorFG)

	headerStyle := renderer.NewStyle().Foreground(fgColor).Bold(true)
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	barStyle := renderer.NewStyle().Foreground(fgColor)

	ships := snap.ShipCounts[factionID]
	players := snap.PlayerCounts[factionID]
	territory := snap.Territory.Percents[factionID]
	kills := snap.KillCounts[factionID]
	deaths := snap.DeathCounts[factionID]

	// Territory bar
	barWidth := 20
	filled := int(territory / 100 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	bar := barStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", barWidth-filled))

	// Line 1: Faction header
	line1 := headerStyle.Render(fmt.Sprintf("── %s ── %d online ── %d ships ── ", f.ShortName, players, ships)) +
		bar + dimStyle.Render(fmt.Sprintf(" %.0f%%", territory))

	// Pad to width
	if len(line1) < width {
		line1 += dimStyle.Render(strings.Repeat("─", width-lipgloss.Width(line1)))
	}

	// Line 2: Stats
	line2 := dimStyle.Render(fmt.Sprintf("  Kills: %d  |  Deaths: %d  |  Power: ", kills, deaths)) +
		headerStyle.Render("READY [SPACE]")

	// Line 3: Controls
	line3 := dimStyle.Render("  [1-5] Focus sector  |  [Tab] Views  |  [?] Help  |  [q] Quit")

	return line1 + "\n" + line2 + "\n" + line3
}
