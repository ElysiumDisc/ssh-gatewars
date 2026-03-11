package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HUDData holds the data needed to render the HUD bar.
type HUDData struct {
	HP, MaxHP    int
	Level        int
	XP           int
	WeaponName   string
	PlanetName   string
	Biome        string
	Threat       int
	Flash        string
	OnlineCount  int
}

// RenderHUD renders the bottom HUD bar.
func RenderHUD(data HUDData, width int) string {
	hpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44FF44"))
	if data.HP < data.MaxHP/3 {
		hpStyle = hpStyle.Foreground(lipgloss.Color("#FF4444"))
	} else if data.HP < data.MaxHP*2/3 {
		hpStyle = hpStyle.Foreground(lipgloss.Color("#FFAA44"))
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	flashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF44"))

	hp := hpStyle.Render(fmt.Sprintf("HP:%d/%d", data.HP, data.MaxHP))
	lvl := fmt.Sprintf("Lv%d", data.Level)
	weapon := data.WeaponName
	if weapon == "" {
		weapon = "Fists"
	}

	loc := data.PlanetName
	if data.Threat > 0 {
		loc += fmt.Sprintf(" (%s, Threat %d)", data.Biome, data.Threat)
	}

	left := fmt.Sprintf("%s | %s | %s", hp, lvl, weapon)
	right := dimStyle.Render(loc)

	var b strings.Builder
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")
	b.WriteString(left)

	// Right-align location
	padding := width - lipgloss.Width(left) - lipgloss.Width(right)
	if padding > 0 {
		b.WriteString(strings.Repeat(" ", padding))
	}
	b.WriteString(right)

	// Flash message line
	if data.Flash != "" {
		b.WriteString("\n")
		b.WriteString(flashStyle.Render(data.Flash))
	}

	return b.String()
}
