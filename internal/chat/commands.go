package chat

import (
	"strings"
	"time"
)

// handleSlashCommand dispatches a /command.
func (h *Hub) handleSlashCommand(ev ChatEvent) {
	sess, ok := h.sessions[ev.Fingerprint]
	if !ok {
		return
	}

	cmd := strings.ToLower(ev.Command)
	args := strings.TrimSpace(ev.Args)

	switch cmd {
	case "help":
		h.cmdHelp(sess)
	case "tune":
		h.cmdTune(sess, args)
	case "roster":
		h.cmdRoster(sess)
	case "who":
		h.cmdWho(sess)
	case "callsign":
		h.cmdCallsign(sess, args)
	case "me":
		h.cmdEmote(sess, args)
	case "dm":
		h.handleDM(sess, args)
	case "mute":
		h.cmdMute(sess, args)
	case "unmute":
		h.cmdUnmute(sess, args)
	case "motd":
		h.sendToSession(sess, WalterMsg("ops", MOTD()))
	case "clear":
		h.sendToSession(sess, WalterMsg("", "Chat cleared."))
	case "team":
		h.cmdTeam(sess, args)
	case "indeed":
		h.cmdFunEmote(sess, " raises an eyebrow. \"Indeed.\"")
	case "kree":
		h.cmdFunEmote(sess, " shouts \"KREE!\"")
	case "shol'va", "sholva":
		if args != "" {
			h.cmdFunEmote(sess, " points at "+args+". \"SHOL'VA!\"")
		} else {
			h.sendToSession(sess, WalterMsg("", "Usage: /shol'va <callsign>"))
		}
	default:
		h.sendToSession(sess, WalterMsg("", "Unknown command. Type /help for a list."))
	}
}

func (h *Hub) cmdHelp(sess *ChatSession) {
	help := `Commands:
  /help              Show this list
  /tune <channel>    Switch active channel (#ops, #planet, #sg-name)
  /roster            List all online players
  /who               List players on current channel
  /callsign <name>   Change your display name
  /me <action>       Emote action
  @<callsign> <msg>  Direct message
  /mute <callsign>   Mute a player
  /unmute <callsign> Unmute a player
  /motd              Show message of the day
  /clear             Clear chat panel
  /team create <n>   Create SG team
  /team invite <n>   Invite player to team
  /team leave        Leave your team
  /team kick <n>     Kick member (leader only)
  /team disband      Disband team (leader only)
  /indeed            "Indeed."
  /kree              "KREE!"`
	h.sendToSession(sess, WalterMsg("", help))
}

func (h *Hub) cmdTune(sess *ChatSession, args string) {
	if args == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /tune <channel> (e.g. #ops, #planet, #sg-myteam)"))
		return
	}

	target := strings.TrimPrefix(args, "#")

	if target == "ops" {
		if sess.Subscriptions["ops"] {
			sess.ActiveChannel = "ops"
			h.sendToSession(sess, WalterMsg("", "Tuned to #ops."))
		}
		return
	}

	// Planet channel
	for chKey := range sess.Subscriptions {
		if strings.HasPrefix(chKey, "planet:") && (target == "planet" || target == chKey[7:]) {
			sess.ActiveChannel = chKey
			h.sendToSession(sess, WalterMsg("", "Tuned to "+ChannelDisplayName(chKey)+"."))
			return
		}
	}

	// Team channel
	if strings.HasPrefix(target, "sg-") {
		teamKey := "team:" + target[3:]
		if sess.Subscriptions[teamKey] {
			sess.ActiveChannel = teamKey
			h.sendToSession(sess, WalterMsg("", "Tuned to #"+target+"."))
			return
		}
		h.sendToSession(sess, WalterMsg("", "You are not on that team's frequency."))
		return
	}

	h.sendToSession(sess, WalterMsg("", "Unknown channel: "+args))
}

func (h *Hub) cmdRoster(sess *ChatSession) {
	var lines []string
	lines = append(lines, "Online roster:")
	for _, s := range h.sessions {
		lines = append(lines, "  "+s.Callsign)
	}
	h.sendToSession(sess, WalterMsg("", strings.Join(lines, "\n")))
}

