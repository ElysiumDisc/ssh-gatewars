package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/simulation"
)

// BuildTechView renders the tech tree browser. Stub for Layer 4.
func BuildTechView(snap simulation.Snapshot, factionID, selectedTree, width, height int, renderer *lipgloss.Renderer) string {
	var sb strings.Builder

	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	activeStyle := renderer.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	valStyle := renderer.NewStyle().Foreground(lipgloss.Color("#AADDFF"))

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  RESEARCH ALLOCATION") + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────") + "\n\n")

	if factionID < 0 {
		sb.WriteString(dimStyle.Render("  No faction") + "\n")
		return sb.String()
	}

	fs := snap.Factions[factionID]

	for i := 0; i < gamedata.TreeCount; i++ {
		treeColor := renderer.NewStyle().Foreground(lipgloss.Color(gamedata.TreeColors[i])).Bold(true)
		nameStyle := dimStyle
		prefix := "  "
		if i == selectedTree {
			prefix = "> "
			nameStyle = activeStyle
		}

		// Bar (20 chars)
		filled := fs.TechAlloc[i] / 5
		if filled > 20 {
			filled = 20
		}
		bar := treeColor.Render(strings.Repeat("█", filled)) +
			dimStyle.Render(strings.Repeat("░", 20-filled))

		line := fmt.Sprintf("  %s%-22s [%s] %3d%%  Tier %d  RP:%.0f",
			prefix, nameStyle.Render(gamedata.TreeNames[i]),
			bar, fs.TechAlloc[i], fs.TechTiers[i], fs.TechRP[i])
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  Total Research: ") +
		valStyle.Render(fmt.Sprintf("%.1f/s", snap.Factions[factionID].TotalProd)) + "\n")
	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  [↑↓]Select Tree  [←→]Adjust  [Esc]Galaxy Map  [q]Quit") + "\n")

	return sb.String()
}
