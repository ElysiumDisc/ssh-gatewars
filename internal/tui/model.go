package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/chat"
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/entity"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/simulation"
	"ssh-gatewars/internal/store"
	"ssh-gatewars/internal/tui/views"
)

type tickMsg time.Time

// chatMsg wraps a chat message for Bubbletea's Cmd system.
type chatMsg chat.ChatMessage

// ModelConfig holds dependencies for creating a TUI model.
type ModelConfig struct {
	Engine  *simulation.Engine
	Store   *store.PlayerStore
	ChatHub *chat.Hub
	Session *core.SessionInfo
	Width   int
	Height  int
}

// Model is the per-session Bubbletea model.
type Model struct {
	engine  *simulation.Engine
	store   *store.PlayerStore
	chatHub *chat.Hub
	session *core.SessionInfo
	char    *entity.Character

	state    ViewState
	width    int
	height   int

	// Fog of war per planet (reset on planet change)
	fog       *views.FogOfWar
	fogPlanet string // which planet the fog is for

	// DHD state
	dhdSymbols  []int
	dhdCursor   int
	dhdInput    string // typed address string

	// Inventory state
	invCursor   int

	// Address book state
	addrCursor  int

	// Call sign input
	callSignBuf string

	// Flash message
	flash     string
	flashTime time.Time

	// Autosave timer
	lastSave time.Time

	// Chat state
	chatOutbox   chan chat.ChatMessage  // receives messages from hub
	chatMessages []chat.ChatMessage     // visible message buffer
	chatInput    string                 // current typing buffer
	chatPanel    views.ChatPanelState   // Hidden/Compact/Expanded
	chatFocus    FocusTarget            // Game or Chat
	chatChannel  string                 // active channel key

	// Toast notification
	toast     string
	toastTime time.Time

	// Player list modal
	playerListOpen bool

	// Aim mode
	aimTarget core.Pos // reticle position

	// Star map
	starMap *views.StarMapState
}

// NewModel creates a new TUI model for a session.
func NewModel(cfg ModelConfig) tea.Model {
	m := &Model{
		engine:   cfg.Engine,
		store:    cfg.Store,
		chatHub:  cfg.ChatHub,
		session:  cfg.Session,
		state:    ViewSplash,
		width:    cfg.Width,
		height:   cfg.Height,
		lastSave: time.Now(),
		chatOutbox:  make(chan chat.ChatMessage, 200),
		chatPanel:   views.ChatHidden,
		chatFocus:   FocusGame,
		chatChannel: "ops",
	}

	// Try to load existing character
	if c, err := cfg.Store.LoadCharacter(cfg.Session.SSHKey, core.DefaultConfig()); err == nil && c != nil {
		m.char = c
		m.state = ViewSGC
		cfg.Engine.RegisterCharacter(c)
		m.ensureFog()
		m.chatConnect()
	}

	return m
}

// NewRejectModel creates a model that displays a rejection message then quits.
func NewRejectModel(reason string) tea.Model {
	return &rejectModel{reason: reason}
}

type rejectModel struct {
	reason string
}

