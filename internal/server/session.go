package server

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/player"
	"ssh-gatewars/internal/render"
	"ssh-gatewars/internal/simulation"
)

type sessionState int

const (
	stateFactionSelect sessionState = iota
	stateBattlefield
	stateScoreboard
	stateStats
	stateNetworkMap
)

type frameTickMsg time.Time

// Model is the per-session Bubbletea model.
type Model struct {
	engine   *simulation.Engine
	renderer *lipgloss.Renderer
	store    *player.Store

	state    sessionState
	faction  int
	sshKey   string
	sshUser  string
	width    int
	height   int

	viewport  *render.Viewport
	frame     *render.FrameBuilder
	snap      simulation.Snapshot
	starfield *render.Starfield

	registered   bool
	showHelp     bool
	viewOnly     bool // multiplex view-only session
	lastActivity time.Time

	// Brief feedback messages
	flashMsg    string
	flashExpiry time.Time
}

// factionFromUsername maps SSH usernames to faction IDs.
var factionFromUsername = map[string]int{
	"tauri":  faction.Tauri,
	"tau'ri": faction.Tauri,
	"goauld": faction.Goauld,
	"jaffa":  faction.Jaffa,
	"lucian": faction.Lucian,
	"asgard": faction.Asgard,
}

// viewFromCommand maps SSH command strings to session states.
var viewFromCommand = map[string]sessionState{
	"scoreboard": stateScoreboard,
	"network":    stateNetworkMap,
	"stats":      stateStats,
}

