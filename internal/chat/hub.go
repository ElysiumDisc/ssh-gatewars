package chat

import (
	"time"

	"github.com/charmbracelet/log"

	"ssh-gatewars/internal/store"
)

// ChatSession tracks a connected player's chat state.
type ChatSession struct {
	Fingerprint   string
	Callsign      string
	Outbox        chan ChatMessage
	Subscriptions map[string]bool // channel keys
	ActiveChannel string          // where typed messages go
	MuteList      map[string]bool // muted fingerprints
	lastMsg       time.Time       // rate limiting
	msgCount      int             // messages in current window
}

// Hub is the single-goroutine chat message router.
type Hub struct {
	channels   map[string]*Channel
	sessions   map[string]*ChatSession // fingerprint → session
	Incoming   chan ChatEvent
	GameEvents chan GameEvent
	store      *store.PlayerStore
}

// NewHub creates a chat hub.
func NewHub(st *store.PlayerStore) *Hub {
	h := &Hub{
		channels:   make(map[string]*Channel),
		sessions:   make(map[string]*ChatSession),
		Incoming:   make(chan ChatEvent, 1000),
		GameEvents: make(chan GameEvent, 100),
		store:      st,
	}

	// Create permanent #ops channel
	h.channels["ops"] = NewChannel("ops", ChanOps, 100)

	return h
}

// Run starts the hub main loop. Blocks until context channel is closed.
func (h *Hub) Run(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case ev := <-h.Incoming:
			h.handleEvent(ev)
		case gev := <-h.GameEvents:
			h.handleGameEvent(gev)
		}
	}
}

func (h *Hub) handleEvent(ev ChatEvent) {
	switch ev.Type {
	case EventConnect:
		h.handleConnect(ev)
	case EventDisconnect:
		h.handleDisconnect(ev)
	case EventSendMessage:
		h.handleSendMessage(ev)
	case EventJoinChannel:
		h.handleJoinChannel(ev)
	case EventLeaveChannel:
		h.handleLeaveChannel(ev)
	case EventSlashCommand:
		h.handleSlashCommand(ev)
	}
}

func (h *Hub) handleConnect(ev ChatEvent) {
	var mutes map[string]bool
	if muteFPs, err := h.store.GetMutes(ev.Fingerprint); err == nil {
		mutes = make(map[string]bool, len(muteFPs))
		for _, fp := range muteFPs {
			mutes[fp] = true
		}
	} else {
		mutes = make(map[string]bool)
	}

	sess := &ChatSession{
		Fingerprint:   ev.Fingerprint,
		Callsign:      ev.Callsign,
		Outbox:        ev.Outbox,
		Subscriptions: make(map[string]bool),
		ActiveChannel: "ops",
		MuteList:      mutes,
		lastMsg:       time.Now(),
	}
	h.sessions[ev.Fingerprint] = sess

	// Auto-subscribe to #ops
	h.subscribe(sess, "ops")

	// Send MOTD
	h.sendToSession(sess, ChatMessage{
		Channel:        "ops",
		SenderFP:       "system",
		SenderCallsign: "WALTER",
		Kind:           MsgSystem,
		Body:           MOTD(),
		Timestamp:      time.Now(),
	})

	log.Info("chat: player connected", "callsign", ev.Callsign)
}

func (h *Hub) handleDisconnect(ev ChatEvent) {
	sess, ok := h.sessions[ev.Fingerprint]
	if !ok {
		return
	}

	for chKey := range sess.Subscriptions {
		if ch, ok := h.channels[chKey]; ok {
			delete(ch.Members, ev.Fingerprint)
			if (ch.Type == ChanLocal || ch.Type == ChanDM) && len(ch.Members) == 0 {
				delete(h.channels, chKey)
			}
		}
	}

	delete(h.sessions, ev.Fingerprint)
	log.Info("chat: player disconnected", "callsign", sess.Callsign)
}

func (h *Hub) handleSendMessage(ev ChatEvent) {
	sess, ok := h.sessions[ev.Fingerprint]
	if !ok {
		return
	}

	channel := ev.Channel
	if channel == "" {
		channel = sess.ActiveChannel
	}

	// Rate limit: 5 messages per 3 seconds
	now := time.Now()
	if now.Sub(sess.lastMsg) > 3*time.Second {
		sess.msgCount = 0
		sess.lastMsg = now
	}
	sess.msgCount++
	if sess.msgCount > 5 {
		h.sendToSession(sess, WalterMsg("", "Slow down. Radio discipline."))
		return
	}

	ch, ok := h.channels[channel]
	if !ok {
		h.sendToSession(sess, WalterMsg("", "Channel not found."))
		return
	}
	if !ch.Members[ev.Fingerprint] {
		h.sendToSession(sess, WalterMsg("", "You are not on that frequency."))
		return
	}

	msg := ChatMessage{
		Channel:        channel,
		SenderFP:       ev.Fingerprint,
		SenderCallsign: sess.Callsign,
		Kind:           MsgChat,
		Body:           sanitizeMessage(ev.Body),
		Timestamp:      now,
	}

	// Check for @mention DM
	if len(ev.Body) > 1 && ev.Body[0] == '@' {
		h.handleDM(sess, ev.Body)
		return
	}

	// Store if persistent channel
	if ch.Type == ChanOps || ch.Type == ChanTeam {
		h.store.SaveChatMessage(store.ChatMsg{
			Channel:        channel,
			SenderFP:       msg.SenderFP,
			SenderCallsign: msg.SenderCallsign,
			Kind:           int(msg.Kind),
			Body:           msg.Body,
			CreatedAt:      msg.Timestamp,
		})
	}

	ch.Backlog.Push(msg)
	h.fanout(ch, msg)
}