func (m *rejectModel) Init() tea.Cmd { return nil }
func (m *rejectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}
func (m *rejectModel) View() string {
	return fmt.Sprintf("\n  %s\n\n  Press any key to disconnect.\n", m.reason)
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.Tick(time.Second/15, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
	}
	if m.chatHub != nil {
		cmds = append(cmds, m.waitForChatMsg())
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case chatMsg:
		cm := chat.ChatMessage(msg)
		m.chatMessages = append(m.chatMessages, cm)
		// Keep last 200 messages
		if len(m.chatMessages) > 200 {
			m.chatMessages = m.chatMessages[len(m.chatMessages)-200:]
		}
		// Show toast if chat panel is hidden
		if m.chatPanel == views.ChatHidden {
			if cm.Kind == chat.MsgSystem {
				m.toast = fmt.Sprintf("[%s] %s", chat.ChannelDisplayName(cm.Channel), cm.Body)
			} else {
				m.toast = fmt.Sprintf("[%s] <%s> %s", chat.ChannelDisplayName(cm.Channel), cm.SenderCallsign, cm.Body)
			}
			m.toastTime = time.Now()
		}
		// Update active channel from hub
		m.chatChannel = cm.Channel
		return m, m.waitForChatMsg()

	case tickMsg:
		// Drain engine events
		if m.char != nil {
			for _, ev := range m.engine.DrainEvents(m.session.SSHKey) {
				m.flash = ev
				m.flashTime = time.Now()
			}
		}

		// Clear old flash
		if m.flash != "" && time.Since(m.flashTime) > 4*time.Second {
			m.flash = ""
		}

		// Clear old toast
		if m.toast != "" && time.Since(m.toastTime) > 5*time.Second {
			m.toast = ""
		}

		// Autosave
		if m.char != nil && time.Since(m.lastSave) > 60*time.Second {
			m.saveCharacter()
			m.lastSave = time.Now()
		}

		// Update fog
		m.ensureFog()
		if m.char != nil && m.fog != nil {
			m.fog.Reveal(m.char.Pos, 8)
		}

		return m, tea.Tick(time.Second/15, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})

	case tea.KeyMsg:
		m.session.LastInput = time.Now()
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If chat has focus, route all keys to chat input
	if m.chatFocus == FocusChat {
		return m.updateChatInput(msg)
	}

	// Player list modal intercepts keys
	if m.playerListOpen {
		if msg.String() == "tab" || msg.String() == "esc" {
			m.playerListOpen = false
		}
		return m, nil
	}

	switch m.state {
	case ViewSplash:
		return m.updateSplash(msg)
	case ViewCallSign:
		return m.updateCallSign(msg)
	case ViewSGC, ViewPlanet:
		return m.updateExplore(msg)
	case ViewAimMode:
		return m.updateAimMode(msg)
	case ViewStarMap:
		return m.updateStarMap(msg)
	case ViewDHD:
		return m.updateDHD(msg)
	case ViewInventory:
		return m.updateInventory(msg)
	case ViewAddressBook:
		return m.updateAddressBook(msg)
	case ViewHelp:
		m.state = m.previousExploreState()
		return m, nil
	case ViewDead:
		m.state = ViewSGC
		return m, nil
	}
	return m, nil
}

func (m *Model) updateSplash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if MapKey(msg) == KeyQuit {
		return m, tea.Quit
	}
	if m.char != nil {
		// Returning player
		if m.char.Location == "sgc" {
			m.state = ViewSGC
		} else {
			m.state = ViewPlanet
		}
	} else {
		m.state = ViewCallSign
		m.callSignBuf = ""
	}
	return m, nil
}

func (m *Model) updateCallSign(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.callSignBuf) > 0 {
			m.createCharacter(m.callSignBuf)
			m.state = ViewSGC
		}
	case "backspace":
		if len(m.callSignBuf) > 0 {
			m.callSignBuf = m.callSignBuf[:len(m.callSignBuf)-1]
		}
	case "esc", "ctrl+c":
		return m, tea.Quit
	default:
		r := msg.String()
		if len(r) == 1 && len(m.callSignBuf) < 20 {
			m.callSignBuf += r
		}
	}
	return m, nil
}

