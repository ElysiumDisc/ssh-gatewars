package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"ssh-gatewars/internal/chat"
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/engine"
	"ssh-gatewars/internal/game"
	"ssh-gatewars/internal/store"
	"ssh-gatewars/internal/tui/views"
)

// tickMsg is sent at the render frame rate.
type tickMsg time.Time

// chatMsg wraps an incoming chat message.
type chatMsg chat.ChatMessage

// ModelConfig holds dependencies for creating a TUI model.
type ModelConfig struct {
	Engine  *engine.Engine
	Store   *store.PlayerStore
	ChatHub *chat.Hub
	Session *core.SessionInfo
	Width   int
	Height  int
}

// Model is the top-level Bubbletea model for a player session.
type Model struct {
	state   State
	session *core.SessionInfo
	engine  *engine.Engine
	store   *store.PlayerStore
	chatHub *chat.Hub

	// Player data
	callsign    string
	player      *store.PlayerRecord
	chatOutbox  chan chat.ChatMessage
	chatVisible bool

	// Active defense state
	activePlanetID int
	defenseSnap    *engine.DefenseSnapshot

	// View models
	splash   views.SplashModel
	throne   views.ThroneModel
	galaxy   views.GalaxyModel
	defense  views.DefenseModel
	chatView views.ChatModel

	// Chat input
	chatInput string
	chatMode  bool // true when typing in chat

	// Animation
	frameCount int

	// Layout
	width, height int

	// Chat messages buffer
	chatMessages []chat.ChatMessage
}

// NewModel creates a new TUI model for a player session.
func NewModel(cfg ModelConfig) *Model {
	return &Model{
		state:       StateSplash,
		session:     cfg.Session,
		engine:      cfg.Engine,
		store:       cfg.Store,
		chatHub:     cfg.ChatHub,
		chatOutbox:  make(chan chat.ChatMessage, 256),
		width:       cfg.Width,
		height:      cfg.Height,
		splash:      views.NewSplashModel(),
		throne:      views.NewThroneModel(),
		galaxy:      views.NewGalaxyModel(),
		defense:     views.NewDefenseModel(),
		chatView:    views.NewChatModel(),
		chatMessages: make([]chat.ChatMessage, 0, 100),
		activePlanetID: -1,
	}
}

// NewRejectModel creates a model that shows a rejection message and exits.
func NewRejectModel(reason string) tea.Model {
	return &rejectModel{reason: reason}
}

type rejectModel struct {
	reason string
}

func (m *rejectModel) Init() tea.Cmd  { return tea.Quit }
func (m *rejectModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, tea.Quit }
func (m *rejectModel) View() string   { return "\n  " + m.reason + "\n\n" }

// Init starts the model.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.tickCmd(),
		m.listenChat(),
	)
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second/time.Duration(15), func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) listenChat() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.chatOutbox
		if !ok {
			return nil
		}
		return chatMsg(msg)
	}
}

// Update handles input and state transitions.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Refresh defense snapshot if active
		if m.state == StateDefense && m.activePlanetID >= 0 {
			snap := m.engine.GetDefenseSnapshot(m.activePlanetID)
			if snap != nil {
				m.defenseSnap = snap
			}
			// Check if defense ended
			if snap != nil && (snap.Liberated || snap.Failed) {
				m.handleDefenseEnd(snap)
			}
		}
		m.splash.Tick()
		m.throne.Tick()
		m.frameCount++
		return m, m.tickCmd()

	case chatMsg:
		cm := chat.ChatMessage(msg)
		m.chatMessages = append(m.chatMessages, cm)
		if len(m.chatMessages) > 100 {
			m.chatMessages = m.chatMessages[len(m.chatMessages)-100:]
		}
		return m, m.listenChat()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global: ctrl+c to quit
	if key == "ctrl+c" {
		m.cleanup()
		return m, tea.Quit
	}

	// Chat mode intercepts all keys
	if m.chatMode {
		return m.handleChatInput(msg)
	}

	switch m.state {
	case StateSplash:
		// Any key advances to callsign
		m.state = StateCallsign
		// Check if we have a saved callsign
		if cs, err := m.store.GetCallsign(m.session.SSHKey); err == nil && cs != "" {
			m.callsign = cs
			m.loadPlayer()
			m.connectChat()
			m.state = StateAtlantis
		}
		return m, nil

	case StateCallsign:
		return m.handleCallsignInput(msg)

	case StateAtlantis:
		switch key {
		case "t":
			m.loadPlayer() // refresh before showing upgrades
			m.state = StateThrone
		case "g":
			m.state = StateGalaxy
			m.galaxy.Reset(m.engine.GetGalaxySnapshot())
		case "c":
			m.chatMode = true
		case "q":
			m.cleanup()
			return m, tea.Quit
		}
		return m, nil

	case StateThrone:
		return m.handleThroneInput(msg)

	case StateGalaxy:
		switch key {
		case "q", "esc":
			m.state = StateAtlantis
		case "up", "k":
			m.galaxy.MoveSelection(-1)
		case "down", "j":
			m.galaxy.MoveSelection(1)
		case "enter":
			planetID := m.galaxy.SelectedPlanetID()
			if planetID >= 0 {
				m.deployToPlanet(planetID)
			}
		case "c":
			m.chatMode = true
		}
		return m, nil

	case StateDefense:
		switch key {
		case "q", "esc":
			m.retreatFromPlanet()
			m.state = StateAtlantis
		case "c":
			m.chatMode = true
		case "tab":
			m.chatVisible = !m.chatVisible
		case "1":
			m.setTactic(game.TacticSpread)
		case "2":
			m.setTactic(game.TacticFocus)
		case "3":
			m.setTactic(game.TacticPerimeter)
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) handleCallsignInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "enter":
		if len(m.callsign) >= 2 {
			m.store.UpsertPlayer(m.session.SSHKey, m.session.DisplayName, m.callsign)
			m.store.SetCallsign(m.session.SSHKey, m.callsign)
			m.loadPlayer()
			m.connectChat()
			m.state = StateAtlantis
		}
	case "backspace":
		if len(m.callsign) > 0 {
			m.callsign = m.callsign[:len(m.callsign)-1]
		}
	default:
		if len(key) == 1 && len(m.callsign) < 16 {
			c := key[0]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
				m.callsign += key
			}
		}
	}
	return m, nil
}

