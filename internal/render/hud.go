package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

const HUDRows = 3

// BuildHUD creates the bottom status bar.
func BuildHUD(factionID int, snap simulation.Snapshot, width int, renderer *lipgloss.Renderer) string {
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	naqStyle := renderer.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)

	// Line 1: Campaign status
	var line1 string
	switch snap.Campaign.State {
	case simulation.CampaignActive:
		line1 = renderer.NewStyle().Foreground(lipgloss.Color("#40E0D0")).Bold(true).
			Render("── GALACTIC CONQUEST ──")
	case simulation.CampaignWon:
		winner := "???"
		if snap.Campaign.Winner >= 0 && snap.Campaign.Winner < faction.Count {
			winner = faction.Factions[snap.Campaign.Winner].Name
		}
		line1 = renderer.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).
			Render(fmt.Sprintf("── VICTORY! %s conquers the galaxy! ──", winner))
	}

	if snap.Paused {
		line1 += dimStyle.Render(" [STANDBY]")
	}

	padW := width - lipgloss.Width(line1)
	if padW > 0 {
		line1 += dimStyle.Render(strings.Repeat("─", padW))
	}

	// Line 2: Faction + resources
	factionName := "???"
	naq := 0.0
	if factionID >= 0 && factionID < faction.Count {
		factionName = faction.Factions[factionID].ShortName
		naq = snap.Factions[factionID].Naquadah
	}

	fStyle := renderer.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	if factionID >= 0 && factionID < faction.Count {
		fStyle = renderer.NewStyle().Foreground(lipgloss.Color(faction.Factions[factionID].ColorFG)).Bold(true)
	}

	players := 0
	for _, c := range snap.PlayerCounts {
		players += c
	}

	campaignAge := time.Since(snap.Campaign.StartedAt).Round(time.Minute)
	ageStr := fmt.Sprintf("%dh%dm", int(campaignAge.Hours()), int(campaignAge.Minutes())%60)

	systems := 0
	if factionID >= 0 && factionID < faction.Count {
		systems = snap.Factions[factionID].SystemCount
	}

	line2 := fStyle.Render("  "+factionName) + "  " +
		naqStyle.Render(fmt.Sprintf("Naq:%.0f", naq)) + "  " +
		dimStyle.Render(fmt.Sprintf("Systems:%d  Online:%d  Age:%s", systems, players, ageStr))

	// Line 3: Controls
	line3 := dimStyle.Render("  [Arrows]Navigate  [Enter]System  [t]Tech  [d]Diplo  [Tab]Score  [?]Help  [q]Quit")

	return line1 + "\n" + line2 + "\n" + line3
}