func (h *Hub) cmdWho(sess *ChatSession) {
	ch, ok := h.channels[sess.ActiveChannel]
	if !ok {
		h.sendToSession(sess, WalterMsg("", "No active channel."))
		return
	}

	var lines []string
	lines = append(lines, "Players on "+ChannelDisplayName(sess.ActiveChannel)+":")
	for fp := range ch.Members {
		if s, ok := h.sessions[fp]; ok {
			lines = append(lines, "  "+s.Callsign)
		}
	}
	h.sendToSession(sess, WalterMsg("", strings.Join(lines, "\n")))
}

func (h *Hub) cmdCallsign(sess *ChatSession, args string) {
	if args == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /callsign <name>"))
		return
	}

	name := args
	if len(name) < 2 || len(name) > 16 {
		h.sendToSession(sess, WalterMsg("", "Callsign must be 2-16 characters."))
		return
	}

	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '\'') {
			h.sendToSession(sess, WalterMsg("", "Callsign can only contain letters, numbers, _, -, '"))
			return
		}
	}

	lower := strings.ToLower(name)
	if lower == "walter" || lower == "system" || lower == "admin" {
		h.sendToSession(sess, WalterMsg("", "That callsign is reserved."))
		return
	}

	taken, err := h.store.IsCallsignTaken(name)
	if err == nil && taken {
		existing, _ := h.store.GetCallsign(sess.Fingerprint)
		if strings.ToLower(existing) != lower {
			h.sendToSession(sess, WalterMsg("", "That callsign is already taken."))
			return
		}
	}

	old := sess.Callsign
	sess.Callsign = name
	h.store.SetCallsign(sess.Fingerprint, name)

	h.fanout(h.channels["ops"], WalterMsg("ops",
		old+" is now known as "+name+"."))
}

func (h *Hub) cmdEmote(sess *ChatSession, args string) {
	if args == "" {
		return
	}
	ch, ok := h.channels[sess.ActiveChannel]
	if !ok {
		return
	}
	msg := ChatMessage{
		Channel: sess.ActiveChannel, SenderFP: sess.Fingerprint,
		SenderCallsign: sess.Callsign, Kind: MsgEmote,
		Body: "* " + sess.Callsign + " " + args, Timestamp: time.Now(),
	}
	ch.Backlog.Push(msg)
	h.fanout(ch, msg)
}

func (h *Hub) cmdFunEmote(sess *ChatSession, action string) {
	ch, ok := h.channels[sess.ActiveChannel]
	if !ok {
		return
	}
	msg := ChatMessage{
		Channel: sess.ActiveChannel, SenderFP: sess.Fingerprint,
		SenderCallsign: sess.Callsign, Kind: MsgEmote,
		Body: "* " + sess.Callsign + action, Timestamp: time.Now(),
	}
	ch.Backlog.Push(msg)
	h.fanout(ch, msg)
}

func (h *Hub) cmdMute(sess *ChatSession, args string) {
	if args == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /mute <callsign>"))
		return
	}
	fp, err := h.store.LookupFingerprint(args)
	if err != nil || fp == "" {
		h.sendToSession(sess, WalterMsg("", "Player not found: "+args))
		return
	}
	sess.MuteList[fp] = true
	h.store.AddMute(sess.Fingerprint, fp)
	h.sendToSession(sess, WalterMsg("", "Muted "+args+"."))
}

func (h *Hub) cmdUnmute(sess *ChatSession, args string) {
	if args == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /unmute <callsign>"))
		return
	}
	fp, err := h.store.LookupFingerprint(args)
	if err != nil || fp == "" {
		h.sendToSession(sess, WalterMsg("", "Player not found: "+args))
		return
	}
	delete(sess.MuteList, fp)
	h.store.RemoveMute(sess.Fingerprint, fp)
	h.sendToSession(sess, WalterMsg("", "Unmuted "+args+"."))
}

func (h *Hub) cmdTeam(sess *ChatSession, args string) {
	parts := splitFirst(args, ' ')
	sub := strings.ToLower(parts[0])
	subArgs := ""
	if len(parts) > 1 {
		subArgs = strings.TrimSpace(parts[1])
	}

	switch sub {
	case "create":
		h.teamCreate(sess, subArgs)
	case "invite":
		h.teamInvite(sess, subArgs)
	case "leave":
		h.teamLeave(sess)
	case "kick":
		h.teamKick(sess, subArgs)
	case "disband":
		h.teamDisband(sess)
	default:
		h.sendToSession(sess, WalterMsg("", "Usage: /team create|invite|leave|kick|disband"))
	}
}