func (m *Model) handleThroneInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "q", "esc":
		m.state = StateAtlantis
	case "up", "k":
		m.throne.MoveSelection(-1)
	case "down", "j":
		m.throne.MoveSelection(1)
	case "enter":
		m.executeThroneUpgrade()
	}
	return m, nil
}

func (m *Model) executeThroneUpgrade() {
	if m.player == nil {
		return
	}
	sel := m.throne.Selected

	switch sel {
	case 0: // Chair upgrade
		if m.player.ChairLevel >= views.MaxChairLevel {
			m.throne.SetStatus("Already at max level!", views.StyleDim)
			return
		}
		cost := views.ChairUpgradeCost(m.player.ChairLevel)
		ok, err := m.store.SpendZPM(m.session.SSHKey, cost)
		if err != nil || !ok {
			m.throne.SetStatus("Not enough ZPM!", views.StyleDanger)
			return
		}
		m.store.UpgradeChair(m.session.SSHKey)
		m.loadPlayer()
		m.throne.SetStatus("Shield Generator upgraded!", views.StyleSuccess)

	case 1, 2, 3: // Drone tier
		tier := sel
		if m.player.DroneTier == tier {
			m.throne.SetStatus("Already equipped!", views.StyleDim)
			return
		}
		cost := views.DroneTierCost(tier)
		if cost == 0 {
			return
		}
		ok, err := m.store.SpendZPM(m.session.SSHKey, cost)
		if err != nil || !ok {
			m.throne.SetStatus("Not enough ZPM!", views.StyleDanger)
			return
		}
		m.store.UpgradeDroneTier(m.session.SSHKey, tier)
		m.loadPlayer()
		m.throne.SetStatus("Drone weapons upgraded!", views.StyleSuccess)

	case 4: // Faction switch
		currentFaction := game.Faction(m.player.Faction)
		newFaction := game.FactionAncient
		if currentFaction == game.FactionAncient {
			newFaction = game.FactionOri
		}
		// Reset upgrades + switch faction
		m.store.ResetPlayer(m.session.SSHKey)
		m.store.SetFaction(m.session.SSHKey, int(newFaction))
		m.loadPlayer()
		name := game.FactionDefs[newFaction].Name
		m.throne.SetStatus("Switched to "+name+" Path! Upgrades reset.", views.StyleGold)

	case 5: // Full reset
		m.store.ResetPlayer(m.session.SSHKey)
		m.loadPlayer()
		m.throne.SetStatus("All progress reset. Fresh start.", views.StyleDanger)
	}
}

func (m *Model) handleChatInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.chatMode = false
		m.chatInput = ""
	case "enter":
		if m.chatInput != "" {
			if strings.HasPrefix(m.chatInput, "/") {
				// Slash command
				parts := strings.SplitN(m.chatInput[1:], " ", 2)
				cmd := parts[0]
				args := ""
				if len(parts) > 1 {
					args = parts[1]
				}
				m.chatHub.Incoming <- chat.ChatEvent{
					Type:        chat.EventSlashCommand,
					Fingerprint: m.session.SSHKey,
					Callsign:    m.callsign,
					Command:     cmd,
					Args:        args,
				}
			} else {
				m.chatHub.Incoming <- chat.ChatEvent{
					Type:        chat.EventSendMessage,
					Fingerprint: m.session.SSHKey,
					Callsign:    m.callsign,
					Body:        m.chatInput,
				}
			}
			m.chatInput = ""
		}
		m.chatMode = false
	case "backspace":
		if len(m.chatInput) > 0 {
			m.chatInput = m.chatInput[:len(m.chatInput)-1]
		}
	default:
		if len(key) == 1 && len(m.chatInput) < 500 {
			m.chatInput += key
		}
	}
	return m, nil
}

