package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/game"
	"ssh-gatewars/internal/store"
)

// ThroneModel holds state for the upgrade throne view.
type ThroneModel struct {
	Selected    int    // 0=chair, 1=swift, 2=blast, 3=piercing, 4=faction toggle, 5=reset
	StatusMsg   string // feedback after action
	StatusStyle lipgloss.Style
	StatusTTL   int // frames remaining to show status
}

const throneOptionCount = 6 // chair, 3 drone tiers, faction, reset

func NewThroneModel() ThroneModel {
	return ThroneModel{}
}

func (t *ThroneModel) MoveSelection(delta int) {
	t.Selected += delta
	if t.Selected < 0 {
		t.Selected = 0
	}
	if t.Selected >= throneOptionCount {
		t.Selected = throneOptionCount - 1
	}
}

func (t *ThroneModel) SetStatus(msg string, style lipgloss.Style) {
	t.StatusMsg = msg
	t.StatusStyle = style
	t.StatusTTL = 45
}

func (t *ThroneModel) Tick() {
	if t.StatusTTL > 0 {
		t.StatusTTL--
		if t.StatusTTL == 0 {
			t.StatusMsg = ""
		}
	}
}

// ── Upgrade costs ─────────────────────────────────────────────────────

const MaxChairLevel = 10

func ChairUpgradeCost(currentLevel int) int {
	return (currentLevel + 1) * 50
}

func DroneTierCost(tier int) int {
	switch tier {
	case 1:
		return 100
	case 2:
		return 250
	case 3:
		return 500
	}
	return 0
}

var droneTierInfo = []struct {
	Name   string
	Symbol string
	Desc   string
}{
	{"Standard", "✸", "Base drone — reliable, single target"},
	{"Swift", "✦", "1.5x speed — fast intercept"},
	{"Blast", "✸", "2x damage — splash radius 2.0"},
	{"Piercing", "►", "3x damage — passes through targets"},
}

