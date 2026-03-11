package views

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/world"
)

// StarMapStar represents a star on the map.
type StarMapStar struct {
	Address gamedata.GateAddress
	Name    string // planet name (or P-designation)
	WorldX  float64
	WorldY  float64
	Biome   gamedata.Biome
	Threat  int
	IsNamed bool
}

// StarMapState holds pan/zoom/cursor for the star map view.
type StarMapState struct {
	CamX, CamY float64 // center of viewport in world coords
	Zoom       int     // 0=normal, 1=zoomed in, 2=more zoomed
	Cursor     int     // index into stars slice
	Stars      []StarMapStar
	Current    int // index of current location star (-1 if none)
}

// NewStarMapState creates star map state from discovered addresses.
func NewStarMapState(addresses []gamedata.GateAddress, currentLocation string) *StarMapState {
	s := &StarMapState{
		Zoom:    0,
		Current: -1,
	}

	for i, addr := range addresses {
		star := buildStar(addr)
		s.Stars = append(s.Stars, star)

		if addr.Code() == currentLocation || (currentLocation == "sgc" && addr == gamedata.EarthAddress) {
			s.Current = i
		}
	}

	// Center camera on current location or Earth
	if s.Current >= 0 {
		s.CamX = s.Stars[s.Current].WorldX
		s.CamY = s.Stars[s.Current].WorldY
	} else if len(s.Stars) > 0 {
		s.CamX = s.Stars[0].WorldX
		s.CamY = s.Stars[0].WorldY
	}

	s.Cursor = s.Current
	if s.Cursor < 0 {
		s.Cursor = 0
	}

	return s
}

func buildStar(addr gamedata.GateAddress) StarMapStar {
	seed := addr.Seed()
	rng := rand.New(rand.NewSource(seed))

	// Position in world space — spread across a ~200x100 region
	// Earth is at roughly (100, 50); others scattered
	wx := rng.Float64()*180 + 10
	wy := rng.Float64()*80 + 10

	// Named planets get fixed iconic positions radiating from Earth
	name := gamedata.NameForAddress(addr)
	isNamed := name != ""
	if !isNamed {
		name = gamedata.PDesignation(addr)
	}

	biome, threat := world.BiomeForAddress(addr)

	// Override positions for named planets to form a constellation pattern
	if isNamed {
		switch name {
		case "Earth":
			wx, wy = 100, 50
		case "Abydos":
			wx, wy = 120, 35
		case "Chulak":
			wx, wy = 80, 30
		case "Tollana":
			wx, wy = 135, 55
		case "Cimmeria":
			wx, wy = 65, 60
		case "Dakara":
			wx, wy = 75, 75
		case "Langara":
			wx, wy = 130, 70
		case "Atlantis":
			wx, wy = 150, 40
		}
	}

	return StarMapStar{
		Address: addr,
		Name:    name,
		WorldX:  wx,
		WorldY:  wy,
		Biome:   biome,
		Threat:  threat,
		IsNamed: isNamed,
	}
}

// constellationLines defines which named planets are connected for the constellation overlay.
var constellationLines = [][2]string{
	{"Earth", "Abydos"},
	{"Earth", "Chulak"},
	{"Earth", "Cimmeria"},
	{"Earth", "Tollana"},
	{"Abydos", "Atlantis"},
	{"Chulak", "Dakara"},
	{"Tollana", "Langara"},
}