func (h *Hub) teamCreate(sess *ChatSession, name string) {
	if name == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /team create <name>"))
		return
	}
	existing, _ := h.store.GetTeamByPlayer(sess.Fingerprint)
	if existing != nil {
		h.sendToSession(sess, WalterMsg("", "You are already on team "+existing.Name+". Leave first."))
		return
	}
	_, err := h.store.CreateTeam(name, sess.Fingerprint)
	if err != nil {
		h.sendToSession(sess, WalterMsg("", "Team name already taken or invalid."))
		return
	}
	chKey := TeamChannelKey(name)
	h.channels[chKey] = NewChannel(chKey, ChanTeam, 100)
	h.subscribe(sess, chKey)
	h.fanout(h.channels["ops"], WalterMsg("ops",
		"SG team \""+name+"\" has been formed by "+sess.Callsign+"."))
}

func (h *Hub) teamInvite(sess *ChatSession, callsign string) {
	if callsign == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /team invite <callsign>"))
		return
	}
	team, _ := h.store.GetTeamByPlayer(sess.Fingerprint)
	if team == nil {
		h.sendToSession(sess, WalterMsg("", "You are not on a team."))
		return
	}
	if team.LeaderFP != sess.Fingerprint {
		h.sendToSession(sess, WalterMsg("", "Only the team leader can invite."))
		return
	}
	targetFP, _ := h.store.LookupFingerprint(callsign)
	if targetFP == "" {
		h.sendToSession(sess, WalterMsg("", "Player not found: "+callsign))
		return
	}
	members, _ := h.store.GetTeamMembers(team.ID)
	if len(members) >= 4 {
		h.sendToSession(sess, WalterMsg("", "Team is full (max 4 members)."))
		return
	}
	h.store.AddTeamMember(team.ID, targetFP)
	if target, ok := h.sessions[targetFP]; ok {
		chKey := TeamChannelKey(team.Name)
		h.subscribe(target, chKey)
		h.sendToSession(target, WalterMsg("", "You have been added to SG team \""+team.Name+"\"."))
	}
	h.sendToSession(sess, WalterMsg("", callsign+" has been added to the team."))
}

func (h *Hub) teamLeave(sess *ChatSession) {
	team, _ := h.store.GetTeamByPlayer(sess.Fingerprint)
	if team == nil {
		h.sendToSession(sess, WalterMsg("", "You are not on a team."))
		return
	}
	h.store.RemoveTeamMember(team.ID, sess.Fingerprint)
	chKey := TeamChannelKey(team.Name)
	h.unsubscribe(sess, chKey)
	h.sendToSession(sess, WalterMsg("", "You have left SG team \""+team.Name+"\"."))
}

func (h *Hub) teamKick(sess *ChatSession, callsign string) {
	if callsign == "" {
		h.sendToSession(sess, WalterMsg("", "Usage: /team kick <callsign>"))
		return
	}
	team, _ := h.store.GetTeamByPlayer(sess.Fingerprint)
	if team == nil || team.LeaderFP != sess.Fingerprint {
		h.sendToSession(sess, WalterMsg("", "Only the team leader can kick members."))
		return
	}
	targetFP, _ := h.store.LookupFingerprint(callsign)
	if targetFP == "" {
		h.sendToSession(sess, WalterMsg("", "Player not found: "+callsign))
		return
	}
	h.store.RemoveTeamMember(team.ID, targetFP)
	if target, ok := h.sessions[targetFP]; ok {
		chKey := TeamChannelKey(team.Name)
		h.unsubscribe(target, chKey)
		h.sendToSession(target, WalterMsg("", "You have been removed from SG team \""+team.Name+"\"."))
	}
	h.sendToSession(sess, WalterMsg("", callsign+" has been kicked from the team."))
}

func (h *Hub) teamDisband(sess *ChatSession) {
	team, _ := h.store.GetTeamByPlayer(sess.Fingerprint)
	if team == nil || team.LeaderFP != sess.Fingerprint {
		h.sendToSession(sess, WalterMsg("", "Only the team leader can disband."))
		return
	}
	chKey := TeamChannelKey(team.Name)
	if ch, ok := h.channels[chKey]; ok {
		for fp := range ch.Members {
			if s, ok := h.sessions[fp]; ok {
				h.unsubscribe(s, chKey)
			}
		}
		delete(h.channels, chKey)
	}
	h.store.DisbandTeam(team.ID)
	h.fanout(h.channels["ops"], WalterMsg("ops",
		"SG team \""+team.Name+"\" has been disbanded."))
}
