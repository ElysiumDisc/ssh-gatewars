package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/chat"
)

// ChatModel holds chat rendering state.
type ChatModel struct{}

func NewChatModel() ChatModel {
	return ChatModel{}
}

// renderChatPanel draws the chat panel with rounded border and themed colors.
func renderChatPanel(msgs []chat.ChatMessage, input string, chatMode bool, w, h int) string {
	if w < 12 {
		w = 12
	}
	if h < 4 {
		h = 4
	}

	innerW := w - 4 // border + padding

	// Header
	headerText := " COMMS "
	dashLen := innerW - len(headerText)
	if dashLen < 0 {
		dashLen = 0
	}
	halfDash := dashLen / 2
	header := StyleCyan.Render(strings.Repeat("─", halfDash) + headerText + strings.Repeat("─", dashLen-halfDash))

	// Messages area: total height minus header(1) + separator(1) + input(1) + border(2)
	msgHeight := h - 7
	if msgHeight < 1 {
		msgHeight = 1
	}

	// Style and wrap each message — wrap PLAIN text first, then apply color
	var msgLines []string
	for _, msg := range msgs {
		switch msg.Kind {
		case chat.MsgSystem:
			for _, line := range wrapText(msg.Body, innerW) {
				msgLines = append(msgLines, StyleGold.Render(line))
			}
		case chat.MsgAnnounce:
			style := lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
			for _, line := range wrapText(msg.Body, innerW) {
				msgLines = append(msgLines, style.Render(line))
			}
		case chat.MsgEmote:
			for _, line := range wrapText(msg.Body, innerW) {
				msgLines = append(msgLines, StyleDim.Render(line))
			}
		case chat.MsgWhisper:
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("#CC88FF"))
			text := fmt.Sprintf("[%s] %s", msg.SenderCallsign, msg.Body)
			for _, line := range wrapText(text, innerW) {
				msgLines = append(msgLines, style.Render(line))
			}
		default:
			// Normal chat: callsign in cyan, body in bright
			prefix := StyleCyan.Render(truncate(msg.SenderCallsign, 12)) + StyleDim.Render(": ")
			prefixLen := len(truncate(msg.SenderCallsign, 12)) + 2
			bodyW := innerW - prefixLen
			if bodyW < 10 {
				bodyW = 10
			}
			wrapped := wrapText(msg.Body, bodyW)
			for i, line := range wrapped {
				if i == 0 {
					msgLines = append(msgLines, prefix+StyleBright.Render(line))
				} else {
					msgLines = append(msgLines, pad(prefixLen)+StyleBright.Render(line))
				}
			}
		}
	}

	// Show most recent messages that fit
	start := len(msgLines) - msgHeight
	if start < 0 {
		start = 0
	}
	visible := msgLines[start:]

	// Pad to fill height (empty lines at top)
	for len(visible) < msgHeight {
		visible = append([]string{""}, visible...)
	}
	if len(visible) > msgHeight {
		visible = visible[len(visible)-msgHeight:]
	}

	// Separator
	sep := StyleCyanDim.Render(strings.Repeat("─", innerW))

	// Input line
	var inputLine string
	if chatMode {
		cursor := StyleCyan.Render("█")
		inputLine = StyleCyan.Render("> ") + StyleBright.Render(truncate(input, innerW-4)) + cursor
	} else {
		inputLine = FormatKeyHint("c", "chat")
	}

	// Assemble content for rounded box
	var content []string
	content = append(content, header)
	content = append(content, visible...)
	content = append(content, sep)
	content = append(content, inputLine)

	inner := strings.Join(content, "\n")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyanDim).
		Width(w - 2).
		Height(h - 2).
		Render(inner)
}
