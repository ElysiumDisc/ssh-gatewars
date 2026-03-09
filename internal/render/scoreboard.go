package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/simulation"
)

// BuildScoreboard renders the faction standings.
func BuildScoreboard(snap simulation.Snapshot, factionID, width, height int, renderer *lipgloss.Renderer) string {
	var sb strings.Builder

	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	borderStyle := renderer.NewStyle().Foreground(lipgloss.Color("#40E0D0")).Bold(true)
	headerStyle := renderer.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Bold(true)

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  GALACTIC STANDINGS") + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────────────────────────") + "\n\n")

	// Header
	sb.WriteString(headerStyle.Render(fmt.Sprintf("  %-18s %7s %7s %8s %7s %7s",
		"FACTION", "SYSTEMS", "POP", "NAQUADAH", "ONLINE", "TECH")) + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────────────────────────") + "\n")

	for i := 0; i < faction.Count; i++ {
		f := faction.Factions[i]
		fs := snap.Factions[i]
		fStyle := renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)

		// Average tech tier
		totalTier := 0
		for _, t := range fs.TechTiers {
			totalTier += t
		}
		avgTech := float64(totalTier) / float64(gamedata.TreeCount)

		marker := " "
		if i == factionID {
			marker = borderStyle.Render(">")
		}

		line := marker + fStyle.Render(fmt.Sprintf(" %-17s", f.Name)) +
			dimStyle.Render(fmt.Sprintf(" %7d %7.0f %8.0f %7d %6.1f",
				fs.SystemCount, fs.Population, fs.Naquadah, snap.PlayerCounts[i], avgTech))

		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n")

	// Campaign status
	switch snap.Campaign.State {
	case simulation.CampaignActive:
		sb.WriteString(borderStyle.Render("  Campaign: ") + dimStyle.Render("ACTIVE") + "\n")
	case simulation.CampaignWon:
		winner := "???"
		if snap.Campaign.Winner >= 0 && snap.Campaign.Winner < faction.Count {
			winner = faction.Factions[snap.Campaign.Winner].Name
		}
		sb.WriteString(renderer.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).
			Render(fmt.Sprintf("  Victory: %s", winner)) + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  [Tab/Esc] Galaxy Map  |  [q] Quit") + "\n")

	return sb.String()
}