func (m *Model) loadPlayer() {
	rec, err := m.store.GetPlayer(m.session.SSHKey)
	if err != nil || rec == nil {
		m.store.UpsertPlayer(m.session.SSHKey, m.session.DisplayName, m.callsign)
		rec, _ = m.store.GetPlayer(m.session.SSHKey)
	}
	m.player = rec
}

func (m *Model) connectChat() {
	m.chatHub.Incoming <- chat.ChatEvent{
		Type:        chat.EventConnect,
		Fingerprint: m.session.SSHKey,
		Callsign:    m.callsign,
		Outbox:      m.chatOutbox,
	}
}

func (m *Model) deployToPlanet(planetID int) {
	if m.player == nil {
		return
	}
	tier := game.DroneTier(m.player.DroneTier)
	faction := game.Faction(m.player.Faction)
	m.engine.DeployChair(planetID, m.session.SSHKey, m.callsign, m.player.ChairLevel, tier, faction)
	m.activePlanetID = planetID
	m.state = StateDefense
	m.chatVisible = true

	// Join planet chat
	snap := m.engine.GetGalaxySnapshot()
	if snap != nil && planetID < len(snap.Planets) {
		chKey := chat.PlanetChannelKey(snap.Planets[planetID].Name)
		m.chatHub.Incoming <- chat.ChatEvent{
			Type:        chat.EventJoinChannel,
			Fingerprint: m.session.SSHKey,
			Channel:     chKey,
		}
	}
}

func (m *Model) setTactic(tactic game.DroneTactic) {
	if m.activePlanetID >= 0 {
		m.engine.SetChairTactic(m.activePlanetID, m.session.SSHKey, tactic)
	}
}

func (m *Model) retreatFromPlanet() {
	if m.activePlanetID >= 0 {
		// Save earned ZPM before retreating (you keep what you killed)
		if m.defenseSnap != nil && m.defenseSnap.ZPMEarned > 0 {
			m.store.AddZPM(m.session.SSHKey, m.defenseSnap.ZPMEarned)
			m.store.AddKills(m.session.SSHKey, m.defenseSnap.TotalKills)
			m.loadPlayer()
		}
		m.engine.RetreatChair(m.activePlanetID, m.session.SSHKey)
		m.activePlanetID = -1
		m.defenseSnap = nil
	}
}

func (m *Model) handleDefenseEnd(snap *engine.DefenseSnapshot) {
	if snap.Liberated {
		// Award kill ZPM + bounty
		totalZPM := snap.ZPMEarned + snap.BountyZPM
		m.store.AddZPM(m.session.SSHKey, totalZPM)
		m.store.RecordPlanetFreed(m.session.SSHKey)
		m.store.AddKills(m.session.SSHKey, snap.TotalKills)
	} else {
		// Still award some ZPM for the effort
		m.store.AddZPM(m.session.SSHKey, snap.ZPMEarned/2)
		m.store.AddKills(m.session.SSHKey, snap.TotalKills)
	}
	m.loadPlayer() // refresh stats
	m.activePlanetID = -1
}

func (m *Model) cleanup() {
	if m.activePlanetID >= 0 {
		// Save earned ZPM before disconnecting
		if m.defenseSnap != nil && m.defenseSnap.ZPMEarned > 0 {
			m.store.AddZPM(m.session.SSHKey, m.defenseSnap.ZPMEarned)
			m.store.AddKills(m.session.SSHKey, m.defenseSnap.TotalKills)
		}
		m.engine.RetreatChair(m.activePlanetID, m.session.SSHKey)
	}
	m.chatHub.Incoming <- chat.ChatEvent{
		Type:        chat.EventDisconnect,
		Fingerprint: m.session.SSHKey,
	}
}

// View renders the current state.
func (m *Model) View() string {
	switch m.state {
	case StateSplash:
		return views.RenderSplash(m.splash, m.width, m.height)
	case StateCallsign:
		return views.RenderCallsign(m.callsign, m.frameCount, m.width, m.height)
	case StateAtlantis:
		return views.RenderAtlantis(m.player, m.callsign, m.chatMessages, m.chatInput, m.chatMode, m.engine.OnlinePlayerCount(), m.width, m.height)
	case StateThrone:
		return views.RenderThrone(m.player, m.throne, m.frameCount, m.width, m.height)
	case StateGalaxy:
		return views.RenderGalaxy(m.galaxy, m.width, m.height)
	case StateDefense:
		return views.RenderDefense(m.defenseSnap, m.chatMessages, m.chatInput, m.chatMode, m.chatVisible, m.frameCount, m.session.SSHKey, m.width, m.height)
	}
	return ""
}
