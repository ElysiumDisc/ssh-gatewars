package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Color palette ─────────────────────────────────────────────────────

var (
	ColorCyan    = lipgloss.Color("#00D9FF") // Ancient cyan — primary UI
	ColorCyanDim = lipgloss.Color("#007A8F") // secondary borders, subtle text
	ColorGold    = lipgloss.Color("#FFD700") // ZPM gold — accents, energy
	ColorDanger  = lipgloss.Color("#FF3333") // damage, alerts
	ColorSuccess = lipgloss.Color("#00FF88") // liberated, positive

	ColorRepGrey    = lipgloss.Color("#808080") // basic replicator
	ColorRepHot     = lipgloss.Color("#A0A0A0") // armored replicator
	ColorRepQueen   = lipgloss.Color("#FF2020") // queen replicator
	ColorDroneGold  = lipgloss.Color("#FFD700")
	ColorDroneCyan  = lipgloss.Color("#00D9FF")
	ColorDroneMag   = lipgloss.Color("#FF00FF")
	ColorDroneWhite = lipgloss.Color("#FFFFFF")

	ColorDim    = lipgloss.Color("#4A5568") // dim text
	ColorMid    = lipgloss.Color("#8899AA") // mid text
	ColorBright = lipgloss.Color("#E0E8F0") // bright text
	ColorBg     = lipgloss.Color("#0A0E14") // dark background
)

// ── Pre-built styles ──────────────────────────────────────────────────

var (
	StyleTitle    = lipgloss.NewStyle().Foreground(ColorGold).Bold(true)
	StyleSubtitle = lipgloss.NewStyle().Foreground(ColorCyanDim)
	StyleDim      = lipgloss.NewStyle().Foreground(ColorDim)
	StyleMid      = lipgloss.NewStyle().Foreground(ColorMid)
	StyleBright   = lipgloss.NewStyle().Foreground(ColorBright)

	StyleCyan    = lipgloss.NewStyle().Foreground(ColorCyan)
	StyleCyanDim = lipgloss.NewStyle().Foreground(ColorCyanDim)
	StyleGold    = lipgloss.NewStyle().Foreground(ColorGold)
	StyleDanger  = lipgloss.NewStyle().Foreground(ColorDanger)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)

	StyleHighlight = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0A0E14")).
			Background(ColorGold).
			Bold(true)

	StyleKeyHint = lipgloss.NewStyle().Foreground(ColorMid)
	StyleKeyChar = lipgloss.NewStyle().Foreground(ColorCyan).Bold(true)

	// ── Entity style lookup tables (allocated once) ───────────────

	droneStyles = map[int]lipgloss.Style{
		0: lipgloss.NewStyle().Foreground(ColorDroneGold),
		1: lipgloss.NewStyle().Foreground(ColorDroneCyan),
		2: lipgloss.NewStyle().Foreground(ColorDroneMag),
		3: lipgloss.NewStyle().Foreground(ColorDroneWhite),
	}
	repStyles = map[int]lipgloss.Style{
		0: lipgloss.NewStyle().Foreground(ColorRepGrey),
		1: lipgloss.NewStyle().Foreground(ColorRepHot),
		2: lipgloss.NewStyle().Foreground(ColorRepQueen).Bold(true),
	}
)

// ── Layout helpers ────────────────────────────────────────────────────

// CenterH centers a string horizontally within width w.
func CenterH(s string, w int) string {
	return lipgloss.PlaceHorizontal(w, lipgloss.Center, s)
}

// CenterV centers a string vertically within height h.
func CenterV(s string, h int) string {
	return lipgloss.PlaceVertical(h, lipgloss.Center, s)
}

// Center centers a string both horizontally and vertically.
func Center(s string, w, h int) string {
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, s)
}

// SideBySide joins two blocks horizontally.
func SideBySide(left, right string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// FormatKeyHint renders [k] Label with styled key character.
func FormatKeyHint(key, label string) string {
	return StyleMid.Render("[") +
		StyleKeyChar.Render(key) +
		StyleMid.Render("] ") +
		StyleMid.Render(label)
}

// PanelBox wraps content in a double-line border.
func PanelBox(content string, w int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2). // account for border chars
		Render(content)
}

// RoundedBox wraps content in a rounded border.
func RoundedBox(content string, w int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyanDim).
		Width(w - 2).
		Render(content)
}

// ProgressBar renders a horizontal bar at the given percentage.
func ProgressBar(pct float64, width int, filledColor, emptyColor lipgloss.Color) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	fStyle := lipgloss.NewStyle().Foreground(filledColor)
	eStyle := lipgloss.NewStyle().Foreground(emptyColor)
	return fStyle.Render(strings.Repeat("▓", filled)) + eStyle.Render(strings.Repeat("░", empty))
}

// ShieldBar renders a shield health bar with color gradient: green→yellow→red.
func ShieldBar(current, max, width int) string {
	if max <= 0 {
		return strings.Repeat("░", width)
	}
	pct := float64(current) / float64(max)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	var color lipgloss.Color
	switch {
	case pct > 0.6:
		color = ColorSuccess
	case pct > 0.3:
		color = lipgloss.Color("#FFAA00") // yellow-orange
	default:
		color = ColorDanger
	}

	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	fStyle := lipgloss.NewStyle().Foreground(color)
	eStyle := lipgloss.NewStyle().Foreground(ColorDim)
	return fStyle.Render(strings.Repeat("▓", filled)) + eStyle.Render(strings.Repeat("░", empty))
}

// DroneStyle returns the pre-built style for a drone tier.
func DroneStyle(tier int) lipgloss.Style {
	if s, ok := droneStyles[tier]; ok {
		return s
	}
	return droneStyles[0]
}

// RepStyle returns the pre-built style for a replicator type.
func RepStyle(rtype int) lipgloss.Style {
	if s, ok := repStyles[rtype]; ok {
		return s
	}
	return repStyles[0]
}

// droneColorName renders a drone tier name in its color.
func droneColorName(tier int) string {
	names := map[int]string{0: "Standard", 1: "Swift", 2: "Blast", 3: "Piercing"}
	name := names[tier]
	if name == "" {
		name = "Standard"
	}
	return DroneStyle(tier).Render(name)
}

// wrapText performs word-aware wrapping on plain text before styling.
func wrapText(s string, maxLen int) []string {
	if maxLen <= 0 {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		if len(current)+1+len(w) > maxLen {
			lines = append(lines, current)
			current = w
		} else {
			current += " " + w
		}
	}
	lines = append(lines, current)
	return lines
}

// pad returns n spaces, or empty string if n <= 0.
func pad(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

// truncate trims a string to maxLen runes.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

// maxInt returns the larger of two ints.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// clampInt clamps v between lo and hi.
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// fmtInt formats an integer with the given style.
func fmtInt(n int, s lipgloss.Style) string {
	return s.Render(fmt.Sprintf("%d", n))
}
