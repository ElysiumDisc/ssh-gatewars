package render

// Viewport maps world coordinates to terminal screen coordinates.
type Viewport struct {
	Width   int // terminal columns available for battlefield
	Height  int // terminal rows available for battlefield (minus HUD)
	WorldW  int
	WorldH  int
}

// NewViewport creates a viewport. hudRows are reserved at the bottom.
func NewViewport(termW, termH, worldW, worldH, hudRows int) *Viewport {
	h := termH - hudRows
	if h < 1 {
		h = 1
	}
	return &Viewport{
		Width:  termW,
		Height: h,
		WorldW: worldW,
		WorldH: worldH,
	}
}

// WorldToScreen converts world-space coordinates to screen-space.
func (v *Viewport) WorldToScreen(wx, wy float64) (sx, sy int, visible bool) {
	sx = int(wx * float64(v.Width) / float64(v.WorldW))
	sy = int(wy * float64(v.Height) / float64(v.WorldH))
	visible = sx >= 0 && sx < v.Width && sy >= 0 && sy < v.Height
	return
}

// ScreenToWorld converts screen coordinates back to world-space.
func (v *Viewport) ScreenToWorld(sx, sy int) (wx, wy int) {
	wx = sx * v.WorldW / v.Width
	wy = sy * v.WorldH / v.Height
	return
}