func (m *Model) updateExplore(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := MapKey(msg)

	switch action {
	case KeyQuit:
		m.chatDisconnect()
		m.saveCharacter()
		m.engine.UnregisterCharacter(m.session.SSHKey)
		return m, tea.Quit
	case KeyUp:
		m.enqueueMove(core.DirUp)
	case KeyDown:
		m.enqueueMove(core.DirDown)
	case KeyLeft:
		m.enqueueMove(core.DirLeft)
	case KeyRight:
		m.enqueueMove(core.DirRight)
	case KeyInteract, KeyPickup:
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:      simulation.ActionInteract,
			PlayerKey: m.session.SSHKey,
		})
	case KeyDial:
		m.state = ViewDHD
		m.dhdSymbols = nil
		m.dhdCursor = 0
		m.dhdInput = ""
	case KeyInventory:
		m.state = ViewInventory
		m.invCursor = 0
	case KeyAddressBook:
		m.state = ViewAddressBook
		m.addrCursor = 0
	case KeyFire:
		if m.char != nil && m.char.Weapon != nil {
			wDef := gamedata.Items[m.char.Weapon.DefID]
			if wDef.WType == gamedata.WeaponRanged {
				m.state = ViewAimMode
				m.aimTarget = m.char.Pos
			} else {
				m.flash = "Equip a ranged weapon to fire."
				m.flashTime = time.Now()
			}
		}
	case KeyReload:
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:      simulation.ActionReload,
			PlayerKey: m.session.SSHKey,
		})
	case KeyChat:
		m.toggleChat()
	case KeyPlayerList:
		m.playerListOpen = !m.playerListOpen
	case KeyStarMap:
		if m.char != nil {
			m.starMap = views.NewStarMapState(m.char.DiscoveredAddresses, m.char.Location)
			m.state = ViewStarMap
		}
	case KeyHelp:
		m.state = ViewHelp
	case KeyCancel:
		if m.chatPanel != views.ChatHidden {
			m.chatPanel = views.ChatHidden
			m.chatFocus = FocusGame
		} else if m.state == ViewPlanet {
			// Can't escape from a planet — must dial the gate
		} else {
			m.chatDisconnect()
			m.saveCharacter()
			m.engine.UnregisterCharacter(m.session.SSHKey)
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) updateDHD(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = m.previousExploreState()
		return m, nil
	case "backspace":
		if len(m.dhdSymbols) > 0 {
			m.dhdSymbols = m.dhdSymbols[:len(m.dhdSymbols)-1]
		}
		if len(m.dhdInput) > 0 {
			m.dhdInput = m.dhdInput[:len(m.dhdInput)-1]
		}
	case "left":
		if m.dhdCursor > 0 {
			m.dhdCursor--
		}
	case "right":
		if m.dhdCursor < gamedata.GlyphCount-1 {
			m.dhdCursor++
		}
	case "up":
		if m.dhdCursor >= 10 {
			m.dhdCursor -= 10
		}
	case "down":
		if m.dhdCursor+10 < gamedata.GlyphCount {
			m.dhdCursor += 10
		}
	case "enter":
		if len(m.dhdSymbols) < 7 {
			// Lock current glyph
			m.dhdSymbols = append(m.dhdSymbols, m.dhdCursor)
		}
		if len(m.dhdSymbols) == 7 {
			// Try to dial
			var addr gamedata.GateAddress
			copy(addr[:], m.dhdSymbols)
			if addr.IsValid() {
				m.engine.EnqueueAction(simulation.PlayerAction{
					Type:      simulation.ActionDialGate,
					PlayerKey: m.session.SSHKey,
					Address:   addr,
				})
				m.state = ViewPlanet
				m.fogPlanet = "" // force fog reset
			} else {
				m.flash = "Invalid address — duplicate glyphs!"
				m.flashTime = time.Now()
				m.dhdSymbols = nil
			}
		}
	default:
		// Typed input for quick-dial (e.g. "26-6-14-31-11-29-0")
		r := msg.String()
		if len(r) == 1 && (r[0] >= '0' && r[0] <= '9' || r[0] == '-') {
			m.dhdInput += r
			// Try to parse as complete address
			if addr, ok := gamedata.ParseAddress(m.dhdInput); ok {
				m.engine.EnqueueAction(simulation.PlayerAction{
					Type:      simulation.ActionDialGate,
					PlayerKey: m.session.SSHKey,
					Address:   addr,
				})
				m.state = ViewPlanet
				m.fogPlanet = ""
			}
		}
	}
	return m, nil
}

func (m *Model) updateInventory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.char == nil {
		return m, nil
	}

	action := MapKey(msg)
	switch action {
	case KeyCancel:
		m.state = m.previousExploreState()
	case KeyUp:
		if m.invCursor > 0 {
			m.invCursor--
		}
	case KeyDown:
		if m.invCursor < len(m.char.Inventory)-1 {
			m.invCursor++
		}
	case KeyEquip:
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:      simulation.ActionEquip,
			PlayerKey: m.session.SSHKey,
			ItemIndex: m.invCursor,
		})
	case KeyUse:
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:      simulation.ActionUseItem,
			PlayerKey: m.session.SSHKey,
			ItemIndex: m.invCursor,
		})
	case KeyDrop:
		// Drop item on ground (simplified: just remove)
		if m.invCursor >= 0 && m.invCursor < len(m.char.Inventory) {
			m.char.RemoveItem(m.char.Inventory[m.invCursor].DefID, 1)
			if m.invCursor >= len(m.char.Inventory) && m.invCursor > 0 {
				m.invCursor--
			}
		}
	}
	return m, nil
}

