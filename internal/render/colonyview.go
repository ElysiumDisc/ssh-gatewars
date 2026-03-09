package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

var SliderNames = [5]string{"Ship", "Defense", "Industry", "Ecology", "Research"}

// BuildColonyView renders the colony management screen.
func BuildColonyView(snap simulation.Snapshot, systemID, factionID, selectedSlider, width, height int, renderer *lipgloss.Renderer) string {
	col, ok := snap.Colonies[systemID]
	if !ok {
		return "  No colony here"
	}

	if systemID >= len(snap.Systems) {
		return "  Invalid system"
	}
	sys := snap.Systems[systemID]

	var sb strings.Builder

	f := faction.Factions[col.Faction]
	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color(f.ColorFG))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	valStyle := renderer.NewStyle().Foreground(lipgloss.Color("#AADDFF"))
	borderStyle := renderer.NewStyle().Foreground(lipgloss.Color("#40E0D0"))
	activeStyle := renderer.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	barFill := renderer.NewStyle().Foreground(lipgloss.Color("#44DD44"))
	barEmpty := renderer.NewStyle().Foreground(lipgloss.Color("#333333"))

	sb.WriteString("\n")
	sb.WriteString(borderStyle.Render("  ┌─ COLONY: ") + titleStyle.Render(sys.Name) + borderStyle.Render(" ─────────────────────────────┐") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")

	// Population & factories
	sb.WriteString(borderStyle.Render("  │  ") +
		dimStyle.Render("Population: ") + valStyle.Render(fmt.Sprintf("%.1f / %d", col.Population, col.MaxPop)) + "\n")
	sb.WriteString(borderStyle.Render("  │  ") +
		dimStyle.Render("Factories:  ") + valStyle.Render(fmt.Sprintf("%d / %d", col.Factories, col.MaxFactory)) + "\n")
	sb.WriteString(borderStyle.Render("  │  ") +
		dimStyle.Render("Waste:      ") + valStyle.Render(fmt.Sprintf("%.1f", col.Waste)) + "\n")
	sb.WriteString(borderStyle.Render("  │  ") +
		dimStyle.Render("Output:     ") + valStyle.Render(fmt.Sprintf("%.1f/s", col.TotalOutput)) + "\n")
	sb.WriteString(borderStyle.Render("  │  ") +
		dimStyle.Render("Defenses:   ") + valStyle.Render(fmt.Sprintf("%d missile bases", col.MissileBases)) + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")

	// Sliders
	sliderValues := [5]int{col.SliderShip, col.SliderDefense, col.SliderIndustry, col.SliderEcology, col.SliderResearch}
	sliderOutputs := [5]float64{col.ShipOutput, col.DefenseOutput, col.IndustryOutput, col.EcologyOutput, col.ResearchOutput}

	own := col.Faction == factionID

	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("PRODUCTION SLIDERS") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")

	for i := 0; i < 5; i++ {
		prefix := "  "
		nameStyle := dimStyle
		if own && i == selectedSlider {
			prefix = "> "
			nameStyle = activeStyle
		}

		// Slider bar (20 chars wide)
		filled := sliderValues[i] / 5
		if filled > 20 {
			filled = 20
		}
		bar := barFill.Render(strings.Repeat("█", filled)) +
			barEmpty.Render(strings.Repeat("░", 20-filled))

		line := borderStyle.Render("  │") + nameStyle.Render(fmt.Sprintf("%s%-10s", prefix, SliderNames[i])) +
			"[" + bar + "] " +
			valStyle.Render(fmt.Sprintf("%3d%%", sliderValues[i])) +
			dimStyle.Render(fmt.Sprintf("  %.1f/s", sliderOutputs[i]))

		sb.WriteString(line + "\n")
	}

	sb.WriteString(borderStyle.Render("  │") + "\n")

	// Build queue
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("BUILD QUEUE: "))
	if len(col.BuildQueue) == 0 {
		sb.WriteString(dimStyle.Render("(empty — ships not yet available)"))
	} else {
		for i, name := range col.BuildQueue {
			if i > 0 {
				sb.WriteString(dimStyle.Render(", "))
			}
			sb.WriteString(valStyle.Render(name))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  └───────────────────────────────────────────────────┘") + "\n")

	if own {
		sb.WriteString(dimStyle.Render("  [↑↓]Select Slider  [←→]Adjust  [Esc]System View  [q]Quit") + "\n")
	} else {
		sb.WriteString(dimStyle.Render("  (Enemy colony — view only)  [Esc]System View  [q]Quit") + "\n")
	}

	return sb.String()
}
