package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/simulation"
)

// FogOfWar tracks which tiles a player has explored.
type FogOfWar struct {
	Width, Height int
	Explored      []bool
}

// NewFogOfWar creates a blank fog map.
func NewFogOfWar(w, h int) *FogOfWar {
	return &FogOfWar{
		Width:    w,
		Height:   h,
		Explored: make([]bool, w*h),
	}
}

// Reveal marks tiles visible from a position within a radius.
func (f *FogOfWar) Reveal(center core.Pos, radius int) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			x := center.X + dx
			y := center.Y + dy
			if x >= 0 && x < f.Width && y >= 0 && y < f.Height {
				f.Explored[y*f.Width+x] = true
			}
		}
	}
}

// IsExplored returns true if a tile has been seen.
func (f *FogOfWar) IsExplored(x, y int) bool {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return false
	}
	return f.Explored[y*f.Width+x]
}

// RenderPlanet renders the tile map viewport with entities.
func RenderPlanet(snap *simulation.PlanetSnapshot, playerPos core.Pos, fog *FogOfWar, viewW, viewH int, playerKey string) string {
	if snap == nil {
		return "Loading..."
	}

	// Camera centered on player
	camX := playerPos.X - viewW/2
	camY := playerPos.Y - viewH/2

	// Build entity lookup for fast rendering
	enemyMap := make(map[[2]int]simulation.EnemySnapshot)
	for _, e := range snap.Enemies {
		enemyMap[[2]int{e.Pos.X, e.Pos.Y}] = e
	}
	itemMap := make(map[[2]int]simulation.ItemSnapshot)
	for _, it := range snap.Items {
		itemMap[[2]int{it.Pos.X, it.Pos.Y}] = it
	}
	playerMap := make(map[[2]int]simulation.PlayerSnapshot)
	for _, p := range snap.Players {
		if p.Key != playerKey {
			playerMap[[2]int{p.Pos.X, p.Pos.Y}] = p
		}
	}

	fogStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	playerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	otherPlayerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44AAFF")).Bold(true)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA00")).Bold(true)
	gateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44AAFF")).Bold(true)

	var b strings.Builder

	for vy := 0; vy < viewH; vy++ {
		for vx := 0; vx < viewW; vx++ {
			wx := camX + vx
			wy := camY + vy

			// Player position
			if wx == playerPos.X && wy == playerPos.Y {
				b.WriteString(playerStyle.Render("@"))
				continue
			}

			// Out of map bounds
			if wx < 0 || wx >= snap.MapWidth || wy < 0 || wy >= snap.MapHeight {
				b.WriteString(" ")
				continue
			}

			// Unexplored
			if !fog.IsExplored(wx, wy) {
				b.WriteString(" ")
				continue
			}

			// Check for entities (only show if in visible range)
			inView := (wx-playerPos.X)*(wx-playerPos.X)+(wy-playerPos.Y)*(wy-playerPos.Y) <= 64 // ~8 tile radius

			if inView {
				if e, ok := enemyMap[[2]int{wx, wy}]; ok {
					def := gamedata.Enemies[e.DefID]
					eStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(def.Color))
					b.WriteString(eStyle.Render(string(def.Glyph)))
					continue
				}
				if _, ok := playerMap[[2]int{wx, wy}]; ok {
					b.WriteString(otherPlayerStyle.Render("@"))
					continue
				}
				if _, ok := itemMap[[2]int{wx, wy}]; ok {
					b.WriteString(itemStyle.Render("!"))
					continue
				}
			}

			// Tile
			tileType := snap.Tiles[wy*snap.MapWidth+wx]
			info := gamedata.Tiles[tileType]

			if tileType == gamedata.TileStargate {
				b.WriteString(gateStyle.Render(string(info.Glyph)))
			} else if inView {
				tStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(info.Color))
				b.WriteString(tStyle.Render(string(info.Glyph)))
			} else {
				// Explored but not in current view range — dim
				b.WriteString(fogStyle.Render(string(info.Glyph)))
			}
		}
		if vy < viewH-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
