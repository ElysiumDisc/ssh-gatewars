package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/simulation"
)

// BuildSystemView renders the system detail panel.
func BuildSystemView(snap simulation.Snapshot, systemID, factionID, width, height int, renderer *lipgloss.Renderer) string {
	if systemID < 0 || systemID >= len(snap.Systems) {
		return "Invalid system"
	}

	sys := snap.Systems[systemID]
	var sb strings.Builder

	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	valStyle := renderer.NewStyle().Foreground(lipgloss.Color("#AADDFF"))
	borderStyle := renderer.NewStyle().Foreground(lipgloss.Color("#40E0D0"))

	sb.WriteString("\n")
	sb.WriteString(borderStyle.Render("  ┌─ SYSTEM: ") + titleStyle.Render(sys.Name) + borderStyle.Render(" ─────────────────────────┐") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")

	// Star info
	starColor := simulation.StarTypeColors[sys.StarType]
	sStyle := renderer.NewStyle().Foreground(lipgloss.Color(starColor)).Bold(true)
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Star: ") + sStyle.Render(simulation.StarTypeNames[sys.StarType]) + "\n")

	// Planet info
	if sys.HasPlanet {
		pt := gamedata.PlanetTypes[sys.PlanetType]
		pColor := renderer.NewStyle().Foreground(lipgloss.Color(pt.Color)).Bold(true)
		sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Planet: ") +
			pColor.Render(pt.Name) +
			dimStyle.Render("  Size: ") + valStyle.Render(gamedata.SizeNames[sys.PlanetSize]) +
			dimStyle.Render("  Minerals: ") + valStyle.Render(gamedata.MineralNames[sys.Minerals]) + "\n")

		maxPop := gamedata.MaxPop(sys.PlanetType, sys.PlanetSize)
		sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render(fmt.Sprintf("Max Population: %d", maxPop)) + "\n")

		if !pt.Habitable {
			sb.WriteString(borderStyle.Render("  │  ") +
				renderer.NewStyle().Foreground(lipgloss.Color("#FF4444")).Render("Hostile environment — requires "+pt.TechRequired) + "\n")
		}
	} else {
		sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("No planet in this system") + "\n")
	}

	// Special
	if sys.Special == simulation.SpecialDakara {
		sb.WriteString(borderStyle.Render("  │  ") +
			renderer.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render("Ancient Superweapon — Dakara Device") + "\n")
	} else if sys.Special == simulation.SpecialArtifact {
		sb.WriteString(borderStyle.Render("  │  ") +
			renderer.NewStyle().Foreground(lipgloss.Color("#88FFFF")).Bold(true).Render("Ancient Outpost — Research Bonus") + "\n")
	}

	sb.WriteString(borderStyle.Render("  │") + "\n")

	// Colony info
	if col, ok := snap.Colonies[sys.ID]; ok {
		f := faction.Factions[col.Faction]
		fStyle := renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)
		sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Owner: ") + fStyle.Render(f.Name) + "\n")
		sb.WriteString(borderStyle.Render("  │  ") +
			dimStyle.Render(fmt.Sprintf("Population: %.1f / %d", col.Population, col.MaxPop)) + "\n")
		sb.WriteString(borderStyle.Render("  │  ") +
			dimStyle.Render(fmt.Sprintf("Factories:  %d / %d", col.Factories, col.MaxFactory)) + "\n")
		sb.WriteString(borderStyle.Render("  │  ") +
			dimStyle.Render(fmt.Sprintf("Output:     %.1f/s", col.TotalOutput)) + "\n")
		sb.WriteString(borderStyle.Render("  │  ") +
			dimStyle.Render(fmt.Sprintf("Defenses:   %d missile bases", col.MissileBases)) + "\n")

		if col.Faction == factionID {
			sb.WriteString(borderStyle.Render("  │") + "\n")
			sb.WriteString(borderStyle.Render("  │  ") +
				renderer.NewStyle().Foreground(lipgloss.Color("#AAFFAA")).Render("[Enter] Manage Colony") + "\n")
		}
	} else if sys.Owner >= 0 {
		f := faction.Factions[sys.Owner]
		fStyle := renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)
		sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Controlled by: ") + fStyle.Render(f.Name) + "\n")
	} else {
		sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Uncolonized") + "\n")
	}

	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  └───────────────────────────────────────────┘") + "\n")
	sb.WriteString(dimStyle.Render("  [Arrows]Navigate  [Enter]Manage  [Esc]Galaxy Map  [q]Quit") + "\n")

	return sb.String()
}
