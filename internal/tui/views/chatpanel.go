package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/chat"
)

// ChatPanelState describes the visibility of the chat panel.
type ChatPanelState int

const (
	ChatHidden   ChatPanelState = iota // no chat visible
	ChatCompact                         // bottom 6 rows
	ChatExpanded                        // bottom half
)

// RenderChatPanel renders the chat overlay.
func RenderChatPanel(
	messages []chat.ChatMessage,
	activeChannel string,
	inputText string,
	state ChatPanelState,
	width, rows int,
) string {
	if state == ChatHidden || rows <= 0 {
		return ""
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44AAFF")).Bold(true)
	systemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDAA00"))
	emoteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AA88CC"))
	whisperStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DD88DD"))
	announceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	channelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666688"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	backlogStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))

	var b strings.Builder

	// Separator line
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	// Message area (rows - 2 for separator and input line)
	msgRows := rows - 2
	if msgRows < 1 {
		msgRows = 1
	}

	// Take last N messages
	start := 0
	if len(messages) > msgRows {
		start = len(messages) - msgRows
	}
	visible := messages[start:]

	for _, msg := range visible {
		ts := timeStyle.Render(fmt.Sprintf("[%02d:%02d]", msg.Timestamp.Hour(), msg.Timestamp.Minute()))
		chTag := channelStyle.Render("[" + chat.ChannelDisplayName(msg.Channel) + "]")

		var line string
		switch msg.Kind {
		case chat.MsgSystem:
			line = fmt.Sprintf("%s %s %s", ts, chTag, systemStyle.Render(msg.Body))
		case chat.MsgEmote:
			line = fmt.Sprintf("%s %s %s", ts, chTag, emoteStyle.Render(msg.Body))
		case chat.MsgWhisper:
			line = fmt.Sprintf("%s %s %s <%s> %s", ts,
				whisperStyle.Render("[DM]"),
				chTag,
				nameStyle.Render(msg.SenderCallsign),
				msg.Body)
		case chat.MsgAnnounce:
			line = fmt.Sprintf("%s %s", ts, announceStyle.Render("⚠ "+msg.Body))
		default: // MsgChat
			line = fmt.Sprintf("%s %s <%s> %s", ts, chTag,
				nameStyle.Render(msg.SenderCallsign),
				msg.Body)
		}

		if msg.IsBacklog {
			line = backlogStyle.Render(line)
		}

		// Truncate to width
		if lipgloss.Width(line) > width {
			line = line[:width]
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Pad empty rows
	for i := len(visible); i < msgRows; i++ {
		b.WriteString("\n")
	}

	// Input line
	prompt := channelStyle.Render("["+chat.ChannelDisplayName(activeChannel)+"]") + " > "
	b.WriteString(inputStyle.Render(prompt + inputText + "_"))

	return b.String()
}

// RenderToast renders a single toast notification line.
func RenderToast(msg chat.ChatMessage, width int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	chTag := chat.ChannelDisplayName(msg.Channel)

	var text string
	if msg.Kind == chat.MsgSystem {
		text = fmt.Sprintf("[%s] %s", chTag, msg.Body)
	} else {
		text = fmt.Sprintf("[%s] <%s> %s", chTag, msg.SenderCallsign, msg.Body)
	}

	if lipgloss.Width(text) > width {
		text = text[:width-3] + "..."
	}

	return style.Render(text)
}

// RenderPlayerList renders the online player list modal.
func RenderPlayerList(players []PlayerListEntry, width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#44AAFF")).Bold(true)
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  ONLINE PLAYERS"))
	b.WriteString("\n\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf("  %-16s  %-5s  %-20s", "Callsign", "Level", "Location")))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 45)))
	b.WriteString("\n")

	for _, p := range players {
		b.WriteString(rowStyle.Render(fmt.Sprintf("  %-16s  Lv%-3d  %-20s", p.Callsign, p.Level, p.Location)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Press Tab or Esc to close"))
	b.WriteString("\n")

	return b.String()
}

// PlayerListEntry holds data for one player in the list.
type PlayerListEntry struct {
	Callsign string
	Level    int
	Location string
}
