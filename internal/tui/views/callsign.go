package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderCallsign draws the callsign entry screen — biometric terminal feel.
func RenderCallsign(current string, frameCount, w, h int) string {
	boxW := 46

	// Cursor pulse
	cursor := "█"
	if (frameCount/10)%2 == 1 {
		cursor = " "
	}
	display := current + cursor

	// Inner content
	var lines []string
	lines = append(lines, "")
	lines = append(lines, CenterH(StyleGold.Render("◉ ATLANTIS EXPEDITION"), boxW-4))
	lines = append(lines, CenterH(StyleSubtitle.Render("PERSONNEL IDENTIFICATION SYSTEM"), boxW-4))
	lines = append(lines, "")
	lines = append(lines, "  Enter your callsign:")
	lines = append(lines, "")

	// Input field in a rounded sub-box
	inputContent := "  > " + StyleGold.Render(display)
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyanDim).
		Width(boxW - 8).
		Padding(0, 1).
		Render(inputContent)
	lines = append(lines, inputBox)

	lines = append(lines, "")
	lines = append(lines, StyleDim.Render("  2-16 chars: letters, numbers, _-"))
	lines = append(lines, StyleDim.Render("  Press ENTER to confirm"))
	lines = append(lines, "")

	inner := strings.Join(lines, "\n")

	// Outer double-line panel
	panel := PanelBox(inner, boxW)

	return Center(panel, w, h)
}