// RenderThrone draws the upgrade throne view.
func RenderThrone(player *store.PlayerRecord, throne ThroneModel, frameCount, w, h int) string {
	if player == nil {
		return Center(StyleDim.Render("Loading commander data..."), w, h)
	}

	faction := game.Faction(player.Faction)
	factionDef := game.FactionDefs[faction]

	// ── Layout ─────────────────────────────────────────────────
	innerW := w - 6
	leftW := innerW * 38 / 100
	if leftW < 30 {
		leftW = 30
	}
	rightW := innerW - leftW - 2

	// ── Left panel: chair art + stats + power preview ──────────
	var leftLines []string
	leftLines = append(leftLines, "")

	// Faction badge
	factionStyle := StyleCyan
	if faction == game.FactionOri {
		factionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600")).Bold(true)
	}
	leftLines = append(leftLines, "  "+factionStyle.Render("◆ "+factionDef.Tag+" PATH"))
	leftLines = append(leftLines, "  "+StyleDim.Render(factionDef.Desc))
	leftLines = append(leftLines, "")

	// Throne art with pulsing glow
	pulse := (frameCount / 10) % 2
	glowStyle := StyleCyan
	if faction == game.FactionOri {
		if pulse == 0 {
			glowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6600"))
		} else {
			glowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CC4400"))
		}
	} else {
		if pulse == 1 {
			glowStyle = StyleCyanDim
		}
	}

	chairArt := []string{
		StyleCyanDim.Render("       ╔═════════╗"),
		StyleCyanDim.Render("       ║") + glowStyle.Render("  ◎   ◎ ") + StyleCyanDim.Render("║"),
		StyleCyanDim.Render("       ║") + StyleGold.Render("    ⌂   ") + StyleCyanDim.Render("║"),
		StyleCyanDim.Render("   ════╩═════════╩════"),
		StyleCyanDim.Render("   ╚═══════════════════╝"),
	}
	leftLines = append(leftLines, chairArt...)
	leftLines = append(leftLines, "")

	// ZPM balance
	leftLines = append(leftLines, "  "+StyleGold.Render(fmt.Sprintf("⚡ ZPM: %d", player.ZPMBalance)))
	leftLines = append(leftLines, "")

	// Current stats
	maxDrones := game.CalcMaxDronesFaction(faction, player.ChairLevel)
	maxShield := game.CalcMaxShieldFaction(faction, player.ChairLevel)
	salvo := game.CalcSalvoCount(faction, player.ChairLevel)
	fireRate := game.CalcEffectiveFireRate(faction, player.ChairLevel, 10)
	leftLines = append(leftLines,
		"  "+StyleBright.Render("Chair Level")+StyleDim.Render(": ")+fmtInt(player.ChairLevel, StyleCyan),
		"  "+StyleBright.Render("Max Drones")+StyleDim.Render(":  ")+fmtInt(maxDrones, StyleGold),
		"  "+StyleBright.Render("Salvo")+StyleDim.Render(":       ")+fmtInt(salvo, StyleGold)+StyleDim.Render("/shot"),
		"  "+StyleBright.Render("Fire Rate")+StyleDim.Render(":   ")+StyleMid.Render(fmt.Sprintf("%.1fs", float64(fireRate)/10.0)),
		"  "+StyleBright.Render("Max Shield")+StyleDim.Render(":  ")+fmtInt(maxShield, StyleMid),
		"  "+StyleBright.Render("Drone")+StyleDim.Render(":       ")+droneColorName(player.DroneTier),
	)
	leftLines = append(leftLines, "")

	// Power preview — next level
	if player.ChairLevel < MaxChairLevel {
		nextLv := player.ChairLevel + 1
		nextDrones := game.CalcMaxDronesFaction(faction, nextLv)
		nextSalvo := game.CalcSalvoCount(faction, nextLv)
		nextFR := game.CalcEffectiveFireRate(faction, nextLv, 10)
		nextShield := game.CalcMaxShieldFaction(faction, nextLv)
		leftLines = append(leftLines, "  "+StyleCyanDim.Render("── NEXT LEVEL ──"))
		leftLines = append(leftLines,
			"  "+StyleDim.Render(fmt.Sprintf("Drones: %d→%d  Salvo: %d→%d", maxDrones, nextDrones, salvo, nextSalvo)),
		)
		leftLines = append(leftLines,
			"  "+StyleDim.Render(fmt.Sprintf("Fire: %.1fs→%.1fs  Shield: %d→%d", float64(fireRate)/10.0, float64(nextFR)/10.0, maxShield, nextShield)),
		)
	}

	// Status message
	if throne.StatusMsg != "" {
		leftLines = append(leftLines, "")
		leftLines = append(leftLines, "  "+throne.StatusStyle.Render(throne.StatusMsg))
	}

	leftContent := strings.Join(leftLines, "\n")
	leftStyled := lipgloss.NewStyle().Width(leftW).Render(leftContent)

	// ── Right panel: upgrade options ───────────────────────────
	var rightLines []string
	rightLines = append(rightLines, "")
	rightLines = append(rightLines, " "+StyleCyanDim.Render("CHAIR SYSTEMS"))
	rightLines = append(rightLines, "")

	// [0] Chair upgrade
	chairCost := ChairUpgradeCost(player.ChairLevel)
	chairMaxed := player.ChairLevel >= MaxChairLevel

	prefix := selPrefix(throne.Selected == 0)
	if chairMaxed {
		rightLines = append(rightLines,
			prefix+" "+StyleBright.Render("Shield Generator")+StyleDim.Render("  MAX"))
	} else {
		costStyle := costColor(player.ZPMBalance >= chairCost)
		lvl := fmt.Sprintf("Lv%d → Lv%d", player.ChairLevel, player.ChairLevel+1)
		rightLines = append(rightLines,
			prefix+" "+StyleBright.Render("Shield Generator")+StyleDim.Render("  ")+StyleMid.Render(lvl))
		rightLines = append(rightLines,
			"    "+StyleDim.Render("Cost: ")+costStyle.Render(fmt.Sprintf("⚡%d", chairCost))+
				StyleDim.Render(fmt.Sprintf("  +%d drones, +%d shield", factionDef.MaxDronesPerLv, 5)))
	}

	rightLines = append(rightLines, "")
	rightLines = append(rightLines, " "+StyleCyanDim.Render("DRONE WEAPONS"))
	rightLines = append(rightLines, "")

	// [1-3] Drone tiers
	for tier := 0; tier < 4; tier++ {
		info := droneTierInfo[tier]
		isCurrent := player.DroneTier == tier
		isUpgrade := tier > 0

		prefix = "  "
		if isUpgrade && throne.Selected == tier {
			prefix = StyleGold.Render(" ▸")
		}

		symbol := DroneStyle(tier).Render(info.Symbol)
		name := DroneStyle(tier).Render(info.Name)

		// Faction damage preview
		baseDmg := game.DroneTiers[game.DroneTier(tier)].Damage
		effDmg := game.CalcDroneDamage(faction, baseDmg)

		if isCurrent {
			rightLines = append(rightLines,
				prefix+" "+symbol+" "+name+StyleSuccess.Render("  (equipped)")+StyleDim.Render(fmt.Sprintf(" dmg:%d", effDmg)))
		} else if !isUpgrade {
			rightLines = append(rightLines,
				prefix+" "+symbol+" "+name+StyleDim.Render(fmt.Sprintf("  dmg:%d", effDmg)))
		} else {
			cost := DroneTierCost(tier)
			cs := costColor(player.ZPMBalance >= cost)
			rightLines = append(rightLines,
				prefix+" "+symbol+" "+name+StyleDim.Render("  ")+cs.Render(fmt.Sprintf("⚡%d", cost))+StyleDim.Render(fmt.Sprintf(" dmg:%d", effDmg)))
		}
		rightLines = append(rightLines, "    "+StyleDim.Render(info.Desc))
		rightLines = append(rightLines, "")
	}

	// [4] Faction toggle
	rightLines = append(rightLines, " "+StyleCyanDim.Render("PATH"))
	rightLines = append(rightLines, "")
	prefix = selPrefix(throne.Selected == 4)
	otherFaction := game.FactionAncient
	if faction == game.FactionAncient {
		otherFaction = game.FactionOri
	}
	otherDef := game.FactionDefs[otherFaction]
	rightLines = append(rightLines,
		prefix+" "+StyleBright.Render("Switch to "+otherDef.Name+" Path"))
	rightLines = append(rightLines,
		"    "+StyleDim.Render(otherDef.Desc))
	rightLines = append(rightLines,
		"    "+StyleDanger.Render("Resets all upgrades!"))
	rightLines = append(rightLines, "")

	// [5] Reset
	prefix = selPrefix(throne.Selected == 5)
	rightLines = append(rightLines,
		prefix+" "+StyleDanger.Render("Reset All Progress"))
	rightLines = append(rightLines,
		"    "+StyleDim.Render("ZPM, chair level, drone tier → 0"))
	rightLines = append(rightLines, "")

	rightContent := strings.Join(rightLines, "\n")
	rightBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorCyanDim).
		Width(rightW).
		Render(rightContent)

	// ── Combine ────────────────────────────────────────────────
	middle := SideBySide(leftStyled, rightBox)

	title := " " + StyleGold.Render("◆ ANCIENT CONTROL CHAIR") + StyleDim.Render(" — ") + StyleCyan.Render("UPGRADE TERMINAL")

	bottomBar := " " +
		FormatKeyHint("↑↓", "select") + "  " +
		FormatKeyHint("⏎", "upgrade") + "  " +
		FormatKeyHint("q", "back")

	content := title + "\n" + middle + "\n" + bottomBar

	return lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorCyan).
		Width(w - 2).
		Height(h - 2).
		Render(content)
}

func selPrefix(selected bool) string {
	if selected {
		return StyleGold.Render(" ▸")
	}
	return "  "
}

func costColor(canAfford bool) lipgloss.Style {
	if canAfford {
		return StyleSuccess
	}
	return StyleDanger
}
