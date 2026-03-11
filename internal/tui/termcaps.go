package tui

import "github.com/muesli/termenv"

// ColorMode describes the color rendering capability of a terminal.
type ColorMode int

const (
	ColorBasic    ColorMode = iota // no color (rare)
	Color256                       // ANSI 256-color
	ColorTrueColor                 // 24-bit RGB
)

// TermCaps describes a terminal's rendering capabilities.
type TermCaps struct {
	Width   int
	Height  int
	Color   ColorMode
	Unicode bool // supports Unicode box drawing and symbols
}

// DetectCaps determines terminal capabilities from a termenv profile.
func DetectCaps(profile termenv.Profile, width, height int) TermCaps {
	caps := TermCaps{
		Width:   width,
		Height:  height,
		Unicode: true, // assume Unicode support by default
	}

	switch profile {
	case termenv.TrueColor:
		caps.Color = ColorTrueColor
	case termenv.ANSI256:
		caps.Color = Color256
	case termenv.ANSI:
		caps.Color = Color256 // treat 16-color as 256
	default:
		caps.Color = ColorBasic
	}

	return caps
}
