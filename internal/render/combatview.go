package render

import (
	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/simulation"
)

// BuildCombatView renders tactical combat. Stub for Layer 3.
func BuildCombatView(snap simulation.Snapshot, width, height int, renderer *lipgloss.Renderer) string {
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("#888888"))
	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	return "\n" +
		titleStyle.Render("  TACTICAL COMBAT") + "\n\n" +
		dimStyle.Render("  Combat view will be available in a future update.") + "\n\n" +
		dimStyle.Render("  [Esc] Galaxy Map  [q] Quit") + "\n"
}
