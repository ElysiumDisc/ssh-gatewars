package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var gateArt = `
       ╭────────────────╮
     ╭─┤  ◆  ◇  △  ▽  ├─╮
   ╭─┤  ╰────────────────╯  ├─╮
   │  ╲         ░░░         ╱  │
   │   ╲    ░░░░░░░░░░    ╱   │
   │    │  ░░░░░░░░░░░░  │    │
   │    │  ░░ GATEWARS ░░  │    │
   │    │  ░░░░░░░░░░░░  │    │
   │   ╱    ░░░░░░░░░░    ╲   │
   │  ╱         ░░░         ╲  │
   ╰─┤  ╭────────────────╮  ├─╯
     ╰─┤  ○  □  ☆  ◈  ├─╯
       ╰────────────────╯
`

// RenderSplash renders the title screen.
func RenderSplash(width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#44AAFF")).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA44")).
		Bold(true)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render(gateArt))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("  A Stargate SG-1 Roguelike  —  via SSH"))
	b.WriteString("\n\n")
	b.WriteString(promptStyle.Render("  Press ENTER to embark through the gate..."))
	b.WriteString("\n")

	return b.String()
}