func (m *Model) updateAddressBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.char == nil {
		return m, nil
	}

	action := MapKey(msg)
	switch action {
	case KeyCancel:
		m.state = m.previousExploreState()
	case KeyUp:
		if m.addrCursor > 0 {
			m.addrCursor--
		}
	case KeyDown:
		if m.addrCursor < len(m.char.DiscoveredAddresses)-1 {
			m.addrCursor++
		}
	case KeyConfirm:
		if m.addrCursor >= 0 && m.addrCursor < len(m.char.DiscoveredAddresses) {
			addr := m.char.DiscoveredAddresses[m.addrCursor]
			m.engine.EnqueueAction(simulation.PlayerAction{
				Type:      simulation.ActionDialGate,
				PlayerKey: m.session.SSHKey,
				Address:   addr,
			})
			m.state = ViewPlanet
			m.fogPlanet = ""
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	switch m.state {
	case ViewSplash:
		return views.RenderSplash(m.width, m.height)
	case ViewCallSign:
		return m.viewCallSign()
	case ViewSGC, ViewPlanet:
		return m.viewExplore()
	case ViewAimMode:
		return m.viewAimMode()
	case ViewStarMap:
		if m.starMap != nil {
			return views.RenderStarMap(m.starMap, m.width, m.height)
		}
	case ViewDHD:
		return views.RenderDHD(m.dhdSymbols, m.dhdCursor, m.width, m.height)
	case ViewInventory:
		if m.char != nil {
			return views.RenderInventory(m.char, m.invCursor, m.width, m.height)
		}
	case ViewAddressBook:
		if m.char != nil {
			return views.RenderAddressBook(m.char.DiscoveredAddresses, m.addrCursor)
		}
	case ViewHelp:
		return views.RenderHelp()
	case ViewDead:
		return m.viewDead()
	}
	return ""
}

func (m *Model) viewCallSign() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#44AAFF")).Bold(true)
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA44"))
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("  Welcome to Stargate Command"))
	b.WriteString("\n\n")
	b.WriteString(promptStyle.Render("  Enter your call sign, soldier: "))
	b.WriteString(inputStyle.Render(m.callSignBuf))
	b.WriteString(inputStyle.Render("_"))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  (Press Enter to confirm)"))
	return b.String()
}

func (m *Model) viewExplore() string {
	if m.char == nil {
		return "Loading..."
	}

	snap := m.engine.GetPlanetSnapshot(m.session.SSHKey)
	if snap == nil {
		return "Loading planet..."
	}

	// Update character position from engine
	c := m.engine.GetCharacter(m.session.SSHKey)
	if c != nil {
		m.char = c
	}

	// Determine chat rows
	chatRows := 0
	switch m.chatPanel {
	case views.ChatCompact:
		chatRows = 8 // 6 messages + separator + input
	case views.ChatExpanded:
		chatRows = m.height / 2
	}

	// Compute view dimensions (reserve 3 for HUD + chat rows)
	viewH := m.height - 3 - chatRows
	if viewH < 5 {
		viewH = 5
	}
	viewW := m.width
	if viewW < 10 {
		viewW = 10
	}

	// Render map
	mapView := views.RenderPlanet(snap, m.char.Pos, m.fog, viewW, viewH, m.session.SSHKey)

	// Render HUD
	weaponName := "Fists"
	if m.char.Weapon != nil {
		weaponName = m.char.Weapon.Def().Name
	}

	hud := views.RenderHUD(views.HUDData{
		HP:          m.char.HP,
		MaxHP:       m.char.MaxHP,
		Level:       m.char.Level,
		XP:          m.char.XP,
		WeaponName:  weaponName,
		PlanetName:  snap.PlanetName,
		Biome:       snap.Biome,
		Threat:      snap.Threat,
		Flash:       m.flash,
		OnlineCount: m.engine.GetOnlineCount(),
	}, m.width)

	result := mapView + "\n" + hud

	// Toast notification (when chat hidden)
	if m.chatPanel == views.ChatHidden && m.toast != "" {
		result += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(m.toast)
	}

	// Chat panel overlay
	if m.chatPanel != views.ChatHidden {
		chatView := views.RenderChatPanel(
			m.chatMessages,
			m.chatChannel,
			m.chatInput,
			m.chatPanel,
			m.width,
			chatRows,
		)
		result += "\n" + chatView
	}

	// Player list modal (rendered on top)
	if m.playerListOpen {
		entries := m.buildPlayerList(snap)
		result = views.RenderPlayerList(entries, m.width, m.height)
	}

	return result
}

func (m *Model) viewAimMode() string {
	if m.char == nil || m.char.Weapon == nil {
		return "No weapon equipped."
	}

	snap := m.engine.GetPlanetSnapshot(m.session.SSHKey)
	if snap == nil {
		return "Loading..."
	}

	c := m.engine.GetCharacter(m.session.SSHKey)
	if c != nil {
		m.char = c
	}

	viewH := m.height - 3
	if viewH < 5 {
		viewH = 5
	}
	viewW := m.width
	if viewW < 10 {
		viewW = 10
	}

	wDef := gamedata.Items[m.char.Weapon.DefID]
	return views.RenderAimOverlay(snap, m.char.Pos, m.aimTarget, m.fog, viewW, viewH, m.session.SSHKey, wDef.Range)
}

