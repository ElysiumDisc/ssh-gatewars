package entity

import (
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// Character represents a player character.
type Character struct {
	ID          int64
	SSHKey      string
	DisplayName string
	CallSign    string

	HP, MaxHP int
	Level     int
	XP        int

	// Current position
	Location string   // "sgc" or gate address code
	Pos      core.Pos

	// Equipment (nil = empty slot)
	Weapon    *Item
	Armor     *Item
	Accessory *Item

	// Inventory
	Inventory []Item
	MaxItems  int

	// Progression
	DiscoveredAddresses []gamedata.GateAddress
	MissionsCompleted   int
	Deaths              int
}

// Item is an instance of an item definition.
type Item struct {
	DefID    string // references gamedata.Items key
	Quantity int
}

// Def returns the item definition for this item.
func (it *Item) Def() gamedata.ItemDef {
	return gamedata.Items[it.DefID]
}

// NewCharacter creates a fresh character with starter gear.
func NewCharacter(sshKey, displayName, callSign string, cfg core.GameConfig) *Character {
	c := &Character{
		SSHKey:      sshKey,
		DisplayName: displayName,
		CallSign:    callSign,
		HP:          cfg.StartHP,
		MaxHP:       cfg.StartMaxHP,
		Level:       1,
		XP:          0,
		Location:    "sgc",
		Pos:         core.Pos{X: 7, Y: 4}, // SGC spawn
		MaxItems:    20,
		Inventory:   make([]Item, 0),
		DiscoveredAddresses: []gamedata.GateAddress{
			gamedata.EarthAddress,
			gamedata.NamedAddresses[1].Address, // Abydos
			gamedata.NamedAddresses[2].Address, // Chulak
		},
	}

	// Starter equipment
	c.Weapon = &Item{DefID: "p90", Quantity: 1}
	c.Armor = &Item{DefID: "tac_vest", Quantity: 1}
	c.Inventory = append(c.Inventory, Item{DefID: "bandage", Quantity: 3})
	c.Inventory = append(c.Inventory, Item{DefID: "gdo", Quantity: 1})

	return c
}

// AttackPower returns total attack (weapon + base).
func (c *Character) AttackPower() int {
	base := 1
	if c.Weapon != nil {
		base += c.Weapon.Def().Attack
	}
	return base
}

// DefensePower returns total defense (armor + base).
func (c *Character) DefensePower() int {
	base := 0
	if c.Armor != nil {
		base += c.Armor.Def().Defense
	}
	return base
}

// TakeDamage reduces HP and returns true if the character dies.
func (c *Character) TakeDamage(dmg int) bool {
	c.HP -= dmg
	if c.HP <= 0 {
		c.HP = 0
		return true
	}
	return false
}

// Heal restores HP up to MaxHP.
func (c *Character) Heal(amount int) {
	c.HP += amount
	if c.HP > c.MaxHP {
		c.HP = c.MaxHP
	}
}

// GainXP adds XP and returns true if leveled up.
func (c *Character) GainXP(xp int, xpPerLevel int) bool {
	c.XP += xp
	threshold := c.Level * xpPerLevel
	if c.XP >= threshold {
		c.XP -= threshold
		c.Level++
		c.MaxHP += 5
		c.HP = c.MaxHP // full heal on level up
		return true
	}
	return false
}

// AddItem adds an item to inventory. Returns false if full.
func (c *Character) AddItem(defID string, qty int) bool {
	// Stack with existing
	for i := range c.Inventory {
		if c.Inventory[i].DefID == defID {
			c.Inventory[i].Quantity += qty
			return true
		}
	}
	if len(c.Inventory) >= c.MaxItems {
		return false
	}
	c.Inventory = append(c.Inventory, Item{DefID: defID, Quantity: qty})
	return true
}

// RemoveItem removes qty of an item. Returns false if not enough.
func (c *Character) RemoveItem(defID string, qty int) bool {
	for i := range c.Inventory {
		if c.Inventory[i].DefID == defID {
			if c.Inventory[i].Quantity < qty {
				return false
			}
			c.Inventory[i].Quantity -= qty
			if c.Inventory[i].Quantity <= 0 {
				c.Inventory = append(c.Inventory[:i], c.Inventory[i+1:]...)
			}
			return true
		}
	}
	return false
}

// HasAddress checks if the player knows a gate address.
func (c *Character) HasAddress(addr gamedata.GateAddress) bool {
	for _, a := range c.DiscoveredAddresses {
		if a == addr {
			return true
		}
	}
	return false
}

// DiscoverAddress adds a new gate address if not already known.
func (c *Character) DiscoverAddress(addr gamedata.GateAddress) bool {
	if c.HasAddress(addr) {
		return false
	}
	c.DiscoveredAddresses = append(c.DiscoveredAddresses, addr)
	return true
}

// IsDead returns true if HP is 0.
func (c *Character) IsDead() bool {
	return c.HP <= 0
}

// Respawn resets character to SGC with reduced HP.
func (c *Character) Respawn() {
	c.Deaths++
	c.Location = "sgc"
	c.Pos = core.Pos{X: 7, Y: 4}
	c.HP = c.MaxHP / 2
	if c.HP < 1 {
		c.HP = 1
	}
}
