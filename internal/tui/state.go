package tui

// ViewState identifies the current TUI view.
type ViewState int

const (
	ViewSplash     ViewState = iota // title screen
	ViewCallSign                    // name entry
	ViewSGC                         // SGC hub (same renderer as planet)
	ViewDHD                         // gate dialing interface
	ViewPlanet                      // planet exploration
	ViewInventory                   // inventory modal
	ViewAddressBook                 // known gate addresses
	ViewHelp                        // help overlay
	ViewDead                        // death screen
	ViewPlayerList                  // online player list modal
	ViewAimMode                     // targeting reticle for ranged fire
	ViewStarMap                     // astroterm-inspired gate network browser
)

// FocusTarget determines where keyboard input goes.
type FocusTarget int

const (
	FocusGame FocusTarget = iota // movement, interact, menus
	FocusChat                    // typing in chat input
)

// ParentState returns the parent view for Esc navigation.
func ParentState(s ViewState) ViewState {
	switch s {
	case ViewInventory, ViewAddressBook, ViewHelp:
		return ViewPlanet // or SGC, handled in model
	case ViewDHD:
		return ViewSGC
	case ViewDead:
		return ViewSGC
	default:
		return s
	}
}
