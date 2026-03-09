package server

import (
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
	stateGalaxyMap
	stateSystemView
	stateColonyManage
	stateTechTree
	stateDiplomacy
	stateScoreboard
	stateHelp
	stateShipDesigner
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

	snap simulation.Snapshot

	registered   bool
	viewOnly     bool
	lastActivity time.Time

	flashMsg    string
	flashExpiry time.Time

	selectedSystem int
	selectedSlider int
	selectedTree   int

	localSliders   [5]int
	localTechAlloc [6]int
}

var factionFromUsername = map[string]int{
	"tauri":   faction.Tauri,
	"tau'ri":  faction.Tauri,
	"goauld":  faction.Goauld,
	"goa'uld": faction.Goauld,
	"asgard":  faction.Asgard,
	"tokra":   faction.Tokra,
	"tok'ra":  faction.Tokra,
	"jaffa":   faction.Jaffa,
}

var viewFromCommand = map[string]sessionState{
	"scoreboard": stateScoreboard,
	"galaxy":     stateGalaxyMap,
}

// NewModel creates a session model.
func NewModel(engine *simulation.Engine, renderer *lipgloss.Renderer, sshKey string, store *player.Store, sshUser string, sshCmd []string) Model {
	m := Model{
		engine:       engine,
		renderer:     renderer,
		sshKey:       sshKey,
		sshUser:      sshUser,
		store:        store,
		state:        stateFactionSelect,
		faction:      -1,
		width:        80,
		height:       24,
		lastActivity: time.Now(),
	}

	if len(sshCmd) > 0 {
		cmd := strings.ToLower(sshCmd[0])
		if viewState, ok := viewFromCommand[cmd]; ok {
			m.state = viewState
			m.viewOnly = true
		}
	}

	if f, ok := factionFromUsername[strings.ToLower(sshUser)]; ok && !m.viewOnly {
		m.faction = f
		m.state = stateGalaxyMap
	}

	if m.faction < 0 && !m.viewOnly && store != nil && !strings.HasPrefix(sshKey, "anon:") {
		if info, err := store.GetPlayer(sshKey); err == nil && info != nil {
			m.faction = info.Faction
			m.state = stateGalaxyMap
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	cmds = append(cmds, tea.Tick(time.Second/15, func(t time.Time) tea.Msg {
		return frameTickMsg(t)
	}))

	if m.faction >= 0 && m.state == stateGalaxyMap {
		cmds = append(cmds, func() tea.Msg {
			m.engine.RegisterPlayer(m.sshKey, m.faction)
			if m.store != nil && !strings.HasPrefix(m.sshKey, "anon:") {
				_ = m.store.SavePlayer(m.sshKey, m.faction)
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
		return m, nil

	case tea.KeyMsg:
		m.lastActivity = time.Now()

		switch m.state {
		case stateFactionSelect:
			return m.updateFactionSelect(msg)
		case stateGalaxyMap:
			return m.updateGalaxyMap(msg)
		case stateSystemView:
			return m.updateSystemView(msg)
		case stateColonyManage:
			return m.updateColonyManage(msg)
		case stateTechTree:
			return m.updateTechTree(msg)
		case stateDiplomacy:
			return m.updateDiplomacy(msg)
		case stateScoreboard:
			return m.updateScoreboard(msg)
		case stateHelp:
			return m.updateHelp(msg)
		case stateShipDesigner:
			return m.updateShipDesigner(msg)
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
		return m.joinFaction(faction.Asgard)
	case "4":
		return m.joinFaction(faction.Tokra)
	case "5":
		return m.joinFaction(faction.Jaffa)
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) joinFaction(f int) (tea.Model, tea.Cmd) {
	m.faction = f
	m.state = stateGalaxyMap
	m.registered = m.engine.RegisterPlayer(m.sshKey, f)
	if m.store != nil && !strings.HasPrefix(m.sshKey, "anon:") {
		_ = m.store.SavePlayer(m.sshKey, f)
	}
	return m, nil
}

func (m Model) updateGalaxyMap(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "right", "l":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "right")
	case "left", "h":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "left")
	case "up", "k":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "up")
	case "down", "j":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "down")
	case "enter":
		m.state = stateSystemView
	case "tab":
		m.state = stateScoreboard
	case "t":
		if m.faction >= 0 && m.faction < faction.Count {
			fs := m.snap.Factions[m.faction]
			copy(m.localTechAlloc[:], fs.TechAlloc[:])
			m.selectedTree = 0
		}
		m.state = stateTechTree
	case "d":
		m.state = stateDiplomacy
	case "s":
		m.state = stateShipDesigner
	case "?":
		m.state = stateHelp
	}
	return m, nil
}

func (m Model) updateSystemView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "esc":
		m.state = stateGalaxyMap
	case "right", "l":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "right")
	case "left", "h":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "left")
	case "up", "k":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "up")
	case "down", "j":
		m.selectedSystem = render.NavigateSystems(m.snap.Systems, m.selectedSystem, "down")
	case "enter":
		if col, ok := m.snap.Colonies[m.selectedSystem]; ok && col.Faction == m.faction {
			m.localSliders = [5]int{col.SliderShip, col.SliderDefense, col.SliderIndustry, col.SliderEcology, col.SliderResearch}
			m.selectedSlider = 0
			m.state = stateColonyManage
		}
	}
	return m, nil
}