// NewModel creates a session model.
func NewModel(engine *simulation.Engine, starfield *render.Starfield, renderer *lipgloss.Renderer, sshKey string, store *player.Store, sshUser string, sshCmd []string) Model {
	m := Model{
		engine:       engine,
		renderer:     renderer,
		starfield:    starfield,
		sshKey:       sshKey,
		sshUser:      sshUser,
		store:        store,
		state:        stateFactionSelect,
		faction:      -1,
		width:        80,
		height:       24,
		lastActivity: time.Now(),
	}

	// Check for multiplex view command (e.g. "ssh sgc.games scoreboard")
	if len(sshCmd) > 0 {
		cmd := strings.ToLower(sshCmd[0])
		if viewState, ok := viewFromCommand[cmd]; ok {
			m.state = viewState
			m.viewOnly = true
		}
	}

	// Check for faction-as-username (e.g. "ssh tauri@sgc.games")
	if f, ok := factionFromUsername[strings.ToLower(sshUser)]; ok && !m.viewOnly {
		m.faction = f
		m.state = stateBattlefield
	}

	// Check for returning player in database
	if m.faction < 0 && !m.viewOnly && store != nil && !strings.HasPrefix(sshKey, "anon:") {
		if info, err := store.GetPlayer(sshKey); err == nil && info != nil {
			m.faction = info.Faction
			m.state = stateBattlefield
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Start render tick
	cmds = append(cmds, tea.Tick(time.Second/15, func(t time.Time) tea.Msg {
		return frameTickMsg(t)
	}))

	// Auto-register if faction was set via username/DB
	if m.faction >= 0 && m.state == stateBattlefield {
		cmds = append(cmds, func() tea.Msg {
			m.engine.RegisterPlayer(m.sshKey, m.faction)
			if m.store != nil && !strings.HasPrefix(m.sshKey, "anon:") {
				m.store.SavePlayer(m.sshKey, m.faction)
			}
			return nil
		})
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildViewport()
		return m, nil

	case tea.KeyMsg:
		m.lastActivity = time.Now()

		switch m.state {
		case stateFactionSelect:
			return m.updateFactionSelect(msg)
		case stateBattlefield:
			return m.updateBattlefield(msg)
		case stateScoreboard, stateStats, stateNetworkMap:
			return m.updateAltView(msg)
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

	// Save to DB
	if m.store != nil && !strings.HasPrefix(m.sshKey, "anon:") {
		m.store.SavePlayer(m.sshKey, f)
	}

	return m, nil
}

func (m Model) updateBattlefield(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.viewOnly {
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.state = stateScoreboard
		}
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case " ": // Space — activate faction power
		if m.faction >= 0 {
			success, notification := m.engine.Powers.TryActivate(m.faction, m.sshKey)
			if success {
				m.engine.AddNotification(m.faction, notification)
				m.flashMsg = "POWER ACTIVATED!"
				m.flashExpiry = time.Now().Add(2 * time.Second)
			} else {
				status := m.engine.Powers.Status(m.faction)
				if status.State == simulation.PowerActive {
					m.flashMsg = fmt.Sprintf("%s is active!", status.Name)
				} else {
					m.flashMsg = fmt.Sprintf("%s on cooldown (%.0fs)", status.Name, status.Remaining.Seconds())
				}
				m.flashExpiry = time.Now().Add(2 * time.Second)
			}
		}

	case "1", "2", "3", "4", "5": // Sector focus
		sector := int(msg.String()[0] - '1')
		m.engine.SetFocusSector(m.sshKey, sector)
		targetName := faction.Factions[sector].ShortName
		m.flashMsg = fmt.Sprintf("Focusing sector %d (%s)", sector+1, targetName)
		m.flashExpiry = time.Now().Add(2 * time.Second)

	case "tab": // Cycle views
		m.state = stateScoreboard
		m.showHelp = false

	case "?": // Toggle help
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m Model) updateAltView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab":
		switch m.state {
		case stateScoreboard:
			m.state = stateNetworkMap
		case stateNetworkMap:
			m.state = stateStats
		case stateStats:
			m.state = stateBattlefield
		}
	case "?":
		m.showHelp = !m.showHelp
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
	case stateScoreboard:
		return m.viewScoreboard()
	case stateNetworkMap:
		return m.viewNetworkMap()
	case stateStats:
		return m.viewStats()
	default:
		return ""
	}
}

func (m Model) viewFactionSelect() string {
	var sb strings.Builder

	titleStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#40E0D0"))
	headerStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	topPad := (m.height - 18) / 2
	if topPad < 0 {
		topPad = 0
	}
	for i := 0; i < topPad; i++ {
		sb.WriteByte('\n')
	}

	center := func(s string) string {
		w := lipgloss.Width(s)
		pad := (m.width - w) / 2
		if pad < 0 {
			pad = 0
		}
		return strings.Repeat(" ", pad) + s
	}

	sb.WriteString(center(titleStyle.Render("SGC TACTICAL NETWORK v0.5")) + "\n")
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
	sb.WriteString(center(dimStyle.Render("Press 1-5 to choose your faction.")) + "\n")
	sb.WriteString(center(dimStyle.Render("Or: ssh <faction>@host to join directly.")) + "\n\n")
	sb.WriteString(center(dimStyle.Render("\"Chevron 7 locked.\" — Sgt. Walter Harriman")) + "\n")

	return sb.String()
}

func (m Model) viewBattlefield() string {
	if m.frame == nil || m.viewport == nil {
		return "Initializing..."
	}

	if m.width < 60 || m.height < 15 {
		style := m.renderer.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
		return style.Render(fmt.Sprintf("Terminal too small (%dx%d). Need at least 60x15.", m.width, m.height))
	}

	battlefield := m.frame.Build(m.snap)
	hud := render.BuildHUD(m.faction, m.snap, m.width, m.renderer)

	result := battlefield + "\n" + hud

	// Overlay notifications
	for _, n := range m.snap.Notifications {
		if n.Faction == m.faction {
			notifStyle := m.renderer.NewStyle().
				Foreground(lipgloss.Color(faction.Factions[m.faction].ColorFG)).
				Bold(true)
			notif := notifStyle.Render("  " + n.Message + "  ")
			result = notif + "\n" + result
		}
	}

	// Flash message
	if m.flashMsg != "" && time.Now().Before(m.flashExpiry) {
		flashStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true)
		result = flashStyle.Render("  "+m.flashMsg+"  ") + "\n" + result
	}

	// Help overlay
	if m.showHelp {
		result = m.renderHelp() + "\n" + result
	}

	return result
}

func (m Model) viewScoreboard() string {
	var sb strings.Builder

	titleStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  FACTION SCOREBOARD") + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────────────") + "\n\n")

	// Sort factions by territory
	type entry struct {
		id   int
		terr float64
	}
	entries := make([]entry, faction.Count)
	for i := 0; i < faction.Count; i++ {
		t := 20.0
		if m.snap.Territory != nil {
			t = m.snap.Territory.Percents[i]
		}
		entries[i] = entry{i, t}
	}
	// Simple sort by territory descending
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].terr > entries[i].terr {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	medals := []string{"1st", "2nd", "3rd", "4th", "5th"}
	for rank, e := range entries {
		f := faction.Factions[e.id]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)

		barWidth := 30
		filled := int(e.terr / 100 * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		bar := fStyle.Render(strings.Repeat("█", filled)) +
			dimStyle.Render(strings.Repeat("░", barWidth-filled))

		sb.WriteString(fmt.Sprintf("  %s  ", medals[rank]))
		sb.WriteString(fStyle.Render(fmt.Sprintf("%-18s", f.Name)))
		sb.WriteString(fmt.Sprintf(" %5.1f%%  ", e.terr))
		sb.WriteString(bar)
		sb.WriteString(dimStyle.Render(fmt.Sprintf("  %d ships  %d online",
			m.snap.ShipCounts[e.id], m.snap.PlayerCounts[e.id])))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  Kills / Deaths per faction:") + "\n")
	for i := 0; i < faction.Count; i++ {
		f := faction.Factions[i]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG))
		sb.WriteString(fmt.Sprintf("    %s  K:%d  D:%d\n",
			fStyle.Render(fmt.Sprintf("%-12s", f.ShortName)),
			m.snap.KillCounts[i], m.snap.DeathCounts[i]))
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  [Tab] Next view  |  [q] Quit") + "\n")

	return sb.String()
}

