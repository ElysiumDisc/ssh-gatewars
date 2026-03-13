package tui

// State represents the current screen in the TUI state machine.
type State int

const (
	StateSplash   State = iota // Animated splash screen
	StateCallsign              // Enter callsign
	StateAtlantis              // Personal hub — upgrades, stats
	StateThrone                // Ancient Chair upgrade terminal
	StateGalaxy                // Astroterm-style galaxy browser
	StateDefense               // Radial defense view — the game
)