func (m Model) updateColonyManage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "esc":
		m.state = stateSystemView
	case "up", "k":
		m.selectedSlider--
		if m.selectedSlider < 0 {
			m.selectedSlider = 4
		}
	case "down", "j":
		m.selectedSlider++
		if m.selectedSlider > 4 {
			m.selectedSlider = 0
		}
	case "right", "l":
		m.localSliders = adjustSliders5(m.localSliders, m.selectedSlider, 5)
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:     simulation.ActionSetSliders,
			PlayerKey: m.sshKey,
			Faction:  m.faction,
			SystemID: m.selectedSystem,
			Sliders:  m.localSliders,
		})
	case "left", "h":
		m.localSliders = adjustSliders5(m.localSliders, m.selectedSlider, -5)
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:     simulation.ActionSetSliders,
			PlayerKey: m.sshKey,
			Faction:  m.faction,
			SystemID: m.selectedSystem,
			Sliders:  m.localSliders,
		})
	}
	return m, nil
}

func (m Model) updateTechTree(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "esc":
		m.state = stateGalaxyMap
	case "up", "k":
		m.selectedTree--
		if m.selectedTree < 0 {
			m.selectedTree = 5
		}
	case "down", "j":
		m.selectedTree++
		if m.selectedTree > 5 {
			m.selectedTree = 0
		}
	case "right", "l":
		m.localTechAlloc = adjustSliders6(m.localTechAlloc, m.selectedTree, 5)
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:      simulation.ActionSetTechAlloc,
			PlayerKey: m.sshKey,
			Faction:   m.faction,
			TechAlloc: m.localTechAlloc,
		})
	case "left", "h":
		m.localTechAlloc = adjustSliders6(m.localTechAlloc, m.selectedTree, -5)
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:      simulation.ActionSetTechAlloc,
			PlayerKey: m.sshKey,
			Faction:   m.faction,
			TechAlloc: m.localTechAlloc,
		})
	}
	return m, nil
}

func (m Model) updateDiplomacy(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "esc":
		m.state = stateGalaxyMap
	}
	return m, nil
}

func (m Model) updateScoreboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "tab", "escape", "esc":
		m.state = stateGalaxyMap
	}
	return m, nil
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "?", "escape", "esc":
		m.state = stateGalaxyMap
	}
	return m, nil
}

func (m Model) updateShipDesigner(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "escape", "esc":
		m.state = stateGalaxyMap
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateFactionSelect:
		return render.BuildSplash(m.snap, m.width, m.height, m.renderer)
	case stateGalaxyMap:
		return m.viewGalaxyMap()
	case stateSystemView:
		return render.BuildSystemView(m.snap, m.selectedSystem, m.faction, m.width, m.height, m.renderer)
	case stateColonyManage:
		return m.viewColonyManage()
	case stateTechTree:
		return m.viewTechTree()
	case stateDiplomacy:
		return render.BuildDiplomacyView(m.snap, m.faction, m.width, m.height, m.renderer)
	case stateScoreboard:
		return render.BuildScoreboard(m.snap, m.faction, m.width, m.height, m.renderer)
	case stateHelp:
		return m.viewHelp()
	case stateShipDesigner:
		return render.BuildShipDesigner(m.snap, m.faction, m.width, m.height, m.renderer)
	default:
		return ""
	}
}

func (m Model) viewGalaxyMap() string {
	mapView := render.BuildGalaxyMap(m.snap, m.selectedSystem, m.faction, m.width, m.height, m.renderer)
	hud := render.BuildHUD(m.faction, m.snap, m.width, m.renderer)
	result := mapView + "\n" + hud

	if m.flashMsg != "" && time.Now().Before(m.flashExpiry) {
		flashStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true)
		result = flashStyle.Render("  "+m.flashMsg) + "\n" + result
	}

	for _, n := range m.snap.Notifications {
		notifStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#FF8800")).Bold(true)
		result = notifStyle.Render("  "+n.Message) + "\n" + result
	}

	return result
}

