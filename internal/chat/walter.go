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
			ev.Callsign+" has connected to Atlantis."))

	case GamePlayerDisconnect:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" has signed off."))

	case GamePlayerDeploy:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" deployed a control chair on "+ev.PlanetName+"."))
		ch := PlanetChannelKey(ev.PlanetName)
		msgs = append(msgs, WalterMsg(ch,
			"Chair deployed — "+ev.Callsign+" is online."))

	case GamePlayerRetreat:
		msgs = append(msgs, WalterMsg("ops",
			ev.Callsign+" retreated from "+ev.PlanetName+"."))

	case GamePlanetLiberated:
		msgs = append(msgs, WalterAnnounce(
			">>> PLANET "+ev.PlanetName+" HAS BEEN LIBERATED! <<<"))

	case GamePlanetFailed:
		msgs = append(msgs, WalterMsg("ops",
			"Defense of "+ev.PlanetName+" has failed. The replicators hold."))

	case GameTeamCreated:
		msgs = append(msgs, WalterMsg("ops",
			"SG team \""+ev.Extra+"\" has been formed."))

	case GameTeamDisbanded:
		msgs = append(msgs, WalterMsg("ops",
			"SG team \""+ev.Extra+"\" has been disbanded."))

	case GameSurgeStart:
		msgs = append(msgs, WalterAnnounce(
			"⚠ REPLICATOR SURGE on "+ev.PlanetName+"! Double spawns — double ZPM rewards!"))

	case GameSurgeEnd:
		msgs = append(msgs, WalterMsg("ops",
			"Replicator surge on "+ev.PlanetName+" has subsided."))

	case GameMilestone:
		msgs = append(msgs, WalterAnnounce(
			">>> "+ev.Extra+" <<<"))

	case GameGalaxyReset:
		msgs = append(msgs, WalterAnnounce(
			">>> GALAXY LIBERATED — NEW THREAT CYCLE BEGINS! <<<"))
		msgs = append(msgs, WalterMsg("ops",
			"The replicators have adapted. All planets are under new invasion. Difficulty increased."))
	}

	return msgs
}

// MOTD returns the message-of-the-day for the ops channel.
func MOTD() string {
	return `    ╔═══════════════════════════════════════════════╗
    ║       ___ _____ ___   _      ___   ___  ___  ║
    ║      / __|_   _| __| | |    | \ \ / / |/ __| ║
    ║     | (_ | | | | _|  | | /| | |\ V /| |\__ \ ║
    ║      \___| |_| |___| |_|/ |_|  |_|  |_||___/ ║
    ║                                               ║
    ║   ATLANTIS COMMAND — COMMS ONLINE             ║
    ║   ANCIENT DEFENSE NETWORK ACTIVE              ║
    ║                                               ║
    ║   The replicators are invading.               ║
    ║   Deploy your chair. Launch your drones.      ║
    ║   Hold the galaxy.                            ║
    ║                                               ║
    ║   Type /help for commands                     ║
    ╚═══════════════════════════════════════════════╝`
}
