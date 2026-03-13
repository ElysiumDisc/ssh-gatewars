package chat

import (
	"fmt"
	"sort"
	"time"
)

// MessageKind classifies a chat message.
type MessageKind int

const (
	MsgChat     MessageKind = iota // normal player message
	MsgSystem                      // Walter system message
	MsgEmote                       // /me action
	MsgWhisper                     // DM
	MsgAnnounce                    // server-wide alert
)

// ChatMessage is a rendered message delivered to a session.
type ChatMessage struct {
	Channel        string
	SenderFP       string
	SenderCallsign string
	Kind           MessageKind
	Body           string
	Timestamp      time.Time
	IsBacklog      bool // true if from backlog delivery
}

// ChannelType classifies channels.
type ChannelType int

const (
	ChanOps   ChannelType = iota // #ops — global
	ChanLocal                    // #planet:<id> — planet-scoped
	ChanTeam                     // #sg-<name> — team
	ChanDM                       // dm:<fp1>:<fp2> — direct message
)

// ChatEventType tags events sent to the hub.
type ChatEventType int

const (
	EventConnect      ChatEventType = iota // player connected
	EventDisconnect                        // player disconnected
	EventSendMessage                       // player sent a chat message
	EventJoinChannel                       // player joins a channel
	EventLeaveChannel                      // player leaves a channel
	EventSlashCommand                      // /command issued
)

// ChatEvent is sent from sessions to the hub.
type ChatEvent struct {
	Type        ChatEventType
	Fingerprint string
	Callsign    string
	Channel     string          // channel key
	Body        string          // message body or command args
	Command     string          // slash command name
	Args        string          // slash command arguments
	Outbox      chan ChatMessage // only for EventConnect
}

// GameEventType identifies engine → chat events.
type GameEventType int

const (
	GamePlayerConnect    GameEventType = iota
	GamePlayerDisconnect
	GamePlayerDeploy
	GamePlayerRetreat
	GamePlanetLiberated
	GamePlanetFailed
	GameTeamCreated
	GameTeamDisbanded
	GameSurgeStart
	GameSurgeEnd
	GameMilestone
	GameGalaxyReset
)

// GameEvent is sent from the engine to the chat hub.
type GameEvent struct {
	Type       GameEventType
	Callsign   string
	PlanetName string
	Extra      string
}

// DMChannelKey returns a deterministic DM channel key for two fingerprints.
func DMChannelKey(fp1, fp2 string) string {
	if fp1 > fp2 {
		fp1, fp2 = fp2, fp1
	}
	return fmt.Sprintf("dm:%s:%s", fp1, fp2)
}

// PlanetChannelKey returns a planet-local channel key.
func PlanetChannelKey(planetName string) string {
	return "planet:" + planetName
}

// TeamChannelKey returns a team channel key.
func TeamChannelKey(teamName string) string {
	return "team:" + teamName
}

// ChannelDisplayName returns a user-friendly name for a channel key.
func ChannelDisplayName(key string) string {
	if key == "ops" {
		return "#ops"
	}
	if len(key) > 7 && key[:7] == "planet:" {
		return "#" + key[7:]
	}
	if len(key) > 5 && key[:5] == "team:" {
		return "#sg-" + key[5:]
	}
	if len(key) > 3 && key[:3] == "dm:" {
		return "[DM]"
	}
	return key
}

// SortedKeys returns map keys sorted alphabetically.
func SortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