func (m Model) viewNetworkMap() string {
	var sb strings.Builder

	titleStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  STARGATE NETWORK MAP") + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────────────") + "\n\n")

	// Show a text-based territory overview (8x4 grid summary)
	gridW, gridH := 8, 4
	cellsPerGridX := simulation.TerritoryGridW / gridW
	cellsPerGridY := simulation.TerritoryGridH / gridH

	for gy := 0; gy < gridH; gy++ {
		sb.WriteString("    ")
		for gx := 0; gx < gridW; gx++ {
			// Count faction dominance in this region
			var counts [faction.Count]int
			for ty := gy * cellsPerGridY; ty < (gy+1)*cellsPerGridY; ty++ {
				for tx := gx * cellsPerGridX; tx < (gx+1)*cellsPerGridX; tx++ {
					if m.snap.Territory != nil {
						owner := m.snap.Territory.Zones[tx][ty]
						if owner >= 0 && owner < faction.Count {
							counts[owner]++
						}
					}
				}
			}
			best := -1
			bestCount := 0
			for f := 0; f < faction.Count; f++ {
				if counts[f] > bestCount {
					bestCount = counts[f]
					best = f
				}
			}
			if best >= 0 {
				fStyle := m.renderer.NewStyle().
					Foreground(lipgloss.Color(faction.Factions[best].ColorFG)).
					Background(lipgloss.Color(faction.Factions[best].ColorBG))
				sb.WriteString(fStyle.Render("██████"))
			} else {
				sb.WriteString(dimStyle.Render("░░░░░░"))
			}
			sb.WriteString(" ")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  Gate locations:") + "\n")
	gateLabels := []string{"lower-right", "left", "right", "lower-left", "top"}
	for i := 0; i < faction.Count; i++ {
		f := faction.Factions[i]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)
		gate := simulation.GatePositions[i]
		sb.WriteString(fmt.Sprintf("    %s  (%3.0f, %3.0f) %s\n",
			fStyle.Render(fmt.Sprintf("%-18s", f.Name)),
			gate.X, gate.Y,
			dimStyle.Render(gateLabels[i])))
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  [Tab] Next view  |  [q] Quit") + "\n")

	return sb.String()
}

func (m Model) viewStats() string {
	var sb strings.Builder

	titleStyle := m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#888888"))

	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render("  YOUR STATS") + "\n")
	sb.WriteString(dimStyle.Render("  ─────────────────────────────────────────────────────") + "\n\n")

	if m.faction >= 0 {
		f := faction.Factions[m.faction]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)
		sb.WriteString("  Faction: " + fStyle.Render(f.Name) + "\n")
	}

	keyDisplay := m.sshKey
	if strings.HasPrefix(keyDisplay, "anon:") {
		keyDisplay = "Anonymous (no SSH key)"
	}
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  SSH Key: %s", keyDisplay)) + "\n")

	sessionType := "Primary"
	if m.viewOnly {
		sessionType = "View-Only"
	}
	sb.WriteString(dimStyle.Render(fmt.Sprintf("  Session: %s", sessionType)) + "\n\n")

	// Power status
	if m.faction >= 0 {
		ps := m.snap.PowerStatuses[m.faction]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(faction.Factions[m.faction].ColorFG)).Bold(true)

		sb.WriteString("  Power: " + fStyle.Render(ps.Name) + "\n")
		switch ps.State {
		case simulation.PowerReady:
			sb.WriteString("  Status: " + m.renderer.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("READY") + "\n")
		case simulation.PowerActive:
			sb.WriteString(fmt.Sprintf("  Status: ACTIVE (%.1fs remaining)\n", ps.Remaining.Seconds()))
		case simulation.PowerCooldown:
			sb.WriteString(fmt.Sprintf("  Status: Cooldown (%.0fs)\n", ps.Remaining.Seconds()))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(dimStyle.Render("  [Tab] Battlefield  |  [q] Quit") + "\n")

	return sb.String()
}

