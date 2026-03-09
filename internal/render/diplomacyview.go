package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

var diplStatusNames = [5]string{"None", "War", "NAP", "Trade", "Alliance"}
var diplStatusColors = [5]string{"#666666", "#FF0000", "#FFAA00", "#00AAFF", "#00FF00"}

// BuildDiplomacyView renders the diplomacy screen. Stub for Layer 5.
func BuildDiplomacyView(snap simulation.Snapshot, factionID, width, height int, renderer *lipgloss.Renderer) string {
	var sb strings.Builder

	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  GALACTIC DIPLOMACY") + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────") + "\n\n")

	// Relations matrix header
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  %-15s", "")))
	for i := 0; i < faction.Count; i++ {
		fStyle := renderer.NewStyle().Foreground(lipgloss.Color(faction.Factions[i].ColorFG)).Bold(true)
		sb.WriteString(fStyle.Render(fmt.Sprintf("%8s", faction.Factions[i].ShortName)))
	}
	sb.WriteString("\n")

	for i := 0; i < faction.Count; i++ {
		fStyle := renderer.NewStyle().Foreground(lipgloss.Color(faction.Factions[i].ColorFG)).Bold(true)
		marker := " "
		if i == factionID {
			marker = ">"
		}
		sb.WriteString(dimStyle.Render("  "+marker) + fStyle.Render(fmt.Sprintf("%-14s", faction.Factions[i].Name)))

		for j := 0; j < faction.Count; j++ {
			if i == j {
				sb.WriteString(dimStyle.Render(fmt.Sprintf("%8s", "---")))
			} else {
				status := snap.Diplomacy.Relations[i][j]
				if status < 0 || status > 4 {
					status = 0
				}
				sStyle := renderer.NewStyle().Foreground(lipgloss.Color(diplStatusColors[status]))
				sb.WriteString(sStyle.Render(fmt.Sprintf("%8s", diplStatusNames[status])))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  Diplomacy actions will be available in a future update.") + "\n\n")
	sb.WriteString(dimStyle.Render("  [Esc] Galaxy Map  [q] Quit") + "\n")

	return sb.String()
}
