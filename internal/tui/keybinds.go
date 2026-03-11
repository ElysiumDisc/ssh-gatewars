package tui

import tea "github.com/charmbracelet/bubbletea"

// KeyAction is a semantic action mapped from a key press.
type KeyAction int

const (
	KeyNone KeyAction = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyConfirm
	KeyCancel
	KeyQuit
	KeyInteract
	KeyDial        // 'g' — dial gate
	KeyInventory   // 'i'
	KeyAddressBook // 'a'
	KeyChat        // 'c'
	KeyPlayerList  // Tab
	KeyHelp        // '?'
	KeyPickup      // 'e' (same as interact)
	KeyUse         // 'u'
	KeyEquip       // 'E'
	KeyDrop        // 'd'
	KeyFire        // 'f' — enter aim mode / fire
	KeyReload      // 'r' — reload weapon
	KeyStarMap     // 'm' — open star map
)

// MapKey translates a bubbletea key message to a KeyAction.
func MapKey(msg tea.KeyMsg) KeyAction {
	switch msg.String() {
	case "up", "k", "w":
		return KeyUp
	case "down", "j", "s":
		return KeyDown
	case "left", "h", "a":
		return KeyLeft
	case "right", "l", "d":
		return KeyRight
	case "enter", " ":
		return KeyConfirm
	case "esc":
		return KeyCancel
	case "q", "ctrl+c":
		return KeyQuit
	case "e":
		return KeyInteract
	case "g":
		return KeyDial
	case "i":
		return KeyInventory
	case "A": // capital A for address book (lowercase 'a' is left)
		return KeyAddressBook
	case "c":
		return KeyChat
	case "tab":
		return KeyPlayerList
	case "?":
		return KeyHelp
	case "u":
		return KeyUse
	case "E":
		return KeyEquip
	case "D": // capital D for drop (lowercase 'd' is right)
		return KeyDrop
	case "f":
		return KeyFire
	case "r":
		return KeyReload
	case "m":
		return KeyStarMap
	default:
		return KeyNone
	}
}
