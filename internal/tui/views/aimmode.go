package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/simulation"
	"ssh-gatewars/internal/world"
)

// RenderAimOverlay renders the aim mode targeting overlay on top of the planet view.
// It draws a targeting line from the player to the reticle and highlights valid targets.
func RenderAimOverlay(
	snap *simulation.PlanetSnapshot,
	playerPos, aimTarget core.Pos,
	fog *FogOfWar,
	viewW, viewH int,
	playerKey string,
	weaponRange int,
) string {
	if snap == nil {
		return "No target data..."
	}

	// Compute viewport offset (centered on player)
	offX := playerPos.X - viewW/2
	offY := playerPos.Y - viewH/2

	// Build the LOS line
	losLine := world.BresenhamLine(playerPos, aimTarget)
	losSet := make(map[core.Pos]bool)
	blocked := false
	for i, p := range losLine {
		if i > 0 && isOpaqueTile(snap, p) {
			blocked = true
		}
		losSet[p] = !blocked
	}

	// Check range
	dist := playerPos.ManhattanDist(aimTarget)
	inRange := dist <= weaponRange

	// Build the tile grid (same as RenderPlanet but with aim overlay)
	var b strings.Builder

	reticleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
	losGoodStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#44FF44"))
	losBadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))

	for vy := 0; vy < viewH; vy++ {
		for vx := 0; vx < viewW; vx++ {
			wx := vx + offX
			wy := vy + offY
			p := core.Pos{X: wx, Y: wy}

			// Reticle position
			if p == aimTarget {
				if inRange && !blocked {
					b.WriteString(reticleStyle.Render("X"))
				} else {
					b.WriteString(losBadStyle.Render("X"))
				}
				continue
			}

			// LOS line
			if clear, onLine := losSet[p]; onLine && p != playerPos {
				if clear {
					b.WriteString(losGoodStyle.Render("·"))
				} else {
					b.WriteString(losBadStyle.Render("·"))
				}
				continue
			}

			// Player
			if p == playerPos {
				b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("@"))
				continue
			}

			// Normal tile rendering (dimmed in aim mode)
			if wx < 0 || wy < 0 || wx >= snap.MapWidth || wy >= snap.MapHeight {
				b.WriteString(" ")
				continue
			}

			idx := wy*snap.MapWidth + wx
			if fog != nil && !fog.IsExplored(wx, wy) {
				b.WriteString(" ")
				continue
			}

			// Check for enemies
			foundEnemy := false
			for _, en := range snap.Enemies {
				if en.Pos == p {
					eDef := gamedata.Enemies[en.DefID]
					b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(eDef.Color)).Render(string(eDef.Glyph)))
					foundEnemy = true
					break
				}
			}
			if foundEnemy {
				continue
			}

			tile := snap.Tiles[idx]
			info := gamedata.Tiles[tile]
			b.WriteString(dimStyle.Render(string(info.Glyph)))
		}
		if vy < viewH-1 {
			b.WriteString("\n")
		}
	}

	// Status line
	status := fmt.Sprintf(" AIM MODE | Range: %d/%d | ", dist, weaponRange)
	if !inRange {
		status += "OUT OF RANGE"
	} else if blocked {
		status += "LOS BLOCKED"
	} else {
		status += "CLEAR — Enter to fire"
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAA44")).Render(status))

	return b.String()
}

func isOpaqueTile(snap *simulation.PlanetSnapshot, p core.Pos) bool {
	if p.X < 0 || p.Y < 0 || p.X >= snap.MapWidth || p.Y >= snap.MapHeight {
		return true
	}
	tile := snap.Tiles[p.Y*snap.MapWidth+p.X]
	info, ok := gamedata.Tiles[tile]
	if !ok {
		return true
	}
	return info.Opaque
}