func (m Model) renderHelp() string {
	var sb strings.Builder

	borderStyle := m.renderer.NewStyle().
		Foreground(lipgloss.Color("#40E0D0")).
		Bold(true)
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

	sb.WriteString(borderStyle.Render("  ┌─ SSH GATEWARS ── CONTROLS ─────────────────────┐") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("                                                 ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  SPACE   Activate faction power                 ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  1-5     Focus spawns toward a sector           ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  TAB     Cycle views                            ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  ?       Toggle this help                       ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  q       Disconnect                             ") + borderStyle.Render("│") + "\n")

	if m.faction >= 0 {
		f := faction.Factions[m.faction]
		ps := m.snap.PowerStatuses[m.faction]
		fStyle := m.renderer.NewStyle().Foreground(lipgloss.Color(f.ColorFG)).Bold(true)

		sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("                                                 ") + borderStyle.Render("│") + "\n")
		sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  YOUR POWER: ") + fStyle.Render(fmt.Sprintf("%-34s", ps.Name)) + borderStyle.Render("│") + "\n")
	}

	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("                                                 ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  │") + dimStyle.Render("  Press ? to close                               ") + borderStyle.Render("│") + "\n")
	sb.WriteString(borderStyle.Render("  └─────────────────────────────────────────────────┘") + "\n")

	return sb.String()
}

// IsViewOnly returns whether this is a view-only multiplex session.
func (m Model) IsViewOnly() bool {
	return m.viewOnly
}

// LastActivity returns the time of the last user input.
func (m Model) LastActivity() time.Time {
	return m.lastActivity
}

// Cleanup should be called when the session ends.
func (m Model) Cleanup() {
	if m.sshKey != "" && m.faction >= 0 {
		m.engine.UnregisterPlayer(m.sshKey)
	}
}
