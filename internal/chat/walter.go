package chat

import "time"

const walterFP = "system"
const walterName = "WALTER"
const walterPrefix = "\u258c" + "WALTER" + "\u2590" + " " // ▌WALTER▐

// WalterMsg creates a system message from Walter.
func WalterMsg(channel, body string) ChatMessage {
	return ChatMessage{
		Channel:        channel,
		SenderFP:       walterFP,
		SenderCallsign: walterName,
		Kind:           MsgSystem,
		Body:           walterPrefix + body,
		Timestamp:      time.Now(),
	}
}

// WalterAnnounce creates a server-wide announcement.
func WalterAnnounce(body string) ChatMessage {
	return ChatMessage{
		Channel:        "ops",
		SenderFP:       walterFP,
		SenderCallsign: walterName,
		Kind:           MsgAnnounce,
		Body:           body,
		Timestamp:      time.Now(),
	}
}

// GameEventToWalter converts a game event into Walter chat messages.
func GameEventToWalter(ev GameEvent) []ChatMessage {
	var msgs []ChatMessage

	switch ev.Type {
	case GamePlayerConnect:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" has reported for duty."))

	case GamePlayerDisconnect:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" has signed off."))

	case GameGateDial:
		if ev.PlanetSeed != "" {
			ch := LocalChannelKey(ev.PlanetSeed)
			msgs = append(msgs,
				WalterMsg(ch, "Chevron 1... encoded."),
				WalterMsg(ch, "Chevron 2... encoded."),
				WalterMsg(ch, "Chevron 3... encoded."),
				WalterMsg(ch, "Chevron 4... encoded."),
				WalterMsg(ch, "Chevron 5... encoded."),
				WalterMsg(ch, "Chevron 6... encoded."),
				WalterMsg(ch, "Chevron 7... locked!"),
			)
		}

	case GamePlayerArrived:
		if ev.PlanetSeed != "" {
			ch := LocalChannelKey(ev.PlanetSeed)
			msgs = append(msgs, WalterMsg(ch,
				"Incoming traveler — "+ev.Callsign+"."))
		}

	case GamePlayerDeparted:
		if ev.PlanetSeed != "" {
			ch := LocalChannelKey(ev.PlanetSeed)
			msgs = append(msgs, WalterMsg(ch,
				ev.Callsign+" has departed through the gate."))
		}

	case GamePlayerKilled:
		msgs = append(msgs, WalterMsg("ops",
			"We've lost "+ev.Callsign+"'s signal on "+ev.PlanetName+"."))

	case GamePlayerLevelUp:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" has been promoted to Level "+ev.Extra+"."))

	case GameEnemyBossKilled:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" has defeated "+ev.Extra+" on "+ev.PlanetName+"!"))

	case GameTeamCreated:
		msgs = append(msgs, WalterMsg("ops",
			"SG team \""+ev.Extra+"\" has been formed."))

	case GameTeamDisbanded:
		msgs = append(msgs, WalterMsg("ops",
			"SG team \""+ev.Extra+"\" has been disbanded."))
	}

	return msgs
}

// MOTD returns the message-of-the-day art for the ops channel.
func MOTD() string {
	return `    ╔══════════════════════════════════════════════╗
    ║     _____ _____  _____                       ║
    ║    / ____|/ ____||  __ \                      ║
    ║   | (___ | |  __ | |    |                     ║
    ║    \___ \| | |_ || |    |                     ║
    ║    ____) | |__| || |____|                     ║
    ║   |_____/ \_____||______|  COMMS ONLINE       ║
    ║                                              ║
    ║   Stargate Command — Secure Channel          ║
    ║   CHEYENNE MOUNTAIN COMPLEX                  ║
    ║   Classification: TOP SECRET / SCI           ║
    ║                                              ║
    ║   Type /help for commands                    ║
    ║   Type /tune #channel to switch frequency    ║
    ╚══════════════════════════════════════════════╝`
}
