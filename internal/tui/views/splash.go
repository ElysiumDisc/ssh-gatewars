package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SplashModel holds animation state for the splash screen.
type SplashModel struct {
	Frame int
}

func NewSplashModel() SplashModel {
	return SplashModel{}
}

func (s *SplashModel) Tick() {
	s.Frame++
}

// RenderSplash draws the animated splash screen — Ancient system boot feel.
func RenderSplash(s SplashModel, w, h int) string {
	// ── ASCII title art ────────────────────────────────────────

	titleLines := []string{
		"   ___   _   _____ ___  _    _   _   ___  ___ ",
		"  / __| /_\\ |_   _| __|| |  | | | | /_\\ | _ \\/ __|",
		" | (_ |/ _ \\  | | | _| | |/\\| | |_| / _ \\|   /\\__ \\",
		"  \\___/_/ \\_\\ |_| |___||__/\\__|\\___/_/ \\_\\_|_\\|___/",
	}
	titleW := 54 // widest line

	// Border width for the box
	boxW := titleW + 6

	// ── Phase calculations (15fps tick) ────────────────────────

	// Phase 1: Title typewriter (frames 0-30, 2 frames per line)
	titleVisible := s.Frame / 3
	if titleVisible > len(titleLines) {
		titleVisible = len(titleLines)
	}

	// Phase 2: Subtitle (frames 20+)
	showSubtitle := s.Frame > 20

	// Phase 3: Tagline (frames 30+)
	showTagline := s.Frame > 30

	// Phase 4: Progress bar (frames 35-55)
	showProgress := s.Frame > 35
	progressPct := float64(s.Frame-35) / 20.0
	if progressPct > 1 {
		progressPct = 1
	}

	// Phase 5: Prompt pulse (frames 55+)
	showPrompt := s.Frame > 55

	// ── Build content ──────────────────────────────────────────

	var lines []string

	// Top border
	lines = append(lines, StyleCyan.Render("╔"+strings.Repeat("═", boxW-2)+"╗"))

	// Empty spacer
	lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))

	// Title lines
	for i := 0; i < len(titleLines); i++ {
		if i < titleVisible {
			// Typewriter: reveal chars up to a point
			charReveal := (s.Frame - i*3) * 4
			line := titleLines[i]
			runes := []rune(line)
			if charReveal < len(runes) {
				line = string(runes[:charReveal])
			}
			styled := StyleGold.Render(line)
			lineW := lipgloss.Width(styled)
			rightPad := boxW - 2 - lineW
			if rightPad < 0 {
				rightPad = 0
			}
			lines = append(lines, StyleCyan.Render("║")+
				" "+styled+pad(rightPad-1)+
				StyleCyan.Render("║"))
		} else {
			lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))
		}
	}

	// Empty spacer
	lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))

	// Subtitle
	if showSubtitle {
		sub := "ANCIENT DEFENSE NETWORK — v3.0"
		styled := StyleSubtitle.Render(sub)
		lineW := lipgloss.Width(styled)
		leftPad := (boxW - 2 - lineW) / 2
		rightPad := boxW - 2 - lineW - leftPad
		lines = append(lines, StyleCyan.Render("║")+pad(leftPad)+styled+pad(rightPad)+StyleCyan.Render("║"))
	} else {
		lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))
	}

	// Tagline
	if showTagline {
		tag := "The replicators are coming."
		styled := StyleGold.Render(tag)
		lineW := lipgloss.Width(styled)
		leftPad := (boxW - 2 - lineW) / 2
		rightPad := boxW - 2 - lineW - leftPad
		lines = append(lines, StyleCyan.Render("║")+pad(leftPad)+styled+pad(rightPad)+StyleCyan.Render("║"))
	} else {
		lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))
	}

	// Empty spacer
	lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))

	// Progress bar
	if showProgress {
		barW := 30
		bar := ProgressBar(progressPct, barW, ColorCyan, ColorDim)
		label := "INITIALIZING..."
		if progressPct >= 1.0 {
			label = "SYSTEMS ONLINE  "
		}
		styled := StyleDim.Render(label) + " " + bar
		lineW := lipgloss.Width(styled)
		leftPad := (boxW - 2 - lineW) / 2
		rightPad := boxW - 2 - lineW - leftPad
		if rightPad < 0 {
			rightPad = 0
		}
		lines = append(lines, StyleCyan.Render("║")+pad(leftPad)+styled+pad(rightPad)+StyleCyan.Render("║"))
	} else {
		lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))
	}

	// Empty spacer
	lines = append(lines, StyleCyan.Render("║")+pad(boxW-2)+StyleCyan.Render("║"))

	// Bottom border
	lines = append(lines, StyleCyan.Render("╚"+strings.Repeat("═", boxW-2)+"╝"))

	// Prompt below box
	if showPrompt {
		pulse := (s.Frame / 8) % 2
		var prompt string
		if pulse == 0 {
			prompt = StyleCyan.Render("Press any key to begin...")
		} else {
			prompt = StyleCyanDim.Render("Press any key to begin...")
		}
		lines = append(lines, "")
		lines = append(lines, CenterH(prompt, boxW))
	}

	content := strings.Join(lines, "\n")
	return Center(content, w, h)
}