func (m Model) viewColonyManage() string {
	col, ok := m.snap.Colonies[m.selectedSystem]
	if !ok {
		return "  No colony here"
	}

	// Override with local slider values for responsive editing
	col.SliderShip = m.localSliders[0]
	col.SliderDefense = m.localSliders[1]
	col.SliderIndustry = m.localSliders[2]
	col.SliderEcology = m.localSliders[3]
	col.SliderResearch = m.localSliders[4]
	col.ShipOutput = col.TotalOutput * float64(col.SliderShip) / 100
	col.DefenseOutput = col.TotalOutput * float64(col.SliderDefense) / 100
	col.IndustryOutput = col.TotalOutput * float64(col.SliderIndustry) / 100
	col.EcologyOutput = col.TotalOutput * float64(col.SliderEcology) / 100
	col.ResearchOutput = col.TotalOutput * float64(col.SliderResearch) / 100

	modSnap := m.snap
	modCols := make(map[int]simulation.ColonySnapshot, len(m.snap.Colonies))
	for k, v := range m.snap.Colonies {
		modCols[k] = v
	}
	modCols[m.selectedSystem] = col
	modSnap.Colonies = modCols

	return render.BuildColonyView(modSnap, m.selectedSystem, m.faction, m.selectedSlider, m.width, m.height, m.renderer)
}

func (m Model) viewTechTree() string {
	modSnap := m.snap
	if m.faction >= 0 && m.faction < faction.Count {
		fs := modSnap.Factions[m.faction]
		copy(fs.TechAlloc[:], m.localTechAlloc[:])
		modSnap.Factions[m.faction] = fs
	}
	return render.BuildTechView(modSnap, m.faction, m.selectedTree, m.width, m.height, m.renderer)
}

func (m Model) viewHelp() string {
	var sb strings.Builder

	borderStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#40E0D0")).Bold(true)
	dimStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	keyStyle := m.renderer.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	sb.WriteString("\n")
	sb.WriteString(borderStyle.Render("  ┌─ GATEWARS — CONTROLS ─────────────────────────────┐") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + keyStyle.Render("GALAXY MAP") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Arrows/hjkl  Navigate between systems") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Enter        Open system view") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("t            Tech tree / research allocation") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("d            Diplomacy view") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("s            Ship designer") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Tab          Scoreboard") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + keyStyle.Render("SYSTEM VIEW") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Arrows/hjkl  Navigate between systems") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Enter        Manage colony (if owned)") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Esc          Back to galaxy map") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + keyStyle.Render("COLONY MANAGEMENT") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Up/Down      Select production slider") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Left/Right   Adjust slider (+/-5%)") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("Esc          Back to system view") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("?  Toggle this help") + "\n")
	sb.WriteString(borderStyle.Render("  │  ") + dimStyle.Render("q  Disconnect") + "\n")
	sb.WriteString(borderStyle.Render("  │") + "\n")
	sb.WriteString(borderStyle.Render("  └───────────────────────────────────────────────────┘") + "\n")

	return sb.String()
}

// IsViewOnly returns whether this is a view-only session.
func (m Model) IsViewOnly() bool {
	return m.viewOnly
}

// LastActivity returns the time of the last user input.
func (m Model) LastActivity() time.Time {
	return m.lastActivity
}

// Cleanup should be called when the session ends.
func (m Model) Cleanup() {
	if m.sshKey != "" {
		m.engine.UnregisterPlayer(m.sshKey)
	}
}

// adjustSliders5 adjusts one slider and rebalances (sum=100).
func adjustSliders5(s [5]int, idx, delta int) [5]int {
	newVal := s[idx] + delta
	if newVal < 0 {
		newVal = 0
	}
	if newVal > 100 {
		newVal = 100
	}
	diff := newVal - s[idx]
	if diff == 0 {
		return s
	}
	s[idx] = newVal

	remaining := -diff
	if remaining > 0 {
		for _, p := range []int{4, 3, 2, 1, 0} {
			if p == idx || remaining <= 0 {
				continue
			}
			take := min(remaining, s[p])
			s[p] -= take
			remaining -= take
		}
	} else {
		give := -remaining
		target := 4
		if target == idx {
			target = 0
		}
		s[target] += give
	}
	return s
}

// adjustSliders6 adjusts one allocation and rebalances (sum=100).
func adjustSliders6(s [6]int, idx, delta int) [6]int {
	newVal := s[idx] + delta
	if newVal < 0 {
		newVal = 0
	}
	if newVal > 100 {
		newVal = 100
	}
	diff := newVal - s[idx]
	if diff == 0 {
		return s
	}
	s[idx] = newVal

	remaining := -diff
	if remaining > 0 {
		for _, p := range []int{5, 4, 3, 2, 1, 0} {
			if p == idx || remaining <= 0 {
				continue
			}
			take := min(remaining, s[p])
			s[p] -= take
			remaining -= take
		}
	} else {
		give := -remaining
		target := 5
		if target == idx {
			target = 0
		}
		s[target] += give
	}
	return s
}
