package server

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/render"
	"ssh-gatewars/internal/simulation"
)

type sessionState int

const (
	stateFactionSelect sessionState = iota
	stateBattlefield
)

type frameTickMsg time.Time

// Model is the per-session Bubbletea model.
type Model struct {
	engine   *simulation.Engine
	renderer *lipgloss.Renderer

	state    sessionState
	faction  int
	sshKey   string
	width    int
	height   int

	viewport *render.Viewport
	frame    *render.FrameBuilder
	snap     simulation.Snapshot

	starfield *render.Starfield
	registered bool
}

// NewModel creates a session model.
func NewModel(engine *simulation.Engine, starfield *render.Starfield, renderer *lipgloss.Renderer, sshKey string) Model {
	return Model{
		engine:    engine,
		renderer:  renderer,
		starfield: starfield,
		sshKey:    sshKey,
		state:     stateFactionSelect,
		faction:   -1,
		width:     80,
		height:    24,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second/15, func(t time.Time) tea.Msg {
		return frameTickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildViewport()
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateFactionSelect:
			return m.updateFactionSelect(msg)
		case stateBattlefield:
			return m.updateBattlefield(msg)
		}

	case frameTickMsg:
		m.snap = m.engine.Snapshot()
		return m, tea.Tick(time.Second/15, func(t time.Time) tea.Msg {
			return frameTickMsg(t)
		})
	}

	return m, nil
}

func (m Model) updateFactionSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1":
		return m.joinFaction(faction.Tauri)
	case "2":
		return m.joinFaction(faction.Goauld)
	case "3":
		return m.joinFaction(faction.Jaffa)
	case "4":
		return m.joinFaction(faction.Lucian)
	case "5":
		return m.joinFaction(faction.Asgard)
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) joinFaction(f int) (tea.Model, tea.Cmd) {
	m.faction = f
	m.state = stateBattlefield
	m.registered = m.engine.RegisterPlayer(m.sshKey, f)
	m.rebuildViewport()
	return m, nil
}

func (m Model) updateBattlefield(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) rebuildViewport() {
	m.viewport = render.NewViewport(m.width, m.height, simulation.WorldW, simulation.WorldH, render.HUDRows)
	m.frame = &render.FrameBuilder{
		Viewport:  m.viewport,
		Starfield: m.starfield,
		Renderer:  m.renderer,
	}
}

func (m Model) View() string {
	switch m.state {
	case stateFactionSelect:
		return m.viewFactionSelect()
	case stateBattlefield:
		return m.viewBattlefield()
	default:
		return ""
	}
}

func (m Model) viewFactionSelect() string {
	var sb strings.Builder

	titleStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#40E0D0"))
	headerStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	// Center vertically
	topPad := (m.height - 18) / 2
	if topPad < 0 {
		topPad = 0
	}
	for i := 0; i < topPad; i++ {
		sb.WriteByte('\n')
	}

	// Center helper
	center := func(s string) string {
		w := lipgloss.Width(s)
		pad := (m.width - w) / 2
		if pad < 0 {
			pad = 0
		}
		return strings.Repeat(" ", pad) + s
	}

	sb.WriteString(center(titleStyle.Render("SGC TACTICAL NETWORK v0.1")) + "\n")
	sb.WriteString(center(dimStyle.Render("Developed by Lt. Col. Carter & SGC Technical Staff")) + "\n\n")
	sb.WriteString(center(headerStyle.Render("CHOOSE YOUR ALLEGIANCE")) + "\n\n")

	snap := m.snap
	for i := 0; i < faction.Count; i++ {
		f := faction.Factions[i]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)
		players := snap.PlayerCounts[i]
		territory := 20.0
		if snap.Territory != nil {
			territory = snap.Territory.Percents[i]
		}

		// Territory bar
		barWidth := 10
		filled := int(territory / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		bar := fStyle.Render(strings.Repeat("█", filled)) +
			dimStyle.Render(strings.Repeat("░", barWidth-filled))

		line := fmt.Sprintf("  [%d] ", i+1) +
			fStyle.Render(fmt.Sprintf("%-18s", f.Name)) +
			dimStyle.Render(fmt.Sprintf(" %2d online  ", players)) +
			bar +
			dimStyle.Render(fmt.Sprintf(" %4.0f%%", territory))

		sb.WriteString(center(line) + "\n")
	}

	sb.WriteString("\n")
	if m.sshKey != "" {
		sb.WriteString(center(dimStyle.Render("Your SSH key: "+m.sshKey[:min(20, len(m.sshKey))]+"...")) + "\n")
	}
	sb.WriteString(center(dimStyle.Render("Press 1-5 to choose your faction.")) + "\n\n")
	sb.WriteString(center(dimStyle.Render("\"Chevron 7 locked.\" — Sgt. Walter Harriman")) + "\n")

	return sb.String()
}

func (m Model) viewBattlefield() string {
	if m.frame == nil || m.viewport == nil {
		return "Initializing..."
	}

	// Check minimum terminal size
	if m.width < 60 || m.height < 15 {
		style := m.renderer.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
		return style.Render(fmt.Sprintf("Terminal too small (%dx%d). Need at least 60x15.", m.width, m.height))
	}

	battlefield := m.frame.Build(m.snap)
	hud := render.BuildHUD(m.faction, m.snap, m.width, m.renderer)

	return battlefield + "\n" + hud
}

// Cleanup should be called when the session ends.
func (m Model) Cleanup() {
	if m.sshKey != "" && m.faction >= 0 {
		m.engine.UnregisterPlayer(m.sshKey)
	}
}
