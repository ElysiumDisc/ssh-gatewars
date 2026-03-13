package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/chat"
	"ssh-gatewars/internal/store"
)

// RenderAtlantis draws the personal hub screen with responsive layout.
func RenderAtlantis(player *store.PlayerRecord, callsign string, msgs []chat.ChatMessage, chatInput string, chatMode bool, onlineCount, w, h int) string {
	// ── Responsive widths ──────────────────────────────────────
	chatPct := 40
	chatW := w * chatPct / 100
	chatW = clampInt(chatW, 25, 45)
	leftW := w - chatW - 1 // 1 for gap

	innerH := h - 4 // outer border top/bottom + top bar + bottom bar

	// ── Top bar ────────────────────────────────────────────────
	topTitle := StyleGold.Render(" ◆ A T L A N T I S ")
	topOnline := StyleDim.Render("  ONLINE: ") + StyleCyan.Render(fmt.Sprintf("%d", onlineCount))
	topZPM := ""
	if player != nil {
		topZPM = StyleDim.Render("  ZPM: ") + StyleGold.Render(fmt.Sprintf("⚡%d", player.ZPMBalance))
	}
	topBarContent := topTitle + topOnline + topZPM
	topBarW := lipgloss.Width(topBarContent)
	topBar := topBarContent + pad(w-topBarW-2)

	// ── Left panel: stats + chair art ──────────────────────────
	statsW := leftW - 4 // some padding
	if statsW < 20 {
		statsW = 20
	}

	var statsLines []string
	if player != nil {
		statsLines = append(statsLines,
			"  "+StyleBright.Render("Callsign")+StyleDim.Render(": ")+StyleCyan.Render(callsign),
			"  "+StyleBright.Render("Chair Lvl")+StyleDim.Render(": ")+fmtInt(player.ChairLevel, StyleGold),
			"  "+StyleBright.Render("Drone")+StyleDim.Render(": ")+droneColorName(player.DroneTier),
			"",
			"  "+StyleCyanDim.Render("── RECORD ──"),
			"  "+StyleMid.Render("Planets Freed")+StyleDim.Render(": ")+fmtInt(player.PlanetsFreed, StyleSuccess),
			"  "+StyleMid.Render("Total Kills")+StyleDim.Render(":   ")+fmtInt(player.TotalKills, StyleBright),
			"  "+StyleMid.Render("Sessions")+StyleDim.Render(":      ")+fmtInt(player.TotalSessions, StyleBright),
		)
	} else {
		statsLines = append(statsLines, "  "+StyleDim.Render("Loading commander data..."))
	}

	statsInner := strings.Join(statsLines, "\n")
	statsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyanDim).
		Width(statsW).
		Padding(1, 0).
		Render(" " + StyleCyanDim.Render("COMMANDER STATUS") + "\n" + statsInner)

	// Chair ASCII art — bigger, more detailed
	chairArt := []string{
		StyleCyanDim.Render("         ╔═════════╗"),
		StyleCyanDim.Render("         ║") + StyleCyan.Render("  ◎   ◎ ") + StyleCyanDim.Render("║"),
		StyleCyanDim.Render("         ║") + StyleGold.Render("    ⌂   ") + StyleCyanDim.Render("║"),
		StyleCyanDim.Render("     ════╩═════════╩════"),
		StyleCyanDim.Render("     ║") + StyleDim.Render("  ▓▓▓▓▓▓▓▓▓▓▓  ") + StyleCyanDim.Render("║"),
		StyleCyanDim.Render("     ╚═════════════════╝"),
	}

	// Assemble left content
	var leftLines []string
	leftLines = append(leftLines, "") // spacer
	leftLines = append(leftLines, strings.Split(statsBox, "\n")...)
	leftLines = append(leftLines, "") // spacer
	leftLines = append(leftLines, chairArt...)

	// Pad left to fill height
	for len(leftLines) < innerH {
		leftLines = append(leftLines, "")
	}

	leftContent := strings.Join(leftLines[:innerH], "\n")
	leftStyled := lipgloss.NewStyle().Width(leftW).Render(leftContent)

	// ── Right panel: chat ──────────────────────────────────────
	chatPanel := renderChatPanel(msgs, chatInput, chatMode, chatW, innerH)

	// ── Bottom bar ─────────────────────────────────────────────
	bottomBar := " " +
		FormatKeyHint("t", "Throne") + "  " +
		FormatKeyHint("g", "Galaxy") + "  " +
		FormatKeyHint("c", "Chat") + "  " +
		FormatKeyHint("q", "Disconnect")

	// ── Assemble in outer frame ────────────────────────────────
	middle := SideBySide(leftStyled, chatPanel)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2)

	inner := topBar + "\n" + middle + "\n" + bottomBar
	return borderStyle.Render(inner)
}
