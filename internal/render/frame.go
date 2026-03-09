package render

// HalfBlockCell represents one terminal character in half-block rendered output.
// Kept for future battle view animations.
type HalfBlockCell struct {
	Char string
	FG   string
	BG   string
	Bold bool
}

// PixelBuffer holds a doubled-resolution pixel grid.
// Kept for future battle view animations.
type PixelBuffer struct {
	Width, Height int
	pixels        []string
}

// NewPixelBuffer creates a pixel buffer.
func NewPixelBuffer(w, h int) *PixelBuffer {
	return &PixelBuffer{
		Width:  w,
		Height: h,
		pixels: make([]string, w*h),
	}
}

func (pb *PixelBuffer) Set(x, y int, color string) {
	if x >= 0 && x < pb.Width && y >= 0 && y < pb.Height {
		pb.pixels[y*pb.Width+x] = color
	}
}

func (pb *PixelBuffer) Get(x, y int) string {
	if x >= 0 && x < pb.Width && y >= 0 && y < pb.Height {
		if c := pb.pixels[y*pb.Width+x]; c != "" {
			return c
		}
	}
	return "#000000"
}
