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
	territory := 20.0
	if snap.Territory != nil {
		territory = snap.Territory.Percents[factionID]
	}
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

	// Line 2: Stats + power cooldown
	ps := snap.PowerStatuses[factionID]
	var powerStr string
	switch ps.State {
	case simulation.PowerReady:
		powerStr = renderer.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("READY [SPACE]")
	case simulation.PowerActive:
		remaining := int(ps.Remaining.Seconds())
		barW := 10
		frac := 0.0
		if ps.Total > 0 {
			frac = float64(ps.Remaining) / float64(ps.Total)
		}
		filled := int(frac * float64(barW))
		if filled > barW {
			filled = barW
		}
		powerBar := headerStyle.Render(strings.Repeat("█", filled)) + dimStyle.Render(strings.Repeat("░", barW-filled))
		powerStr = headerStyle.Render(fmt.Sprintf("ACTIVE %ds ", remaining)) + powerBar
	case simulation.PowerCooldown:
		remaining := int(ps.Remaining.Seconds())
		barW := 10
		frac := 0.0
		if ps.Total > 0 {
			frac = 1.0 - float64(ps.Remaining)/float64(ps.Total)
		}
		filled := int(frac * float64(barW))
		if filled > barW {
			filled = barW
		}
		powerBar := dimStyle.Render(strings.Repeat("░", barW-filled)) + barStyle.Render(strings.Repeat("█", filled))
		powerStr = dimStyle.Render(fmt.Sprintf("CD %ds ", remaining)) + powerBar
	}
	line2 := dimStyle.Render(fmt.Sprintf("  Kills: %d  |  Deaths: %d  |  %s: ", kills, deaths, ps.Name)) + powerStr

	// Line 3: Controls
	line3 := dimStyle.Render("  [1-5] Focus sector  |  [Tab] Views  |  [?] Help  |  [q] Quit")

	return line1 + "\n" + line2 + "\n" + line3
}
