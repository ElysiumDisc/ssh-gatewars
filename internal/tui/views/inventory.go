package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ssh-gatewars/internal/entity"
	"ssh-gatewars/internal/gamedata"
)

// RenderInventory renders the inventory modal.
func RenderInventory(c *entity.Character, cursor int, width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA44")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333366")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	statStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#44FF44"))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  INVENTORY"))
	b.WriteString("\n\n")

	// Equipment section
	b.WriteString(dimStyle.Render("  Equipment:"))
	b.WriteString("\n")
	writeEquipSlot(&b, "  Weapon:    ", c.Weapon, dimStyle)
	writeEquipSlot(&b, "  Armor:     ", c.Armor, dimStyle)
	writeEquipSlot(&b, "  Accessory: ", c.Accessory, dimStyle)
	b.WriteString("\n")

	// Stats
	b.WriteString(dimStyle.Render("  Stats: "))
	b.WriteString(statStyle.Render(fmt.Sprintf("ATK %d  DEF %d  HP %d/%d",
		c.AttackPower(), c.DefensePower(), c.HP, c.MaxHP)))
	b.WriteString("\n\n")

	// Inventory list
	b.WriteString(dimStyle.Render(fmt.Sprintf("  Items (%d/%d):", len(c.Inventory), c.MaxItems)))
	b.WriteString("\n")

	if len(c.Inventory) == 0 {
		b.WriteString(dimStyle.Render("  (empty)"))
		b.WriteString("\n")
	}

	for i, item := range c.Inventory {
		def := gamedata.Items[item.DefID]
		line := fmt.Sprintf("  %s", def.Name)
		if item.Quantity > 1 {
			line += fmt.Sprintf(" x%d", item.Quantity)
		}

		// Show stats
		if def.Attack > 0 {
			line += fmt.Sprintf(" [ATK +%d]", def.Attack)
		}
		if def.Defense > 0 {
			line += fmt.Sprintf(" [DEF +%d]", def.Defense)
		}
		if def.HealAmount > 0 {
			line += fmt.Sprintf(" [Heal %d]", def.HealAmount)
		}

		if i == cursor {
			b.WriteString(selectedStyle.Render("> " + line))
		} else {
			b.WriteString(normalStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Show description of selected item
	if cursor >= 0 && cursor < len(c.Inventory) {
		def := gamedata.Items[c.Inventory[cursor].DefID]
		b.WriteString(dimStyle.Render("  " + def.Description))
		b.WriteString("\n\n")
	}

	b.WriteString(dimStyle.Render("  ↑↓ select | E equip | u use | D drop | Esc close"))
	b.WriteString("\n")

	return b.String()
}

func writeEquipSlot(b *strings.Builder, label string, item *entity.Item, dim lipgloss.Style) {
	b.WriteString(dim.Render(label))
	if item != nil {
		def := gamedata.Items[item.DefID]
		b.WriteString(def.Name)
	} else {
		b.WriteString(dim.Render("(none)"))
	}
	b.WriteString("\n")
}

// RenderAddressBook renders the known gate addresses.
func RenderAddressBook(addresses []gamedata.GateAddress, cursor int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#44AAFF")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333366")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("  ◎ ADDRESS BOOK"))
	b.WriteString("\n\n")

	if len(addresses) == 0 {
		b.WriteString(dimStyle.Render("  No addresses discovered."))
		b.WriteString("\n")
		return b.String()
	}

	for i, addr := range addresses {
		name := gamedata.PlanetName(addr)
		line := fmt.Sprintf("  %s  %s  (%s)", addr.String(), name, addr.Code())

		if i == cursor {
			b.WriteString(selectedStyle.Render("> " + line))
		} else {
			b.WriteString(normalStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  ↑↓ select | Enter dial | Esc close"))
	b.WriteString("\n")

	return b.String()
}

// RenderHelp renders the help overlay.
func RenderHelp() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA44")).
		Bold(true)
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA"))

	lines := []string{
		"",
		titleStyle.Render("  GATEWARS — HELP"),
		"",
		helpStyle.Render("  Movement:     WASD / HJKL / Arrow keys"),
		helpStyle.Render("  Attack:       Walk into an enemy (bump)"),
		helpStyle.Render("  Interact:     E (loot crates, pick up items)"),
		helpStyle.Render("  Dial gate:    G (when adjacent to ◎)"),
		helpStyle.Render("  Inventory:    I"),
		helpStyle.Render("  Address book: Shift+A"),
		helpStyle.Render("  Help:         ?"),
		helpStyle.Render("  Quit:         Q / Esc"),
		"",
		helpStyle.Render("  In inventory:"),
		helpStyle.Render("    E = equip, U = use, Shift+D = drop"),
		"",
		helpStyle.Render("  Press any key to close this help."),
	}
	return strings.Join(lines, "\n")
}