// RenderStarMap renders the astroterm-inspired gate network view.
func RenderStarMap(state *StarMapState, viewW, viewH int) string {
	if len(state.Stars) == 0 {
		return "\n  No discovered gate addresses.\n"
	}

	// Styles
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44AAFF")).Bold(true)
	bgStarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#1A1A2E"))
	bgStarBright := lipgloss.NewStyle().Foreground(lipgloss.Color("#2A2A3E"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Background(lipgloss.Color("#224488"))
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888899"))
	namedLabelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AABBCC")).Bold(true)
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#222244"))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	threatColors := []string{"#444444", "#558844", "#558844", "#888844", "#888844", "#AA6644", "#AA6644", "#CC4444", "#CC4444", "#FF2222"}

	// Zoom scaling: higher zoom = more spread
	scale := 1.0 + float64(state.Zoom)*0.5

	// Map area (leave room for title, info panel, help)
	mapH := viewH - 7
	if mapH < 10 {
		mapH = 10
	}
	mapW := viewW - 2
	if mapW < 20 {
		mapW = 20
	}

	// Build character grid
	grid := make([][]rune, mapH)
	colors := make([][]string, mapH)
	for y := range grid {
		grid[y] = make([]rune, mapW)
		colors[y] = make([]string, mapW)
		for x := range grid[y] {
			grid[y][x] = ' '
			colors[y][x] = ""
		}
	}

	// World → screen coordinate conversion
	toScreen := func(wx, wy float64) (int, int) {
		sx := int(math.Round((wx-state.CamX)*scale*2)) + mapW/2
		sy := int(math.Round((wy-state.CamY)*scale)) + mapH/2
		return sx, sy
	}

	// 1. Draw background stars (procedural, deterministic from screen position)
	bgSeed := int64(42)
	bgRng := rand.New(rand.NewSource(bgSeed))
	for i := 0; i < mapW*mapH/8; i++ {
		bx := bgRng.Intn(mapW)
		by := bgRng.Intn(mapH)
		if grid[by][bx] == ' ' {
			if bgRng.Float64() < 0.3 {
				grid[by][bx] = '·'
				colors[by][bx] = "bgbright"
			} else {
				grid[by][bx] = '.'
				colors[by][bx] = "bg"
			}
		}
	}

	// 2. Draw constellation lines between named planets
	namedPos := make(map[string][2]int) // name → screen pos
	for _, star := range state.Stars {
		if star.IsNamed {
			sx, sy := toScreen(star.WorldX, star.WorldY)
			namedPos[star.Name] = [2]int{sx, sy}
		}
	}
	for _, line := range constellationLines {
		p1, ok1 := namedPos[line[0]]
		p2, ok2 := namedPos[line[1]]
		if ok1 && ok2 {
			drawLineOnGrid(grid, colors, p1[0], p1[1], p2[0], p2[1], mapW, mapH)
		}
	}

	// 3. Place stars on grid
	for i, star := range state.Stars {
		sx, sy := toScreen(star.WorldX, star.WorldY)
		if sx < 0 || sx >= mapW || sy < 0 || sy >= mapH {
			continue
		}

		// Star glyph by threat
		var glyph rune
		switch {
		case star.Threat <= 2:
			glyph = '∗'
		case star.Threat <= 4:
			glyph = '✦'
		case star.Threat <= 6:
			glyph = '★'
		case star.Threat <= 8:
			glyph = '✹'
		default:
			glyph = '✵'
		}
		if star.IsNamed {
			glyph = '◉'
		}

		grid[sy][sx] = glyph

		// Color
		if i == state.Cursor {
			colors[sy][sx] = "cursor"
		} else if i == state.Current {
			colors[sy][sx] = "current"
		} else {
			tc := star.Threat - 1
			if tc < 0 {
				tc = 0
			}
			if tc >= len(threatColors) {
				tc = len(threatColors) - 1
			}
			colors[sy][sx] = "threat:" + threatColors[tc]
		}

		// Label (name next to star for named planets, always; P-designation only for cursor)
		if star.IsNamed || i == state.Cursor {
			label := star.Name
			lx := sx + 2
			if lx+len(label) >= mapW {
				lx = sx - len(label) - 1
			}
			for j, ch := range label {
				px := lx + j
				if px >= 0 && px < mapW && sy >= 0 && sy < mapH {
					if grid[sy][px] == ' ' || grid[sy][px] == '.' || grid[sy][px] == '·' {
						grid[sy][px] = ch
						if star.IsNamed {
							colors[sy][px] = "namedlabel"
						} else {
							colors[sy][px] = "label"
						}
					}
				}
			}
		}
	}

	// Render grid to string
	var b strings.Builder

	// Title
	title := "◎ S T A R G A T E   N E T W O R K ◎"
	pad := (viewW - lipgloss.Width(title)) / 2
	if pad < 0 {
		pad = 0
	}
	b.WriteString(strings.Repeat(" ", pad))
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	for y := 0; y < mapH; y++ {
		b.WriteString(" ")
		for x := 0; x < mapW; x++ {
			ch := string(grid[y][x])
			c := colors[y][x]
			switch {
			case c == "cursor":
				b.WriteString(cursorStyle.Render(ch))
			case c == "current":
				b.WriteString(currentStyle.Render(ch))
			case c == "bg":
				b.WriteString(bgStarStyle.Render(ch))
			case c == "bgbright":
				b.WriteString(bgStarBright.Render(ch))
			case c == "label":
				b.WriteString(labelStyle.Render(ch))
			case c == "namedlabel":
				b.WriteString(namedLabelStyle.Render(ch))
			case c == "line":
				b.WriteString(lineStyle.Render(ch))
			case len(c) > 7 && c[:7] == "threat:":
				color := c[7:]
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(ch))
			default:
				b.WriteString(ch)
			}
		}
		b.WriteString("\n")
	}

	// Info panel for selected star
	b.WriteString("\n")
	if state.Cursor >= 0 && state.Cursor < len(state.Stars) {
		star := state.Stars[state.Cursor]
		biomeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(star.Biome.Color))

		// Threat bar
		threatBar := ""
		for i := 0; i < 10; i++ {
			if i < star.Threat {
				threatBar += "█"
			} else {
				threatBar += "░"
			}
		}

		addr := star.Address.String()
		code := star.Address.Code()

		info := fmt.Sprintf(" %s  │  %s  │  %s  │  Threat: %s %d/10",
			infoStyle.Render(star.Name),
			biomeStyle.Render(star.Biome.Name),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(addr+" ("+code+")"),
			lipgloss.NewStyle().Foreground(lipgloss.Color(threatColors[clampInt(star.Threat-1, 0, 9)])).Render(threatBar),
			star.Threat,
		)
		b.WriteString(info)
		b.WriteString("\n")

		if state.Cursor == state.Current {
			b.WriteString(currentStyle.Render(" ► YOU ARE HERE"))
			b.WriteString("\n")
		}
	}

	// Help
	b.WriteString(helpStyle.Render(" ←→↑↓ pan  │  Tab/Shift+Tab select star  │  +/- zoom  │  Enter dial  │  Esc close"))
	b.WriteString("\n")

	return b.String()
}

// drawLineOnGrid draws a dotted constellation line between two screen points.
func drawLineOnGrid(grid [][]rune, colors [][]string, x0, y0, x1, y1, w, h int) {
	dx := x1 - x0
	dy := y1 - y0
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		return
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(x0) + t*float64(dx)))
		y := int(math.Round(float64(y0) + t*float64(dy)))
		if x >= 0 && x < w && y >= 0 && y < h {
			// Only draw on empty/background cells
			if grid[y][x] == ' ' || grid[y][x] == '.' {
				if i%3 == 0 { // dotted pattern
					grid[y][x] = '·'
					colors[y][x] = "line"
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
