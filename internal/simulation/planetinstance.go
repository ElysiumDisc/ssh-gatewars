package simulation

import (
	"math/rand"
	"sync/atomic"

	"ssh-gatewars/internal/combat"
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/entity"
	"ssh-gatewars/internal/gamedata"
	"ssh-gatewars/internal/world"
)

// PlanetInstance is a live, active planet with players on it.
type PlanetInstance struct {
	Address    gamedata.GateAddress
	Name       string
	Biome      gamedata.Biome
	Threat     int
	Map        *world.TileMap
	Persistent bool // SGC and named planets don't unload

	Enemies     map[entity.EntityID]*entity.Enemy
	Items       map[entity.EntityID]*entity.GroundItem
	Players     map[string]core.Pos // sshKey → position
	Projectiles map[uint32]*combat.Projectile

	nextID     entity.EntityID
	nextProjID uint32
	rng        *rand.Rand
	snap       atomic.Pointer[PlanetSnapshot]
	emptyTicks int // ticks with no players
}

// NewPlanetInstance creates a new planet from a gate address.
func NewPlanetInstance(addr gamedata.GateAddress, cfg core.GameConfig) *PlanetInstance {
	biome, threat := world.BiomeForAddress(addr)
	seed := addr.Seed()
	tileMap := world.GeneratePlanet(seed, biome, cfg.MapWidth, cfg.MapHeight)

	pi := &PlanetInstance{
		Address:     addr,
		Name:        gamedata.PlanetName(addr),
		Biome:       biome,
		Threat:      threat,
		Map:         tileMap,
		Enemies:     make(map[entity.EntityID]*entity.Enemy),
		Items:       make(map[entity.EntityID]*entity.GroundItem),
		Players:     make(map[string]core.Pos),
		Projectiles: make(map[uint32]*combat.Projectile),
		nextID:      1,
		rng:         rand.New(rand.NewSource(seed + 12345)),
	}

	// Spawn enemies
	spawns := world.EnemiesForPlanet(tileMap, biome, threat, seed)
	for _, s := range spawns {
		id := pi.nextID
		pi.nextID++
		pi.Enemies[id] = entity.NewEnemy(id, s.EnemyID, s.Pos)
	}

	return pi
}

// NewSGCInstance creates the SGC home base planet.
func NewSGCInstance(cfg core.GameConfig) *PlanetInstance {
	tileMap := world.GenerateSGC(cfg.SGCWidth, cfg.SGCHeight)
	return &PlanetInstance{
		Address:     gamedata.EarthAddress,
		Name:        "SGC",
		Biome:       gamedata.Biome{ID: "sgc", Name: "Stargate Command"},
		Threat:      0,
		Map:         tileMap,
		Persistent:  true,
		Enemies:     make(map[entity.EntityID]*entity.Enemy),
		Items:       make(map[entity.EntityID]*entity.GroundItem),
		Players:     make(map[string]core.Pos),
		Projectiles: make(map[uint32]*combat.Projectile),
		nextID:      1,
		rng:         rand.New(rand.NewSource(0)),
	}
}

// AddPlayer registers a player on this planet at the spawn position.
func (pi *PlanetInstance) AddPlayer(key string) core.Pos {
	pos := pi.Map.SpawnPos
	pi.Players[key] = pos
	pi.emptyTicks = 0
	return pos
}

// RemovePlayer unregisters a player from this planet.
func (pi *PlanetInstance) RemovePlayer(key string) {
	delete(pi.Players, key)
}

// IsEmpty returns true if no players are on this planet.
func (pi *PlanetInstance) IsEmpty() bool {
	return len(pi.Players) == 0
}

// EnemyAt returns the enemy at a position, or nil.
func (pi *PlanetInstance) EnemyAt(pos core.Pos) *entity.Enemy {
	for _, e := range pi.Enemies {
		if e.Pos == pos && !e.IsDead() {
			return e
		}
	}
	return nil
}

// ItemAt returns the ground item at a position, or nil.
func (pi *PlanetInstance) ItemAt(pos core.Pos) *entity.GroundItem {
	for _, it := range pi.Items {
		if it.Pos == pos {
			return it
		}
	}
	return nil
}

// PlayerAt returns the SSH key of a player at a position, or "".
func (pi *PlanetInstance) PlayerAt(pos core.Pos) string {
	for key, p := range pi.Players {
		if p == pos {
			return key
		}
	}
	return ""
}

// SpawnGroundItem adds a loot item to the floor.
func (pi *PlanetInstance) SpawnGroundItem(defID string, qty int, pos core.Pos) {
	id := pi.nextID
	pi.nextID++
	pi.Items[id] = &entity.GroundItem{
		ID:    id,
		DefID: defID,
		Qty:   qty,
		Pos:   pos,
	}
}

// SpawnProjectile adds a projectile to the planet.
func (pi *PlanetInstance) SpawnProjectile(proj *combat.Projectile) {
	pi.nextProjID++
	proj.ID = pi.nextProjID
	pi.Projectiles[proj.ID] = proj
}

// PublishSnapshot creates an immutable snapshot of the planet state.
func (pi *PlanetInstance) PublishSnapshot(tick uint64) {
	snap := &PlanetSnapshot{
		Tick:        tick,
		Address:     pi.Address,
		AddressCode: pi.Address.Code(),
		PlanetName:  pi.Name,
		Biome:       pi.Biome.ID,
		Threat:      pi.Threat,
		MapWidth:    pi.Map.Width,
		MapHeight:   pi.Map.Height,
		Tiles:       make([]gamedata.TileType, len(pi.Map.Tiles)),
		GatePos:     pi.Map.GatePos,
	}
	copy(snap.Tiles, pi.Map.Tiles)

	for _, e := range pi.Enemies {
		if !e.IsDead() {
			snap.Enemies = append(snap.Enemies, EnemySnapshot{
				ID:    uint32(e.ID),
				DefID: e.DefID,
				HP:    e.HP,
				MaxHP: e.MaxHP,
				Pos:   e.Pos,
				State: int(e.State),
			})
		}
	}

	for _, it := range pi.Items {
		snap.Items = append(snap.Items, ItemSnapshot{
			ID:    uint32(it.ID),
			DefID: it.DefID,
			Qty:   it.Qty,
			Pos:   it.Pos,
		})
	}

	for _, p := range pi.Projectiles {
		snap.Projectiles = append(snap.Projectiles, ProjectileSnapshot{
			ID:    p.ID,
			Pos:   p.Pos,
			Glyph: p.Glyph,
			Color: p.Color,
		})
	}

	pi.snap.Store(snap)
}

// GetSnapshot returns the latest immutable snapshot.
func (pi *PlanetInstance) GetSnapshot() *PlanetSnapshot {
	return pi.snap.Load()
}
