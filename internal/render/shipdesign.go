package render

import (
	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/simulation"
)

// BuildShipDesigner renders the ship designer. Stub for Layer 2.
func BuildShipDesigner(snap simulation.Snapshot, factionID, width, height int, renderer *lipgloss.Renderer) string {
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	return "\n" +
		titleStyle.Render("  SHIP DESIGNER") + "\n\n" +
		dimStyle.Render("  Ship design will be available in a future update.") + "\n\n" +
		dimStyle.Render("  [Esc] Galaxy Map  [q] Quit") + "\n"
}
