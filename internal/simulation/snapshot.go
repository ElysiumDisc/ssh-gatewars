package simulation

import (
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/gamedata"
)

// PlanetSnapshot is an immutable view of a planet instance for rendering.
type PlanetSnapshot struct {
	Tick        uint64
	Address     gamedata.GateAddress
	AddressCode string
	PlanetName  string
	Biome       string
	Threat      int
	MapWidth    int
	MapHeight   int
	Tiles       []gamedata.TileType // copy of tile data
	GatePos     core.Pos
	Enemies     []EnemySnapshot
	Items       []ItemSnapshot
	Players     []PlayerSnapshot
	Projectiles []ProjectileSnapshot
}

// EnemySnapshot is a read-only view of an enemy.
type EnemySnapshot struct {
	ID    uint32
	DefID string
	HP    int
	MaxHP int
	Pos   core.Pos
	State int // AIState as int for rendering (stunned, alert, etc.)
}

// ItemSnapshot is a read-only view of a ground item.
type ItemSnapshot struct {
	ID    uint32
	DefID string
	Qty   int
	Pos   core.Pos
}

// PlayerSnapshot is a read-only view of a player on a planet.
type PlayerSnapshot struct {
	Key      string
	CallSign string
	HP       int
	MaxHP    int
	Pos      core.Pos
}

// ProjectileSnapshot is a read-only view of an in-flight projectile.
type ProjectileSnapshot struct {
	ID    uint32
	Pos   core.Pos
	Glyph rune
	Color string
}

// CharacterSnapshot is a read-only view of the local player's character.
type CharacterSnapshot struct {
	HP, MaxHP    int
	Level        int
	XP           int
	WeaponName   string
	ArmorName    string
	AttackPower  int
	DefensePower int
	AmmoLoaded   int // current ammo in clip
	AmmoMax      int // clip size
	IsReloading  bool
}