func (h *Hub) handleDM(sender *ChatSession, body string) {
	parts := splitFirst(body[1:], ' ')
	targetCallsign := parts[0]
	dmBody := ""
	if len(parts) > 1 {
		dmBody = parts[1]
	}
	if dmBody == "" {
		h.sendToSession(sender, WalterMsg("", "Usage: @callsign message"))
		return
	}

	var target *ChatSession
	for _, s := range h.sessions {
		if s.Callsign == targetCallsign {
			target = s
			break
		}
	}
	if target == nil {
		h.sendToSession(sender, WalterMsg("", targetCallsign+" is not online."))
		return
	}

	now := time.Now()
	dmKey := DMChannelKey(sender.Fingerprint, target.Fingerprint)

	if _, ok := h.channels[dmKey]; !ok {
		h.channels[dmKey] = NewChannel(dmKey, ChanDM, 0)
	}

	h.sendToSession(target, ChatMessage{
		Channel: dmKey, SenderFP: sender.Fingerprint,
		SenderCallsign: sender.Callsign, Kind: MsgWhisper,
		Body: dmBody, Timestamp: now,
	})
	h.sendToSession(sender, ChatMessage{
		Channel: dmKey, SenderFP: sender.Fingerprint,
		SenderCallsign: "-> " + target.Callsign, Kind: MsgWhisper,
		Body: dmBody, Timestamp: now,
	})
}

func (h *Hub) handleJoinChannel(ev ChatEvent) {
	sess, ok := h.sessions[ev.Fingerprint]
	if !ok {
		return
	}
	h.subscribe(sess, ev.Channel)
}

func (h *Hub) handleLeaveChannel(ev ChatEvent) {
	sess, ok := h.sessions[ev.Fingerprint]
	if !ok {
		return
	}
	h.unsubscribe(sess, ev.Channel)
}

func (h *Hub) handleGameEvent(gev GameEvent) {
	msgs := GameEventToWalter(gev)
	for _, msg := range msgs {
		ch, ok := h.channels[msg.Channel]
		if !ok {
			if msg.Channel != "" && len(msg.Channel) > 7 && msg.Channel[:7] == "planet:" {
				ch = NewChannel(msg.Channel, ChanLocal, 50)
				h.channels[msg.Channel] = ch
			} else if msg.Channel == "ops" {
				ch = h.channels["ops"]
			} else {
				continue
			}
		}

		ch.Backlog.Push(msg)

		if ch.Type == ChanOps {
			h.store.SaveChatMessage(store.ChatMsg{
				Channel:        msg.Channel,
				SenderFP:       msg.SenderFP,
				SenderCallsign: msg.SenderCallsign,
				Kind:           int(msg.Kind),
				Body:           msg.Body,
				CreatedAt:      msg.Timestamp,
			})
		}

		h.fanout(ch, msg)
	}
}

func (h *Hub) subscribe(sess *ChatSession, channelKey string) {
	ch, ok := h.channels[channelKey]
	if !ok {
		if len(channelKey) > 7 && channelKey[:7] == "planet:" {
			ch = NewChannel(channelKey, ChanLocal, 50)
			h.channels[channelKey] = ch
		} else {
			return
		}
	}

	ch.Members[sess.Fingerprint] = true
	sess.Subscriptions[channelKey] = true

	for _, msg := range ch.Backlog.All() {
		msg.IsBacklog = true
		h.sendToSession(sess, msg)
	}
}

func (h *Hub) unsubscribe(sess *ChatSession, channelKey string) {
	if ch, ok := h.channels[channelKey]; ok {
		delete(ch.Members, sess.Fingerprint)
	}
	delete(sess.Subscriptions, channelKey)

	if sess.ActiveChannel == channelKey {
		sess.ActiveChannel = "ops"
	}
}

func (h *Hub) fanout(ch *Channel, msg ChatMessage) {
	for fp := range ch.Members {
		if sess, ok := h.sessions[fp]; ok {
			if msg.SenderFP != "system" && sess.MuteList[msg.SenderFP] {
				continue
			}
			h.sendToSession(sess, msg)
		}
	}
}

func (h *Hub) sendToSession(sess *ChatSession, msg ChatMessage) {
	select {
	case sess.Outbox <- msg:
	default:
		select {
		case <-sess.Outbox:
		default:
		}
		select {
		case sess.Outbox <- msg:
		default:
		}
	}
}

func (h *Hub) getSession(fp string) *ChatSession {
	return h.sessions[fp]
}

func (h *Hub) getAllOnline() map[string]string {
	result := make(map[string]string, len(h.sessions))
	for fp, sess := range h.sessions {
		result[fp] = sess.Callsign
	}
	return result
}

func sanitizeMessage(s string) string {
	if len(s) > 500 {
		s = s[:500]
	}
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' || c == '\r' {
			buf = append(buf, ' ')
		} else if c < 32 && c != '\t' {
			// skip
		} else {
			buf = append(buf, c)
		}
	}
	return string(buf)
}

func splitFirst(s string, sep byte) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