func (m *Model) viewDead() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	return fmt.Sprintf("\n\n%s\n\n%s\n",
		style.Render("  YOU HAVE FALLEN"),
		dim.Render("  Press any key to respawn at SGC..."))
}

// Helpers

func (m *Model) createCharacter(callSign string) {
	cfg := core.DefaultConfig()
	m.char = entity.NewCharacter(m.session.SSHKey, m.session.DisplayName, callSign, cfg)

	// Save to DB
	m.store.UpsertPlayer(m.session.SSHKey, m.session.DisplayName, callSign)
	m.store.SaveCharacter(m.char)

	// Register with engine
	m.engine.RegisterCharacter(m.char)
	m.ensureFog()

	// Connect to chat
	m.chatConnect()
}

func (m *Model) saveCharacter() {
	if m.char != nil {
		m.store.SaveCharacter(m.char)
	}
}

func (m *Model) enqueueMove(dir core.Pos) {
	m.engine.EnqueueAction(simulation.PlayerAction{
		Type:      simulation.ActionMove,
		PlayerKey: m.session.SSHKey,
		Dir:       dir,
	})
}

func (m *Model) ensureFog() {
	if m.char == nil {
		return
	}
	snap := m.engine.GetPlanetSnapshot(m.session.SSHKey)
	if snap == nil {
		return
	}
	planetKey := snap.AddressCode
	if m.fogPlanet != planetKey || m.fog == nil {
		m.fog = views.NewFogOfWar(snap.MapWidth, snap.MapHeight)
		m.fogPlanet = planetKey
	}
}

func (m *Model) previousExploreState() ViewState {
	if m.char != nil && m.char.Location == "sgc" {
		return ViewSGC
	}
	return ViewPlanet
}

// --- Aim mode ---

func (m *Model) updateAimMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = m.previousExploreState()
	case "enter", "f":
		// Fire at target
		m.engine.EnqueueAction(simulation.PlayerAction{
			Type:       simulation.ActionFire,
			PlayerKey:  m.session.SSHKey,
			FireTarget: m.aimTarget,
		})
		m.state = m.previousExploreState()
	case "up", "k", "w":
		m.aimTarget.Y--
	case "down", "j", "s":
		m.aimTarget.Y++
	case "left", "h", "a":
		m.aimTarget.X--
	case "right", "l", "d":
		m.aimTarget.X++
	}
	return m, nil
}

// --- Star map ---

func (m *Model) updateStarMap(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.starMap == nil {
		m.state = m.previousExploreState()
		return m, nil
	}

	panSpeed := 3.0
	switch msg.String() {
	case "esc", "m":
		m.state = m.previousExploreState()
	case "up", "k", "w":
		m.starMap.CamY -= panSpeed
	case "down", "j", "s":
		m.starMap.CamY += panSpeed
	case "left", "h", "a":
		m.starMap.CamX -= panSpeed
	case "right", "l", "d":
		m.starMap.CamX += panSpeed
	case "+", "=":
		if m.starMap.Zoom < 4 {
			m.starMap.Zoom++
		}
	case "-", "_":
		if m.starMap.Zoom > 0 {
			m.starMap.Zoom--
		}
	case "tab":
		// Cycle to next star
		if len(m.starMap.Stars) > 0 {
			m.starMap.Cursor = (m.starMap.Cursor + 1) % len(m.starMap.Stars)
			star := m.starMap.Stars[m.starMap.Cursor]
			m.starMap.CamX = star.WorldX
			m.starMap.CamY = star.WorldY
		}
	case "shift+tab":
		// Cycle to previous star
		if len(m.starMap.Stars) > 0 {
			m.starMap.Cursor--
			if m.starMap.Cursor < 0 {
				m.starMap.Cursor = len(m.starMap.Stars) - 1
			}
			star := m.starMap.Stars[m.starMap.Cursor]
			m.starMap.CamX = star.WorldX
			m.starMap.CamY = star.WorldY
		}
	case "enter":
		// Dial selected star's address
		if m.starMap.Cursor >= 0 && m.starMap.Cursor < len(m.starMap.Stars) {
			addr := m.starMap.Stars[m.starMap.Cursor].Address
			if addr == gamedata.EarthAddress && m.char.Location == "sgc" {
				m.flash = "You're already at SGC."
				m.flashTime = time.Now()
			} else {
				m.engine.EnqueueAction(simulation.PlayerAction{
					Type:      simulation.ActionDialGate,
					PlayerKey: m.session.SSHKey,
					Address:   addr,
				})
				m.state = ViewPlanet
				m.fogPlanet = ""
			}
		}
	}
	return m, nil
}

