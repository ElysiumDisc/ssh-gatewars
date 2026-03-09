package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

// BuildSplash renders the faction selection screen.
func BuildSplash(snap simulation.Snapshot, width, height int, renderer *lipgloss.Renderer) string {
	var sb strings.Builder

	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#40E0D0"))
	headerStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	center := func(s string) string {
		w := lipgloss.Width(s)
		pad := (width - w) / 2
		if pad < 0 {
			pad = 0
		}
		return strings.Repeat(" ", pad) + s
	}

	// Stargate ASCII art
	gate := []string{
		"        .-'''''-.        ",
		"      .'           '.      ",
		"     /    .-'''-.    \\     ",
		"    |   /         \\   |    ",
		"    |  |  GATEWARS |  |    ",
		"    |   \\         /   |    ",
		"     \\    '-...-'    /     ",
		"      '.           .'      ",
		"        '-......-'        ",
	}

	sb.WriteString("\n")
	for _, line := range gate {
		sb.WriteString(center(titleStyle.Render(line)) + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(center(titleStyle.Render("SSH GATEWARS — MASTER OF ORION")) + "\n")
	sb.WriteString(center(dimStyle.Render("A Stargate SG-1 4X Strategy Game")) + "\n")
	sb.WriteString("\n")
	sb.WriteString(center(headerStyle.Render("CHOOSE YOUR ALLEGIANCE")) + "\n\n")

	for i := 0; i < faction.Count; i++ {
		f := faction.Factions[i]
		fStyle := renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)
		specStyle := renderer.NewStyle().Foreground(lipgloss.Color("#666666"))
		players := snap.PlayerCounts[i]

		line := fmt.Sprintf("  [%d] ", i+1) +
			fStyle.Render(fmt.Sprintf("%-20s", f.Name)) +
			dimStyle.Render(fmt.Sprintf("%d online", players)) +
			specStyle.Render("  "+f.Special)
		sb.WriteString(center(line) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(center(dimStyle.Render("Press 1-5 to choose. Or: ssh <faction>@host")) + "\n")

	return sb.String()
}