// --- Chat integration ---

// waitForChatMsg returns a Cmd that waits for the next chat message.
func (m *Model) waitForChatMsg() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-m.chatOutbox
		if !ok {
			return nil
		}
		return chatMsg(msg)
	}
}

// chatConnect sends a connect event to the chat hub.
func (m *Model) chatConnect() {
	if m.chatHub == nil || m.char == nil {
		return
	}
	m.chatHub.Incoming <- chat.ChatEvent{
		Type:        chat.EventConnect,
		Fingerprint: m.session.SSHKey,
		Callsign:    m.char.CallSign,
		Outbox:      m.chatOutbox,
	}
}

// chatDisconnect sends a disconnect event to the chat hub.
func (m *Model) chatDisconnect() {
	if m.chatHub == nil {
		return
	}
	m.chatHub.Incoming <- chat.ChatEvent{
		Type:        chat.EventDisconnect,
		Fingerprint: m.session.SSHKey,
	}
}

// toggleChat cycles chat panel: Hidden → Compact → Expanded → Hidden.
// Entering Compact or Expanded sets focus to Chat.
func (m *Model) toggleChat() {
	switch m.chatPanel {
	case views.ChatHidden:
		m.chatPanel = views.ChatCompact
		m.chatFocus = FocusChat
	case views.ChatCompact:
		m.chatPanel = views.ChatExpanded
		m.chatFocus = FocusChat
	case views.ChatExpanded:
		m.chatPanel = views.ChatHidden
		m.chatFocus = FocusGame
		m.chatInput = ""
	}
}

// updateChatInput handles keypresses when chat has focus.
func (m *Model) updateChatInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.chatFocus = FocusGame
		m.chatPanel = views.ChatHidden
		m.chatInput = ""
	case "enter":
		m.sendChatMessage()
	case "backspace":
		if len(m.chatInput) > 0 {
			m.chatInput = m.chatInput[:len(m.chatInput)-1]
		}
	case "tab":
		// Switch focus back to game without closing panel
		m.chatFocus = FocusGame
	case "c":
		// If input is empty, 'c' toggles chat off
		if m.chatInput == "" {
			m.toggleChat()
			return m, nil
		}
		m.chatInput += "c"
	default:
		r := msg.String()
		if len(r) == 1 && len(m.chatInput) < 500 {
			m.chatInput += r
		}
	}
	return m, nil
}

// sendChatMessage sends the current input to the chat hub.
func (m *Model) sendChatMessage() {
	text := strings.TrimSpace(m.chatInput)
	m.chatInput = ""
	if text == "" || m.chatHub == nil {
		return
	}

	// Handle slash commands
	if text[0] == '/' {
		parts := strings.SplitN(text[1:], " ", 2)
		cmd := parts[0]
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}
		m.chatHub.Incoming <- chat.ChatEvent{
			Type:        chat.EventSlashCommand,
			Fingerprint: m.session.SSHKey,
			Command:     cmd,
			Args:        args,
		}
		return
	}

	// Handle @DM shorthand
	if text[0] == '@' {
		m.chatHub.Incoming <- chat.ChatEvent{
			Type:        chat.EventSendMessage,
			Fingerprint: m.session.SSHKey,
			Body:        text,
		}
		return
	}

	// Regular message
	m.chatHub.Incoming <- chat.ChatEvent{
		Type:        chat.EventSendMessage,
		Fingerprint: m.session.SSHKey,
		Body:        text,
	}
}

// buildPlayerList creates PlayerListEntry slice from the current planet snapshot.
func (m *Model) buildPlayerList(snap *simulation.PlanetSnapshot) []views.PlayerListEntry {
	var entries []views.PlayerListEntry
	for _, p := range snap.Players {
		entry := views.PlayerListEntry{
			Callsign: p.CallSign,
			Level:    0, // level not in PlayerSnapshot, just show online
			Location: snap.PlanetName,
		}
		entries = append(entries, entry)
	}
	return entries
}
