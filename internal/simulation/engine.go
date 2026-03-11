package simulation

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"

	"ssh-gatewars/internal/chat"
	"ssh-gatewars/internal/combat"
	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/entity"
	"ssh-gatewars/internal/gamedata"
)

// Engine is the single-writer game simulation.
type Engine struct {
	cfg core.GameConfig

	// Active planet instances
	planets   map[string]*PlanetInstance // address code → instance
	planetsMu sync.RWMutex

	// SGC hub (always loaded)
	sgc *PlanetInstance

	// Player state: sshKey → character
	characters   map[string]*entity.Character
	charactersMu sync.RWMutex

	// Events: per-player message queues
	events   map[string][]string
	eventsMu sync.Mutex

	// Action channel
	actions chan PlayerAction

	// Chat game events (engine → chat hub)
	GameEvents chan chat.GameEvent

	tick uint64
}

// NewEngine creates a new game engine.
func NewEngine(cfg core.GameConfig) *Engine {
	sgc := NewSGCInstance(cfg)
	sgc.PublishSnapshot(0)

	e := &Engine{
		cfg:        cfg,
		planets:    make(map[string]*PlanetInstance),
		sgc:        sgc,
		characters: make(map[string]*entity.Character),
		events:     make(map[string][]string),
		actions:    make(chan PlayerAction, 1000),
		GameEvents: make(chan chat.GameEvent, 100),
	}

	// Register SGC by Earth address
	e.planets[gamedata.EarthAddress.Code()] = sgc

	return e
}

// Run starts the engine tick loop. Blocks until context is cancelled.
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second / time.Duration(e.cfg.TickRate))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.doTick()
		}
	}
}

func (e *Engine) doTick() {
	e.tick++
	e.processActions()
	e.tickPlanets()

	// Periodic cleanup
	if e.tick%uint64(e.cfg.TickRate*e.cfg.UnloadTimeout) == 0 {
		e.unloadEmptyPlanets()
	}
}

// EnqueueAction submits a player action. Non-blocking; drops if full.
func (e *Engine) EnqueueAction(a PlayerAction) {
	select {
	case e.actions <- a:
	default:
	}
}

// RegisterCharacter adds or loads a player character.
func (e *Engine) RegisterCharacter(c *entity.Character) {
	e.charactersMu.Lock()
	defer e.charactersMu.Unlock()
	e.characters[c.SSHKey] = c

	// Place on their current planet
	e.planetsMu.RLock()
	defer e.planetsMu.RUnlock()

	if c.Location == "sgc" {
		e.sgc.Players[c.SSHKey] = c.Pos
	} else if pi, ok := e.planets[c.Location]; ok {
		pi.Players[c.SSHKey] = c.Pos
	} else {
		// Planet not loaded, send to SGC
		c.Location = "sgc"
		c.Pos = e.sgc.Map.SpawnPos
		e.sgc.Players[c.SSHKey] = c.Pos
	}
}

// UnregisterCharacter removes a player from the engine.
func (e *Engine) UnregisterCharacter(key string) {
	e.charactersMu.Lock()
	c := e.characters[key]
	delete(e.characters, key)
	e.charactersMu.Unlock()

	if c == nil {
		return
	}

	e.planetsMu.RLock()
	defer e.planetsMu.RUnlock()

	if c.Location == "sgc" {
		e.sgc.RemovePlayer(key)
	} else if pi, ok := e.planets[c.Location]; ok {
		pi.RemovePlayer(key)
	}
}

// GetCharacter returns a copy of the character (for rendering).
func (e *Engine) GetCharacter(key string) *entity.Character {
	e.charactersMu.RLock()
	defer e.charactersMu.RUnlock()
	return e.characters[key]
}

// GetPlanetSnapshot returns the snapshot for a player's current planet.
func (e *Engine) GetPlanetSnapshot(key string) *PlanetSnapshot {
	e.charactersMu.RLock()
	c := e.characters[key]
	e.charactersMu.RUnlock()

	if c == nil {
		return e.sgc.GetSnapshot()
	}

	e.planetsMu.RLock()
	defer e.planetsMu.RUnlock()

	if c.Location == "sgc" {
		return e.sgc.GetSnapshot()
	}
	if pi, ok := e.planets[c.Location]; ok {
		return pi.GetSnapshot()
	}
	return e.sgc.GetSnapshot()
}

// DrainEvents returns and clears pending event messages for a player.
func (e *Engine) DrainEvents(key string) []string {
	e.eventsMu.Lock()
	defer e.eventsMu.Unlock()
	msgs := e.events[key]
	delete(e.events, key)
	return msgs
}

func (e *Engine) emitGameEvent(ev chat.GameEvent) {
	select {
	case e.GameEvents <- ev:
	default:
	}
}

func (e *Engine) pushEvent(key, msg string) {
	e.eventsMu.Lock()
	defer e.eventsMu.Unlock()
	e.events[key] = append(e.events[key], msg)
}

// GetOnlineCount returns the number of registered characters.
func (e *Engine) GetOnlineCount() int {
	e.charactersMu.RLock()
	defer e.charactersMu.RUnlock()
	return len(e.characters)
}

// processActions drains all queued player actions.
func (e *Engine) processActions() {
	for {
		select {
		case a := <-e.actions:
			e.handleAction(a)
		default:
			return
		}
	}
}

func (e *Engine) handleAction(a PlayerAction) {
	e.charactersMu.Lock()
	c := e.characters[a.PlayerKey]
	if c == nil {
		e.charactersMu.Unlock()
		return
	}

	switch a.Type {
	case ActionMove:
		e.handleMove(c, a)
	case ActionInteract:
		e.handleInteract(c, a)
	case ActionDialGate:
		e.handleDial(c, a)
	case ActionPickup:
		e.handlePickup(c, a)
	case ActionUseItem:
		e.handleUseItem(c, a)
	case ActionEquip:
		e.handleEquip(c, a)
	case ActionFire:
		e.handleFire(c, a)
	case ActionReload:
		e.handleReload(c, a)
	}
	e.charactersMu.Unlock()
}

func (e *Engine) handleMove(c *entity.Character, a PlayerAction) {
	if c.IsDead() {
		return
	}

	newPos := c.Pos.Add(a.Dir)
	pi := e.getPlanetForChar(c)
	if pi == nil {
		return
	}

	// Check for enemy at destination (bump-attack)
	if enemy := pi.EnemyAt(newPos); enemy != nil {
		dmg := combat.CalcDamage(c.AttackPower(), enemy.DefensePower())
		killed := enemy.TakeDamage(dmg)
		e.pushEvent(a.PlayerKey, formatAttackMsg(enemy.Def().Name, dmg, killed))

		if killed {
			xp := combat.XPForKill(enemy.Def().XP, pi.Threat)
			leveled := c.GainXP(xp, e.cfg.XPPerLevel)
			e.pushEvent(a.PlayerKey, formatKillMsg(enemy.Def().Name, xp))
			if leveled {
				e.pushEvent(a.PlayerKey, formatLevelMsg(c.Level))
				e.emitGameEvent(chat.GameEvent{
					Type:        chat.GamePlayerLevelUp,
					Fingerprint: a.PlayerKey,
					Callsign:    c.CallSign,
					Extra:       itoa(c.Level),
				})
			}

			// Loot drop
			if pi.rng.Float64() < e.cfg.LootChance {
				lootID := entity.RollLoot(enemy.Def().LootTable, pi.rng)
				if lootID != "" {
					pi.SpawnGroundItem(lootID, 1, enemy.Pos)
					e.pushEvent(a.PlayerKey, formatDropMsg(gamedata.Items[lootID].Name))
				}
			}
		}
		return
	}

	// Check walkability
	if !pi.Map.IsWalkable(newPos) {
		return
	}

	// Check for other players blocking
	if key := pi.PlayerAt(newPos); key != "" {
		return
	}

	// Move
	c.Pos = newPos
	pi.Players[a.PlayerKey] = newPos
}

func (e *Engine) handleInteract(c *entity.Character, a PlayerAction) {
	if c.IsDead() {
		return
	}

	pi := e.getPlanetForChar(c)
	if pi == nil {
		return
	}

	// Check adjacent tiles for interactable objects
	dirs := []core.Pos{core.DirUp, core.DirDown, core.DirLeft, core.DirRight, {X: 0, Y: 0}}
	for _, d := range dirs {
		pos := c.Pos.Add(d)

		// Crate
		if pi.Map.At(pos) == gamedata.TileCrate {
			// Open crate: spawn loot, replace with floor
			table := "crate_common"
			if pi.Threat >= 5 {
				table = "crate_rare"
			}
			lootID := entity.RollLoot(table, pi.rng)
			if lootID != "" {
				if c.AddItem(lootID, 1) {
					e.pushEvent(a.PlayerKey, "Found "+gamedata.Items[lootID].Name+" in crate!")
				} else {
					pi.SpawnGroundItem(lootID, 1, pos)
					e.pushEvent(a.PlayerKey, "Inventory full! Item dropped on floor.")
				}
			} else {
				e.pushEvent(a.PlayerKey, "The crate is empty.")
			}
			pi.Map.Set(pos, gamedata.TileFloor)
			return
		}

		// Ground item
		if item := pi.ItemAt(pos); item != nil {
			if c.AddItem(item.DefID, item.Qty) {
				e.pushEvent(a.PlayerKey, "Picked up "+gamedata.Items[item.DefID].Name)
				delete(pi.Items, item.ID)
			} else {
				e.pushEvent(a.PlayerKey, "Inventory full!")
			}
			return
		}

		// Console
		if pi.Map.At(pos) == gamedata.TileConsole {
			e.pushEvent(a.PlayerKey, "The console hums with alien energy...")
			return
		}
	}
}

func (e *Engine) handleDial(c *entity.Character, a PlayerAction) {
	pi := e.getPlanetForChar(c)
	if pi == nil {
		return
	}

	// Must be adjacent to stargate
	if c.Pos.ManhattanDist(pi.Map.GatePos) > 1 {
		e.pushEvent(a.PlayerKey, "You must be next to the Stargate to dial.")
		return
	}

	if !a.Address.IsValid() {
		e.pushEvent(a.PlayerKey, "Invalid gate address!")
		return
	}

	destCode := a.Address.Code()

	// Emit departure event
	e.emitGameEvent(chat.GameEvent{
		Type:        chat.GamePlayerDeparted,
		Fingerprint: a.PlayerKey,
		Callsign:    c.CallSign,
		PlanetSeed:  c.Location,
		PlanetName:  pi.Name,
	})

	// Emit dial event (chevron sequence)
	e.emitGameEvent(chat.GameEvent{
		Type:        chat.GameGateDial,
		Fingerprint: a.PlayerKey,
		Callsign:    c.CallSign,
		PlanetSeed:  c.Location,
	})

	// Remove from current planet
	pi.RemovePlayer(a.PlayerKey)

	// Load or create destination planet
	e.planetsMu.Lock()
	dest, ok := e.planets[destCode]
	if !ok {
		dest = NewPlanetInstance(a.Address, e.cfg)
		dest.PublishSnapshot(e.tick)
		e.planets[destCode] = dest
		log.Info("planet loaded", "address", destCode, "name", dest.Name)
	}
	e.planetsMu.Unlock()

	// Place player on destination
	spawnPos := dest.AddPlayer(a.PlayerKey)
	c.Location = destCode
	c.Pos = spawnPos

	// Discover address if new
	if c.DiscoverAddress(a.Address) {
		e.pushEvent(a.PlayerKey, "New address discovered: "+gamedata.PlanetName(a.Address))
	}

	e.pushEvent(a.PlayerKey, "Arrived at "+dest.Name)

	// Emit arrival event
	e.emitGameEvent(chat.GameEvent{
		Type:        chat.GamePlayerArrived,
		Fingerprint: a.PlayerKey,
		Callsign:    c.CallSign,
		PlanetSeed:  destCode,
		PlanetName:  dest.Name,
	})
}

func (e *Engine) handlePickup(c *entity.Character, a PlayerAction) {
	pi := e.getPlanetForChar(c)
	if pi == nil {
		return
	}

	// Check for items at player's position or adjacent
	dirs := []core.Pos{{X: 0, Y: 0}, core.DirUp, core.DirDown, core.DirLeft, core.DirRight}
	for _, d := range dirs {
		pos := c.Pos.Add(d)
		if item := pi.ItemAt(pos); item != nil {
			if c.AddItem(item.DefID, item.Qty) {
				e.pushEvent(a.PlayerKey, "Picked up "+gamedata.Items[item.DefID].Name)
				delete(pi.Items, item.ID)
			} else {
				e.pushEvent(a.PlayerKey, "Inventory full!")
			}
			return
		}
	}
	e.pushEvent(a.PlayerKey, "Nothing to pick up here.")
}

func (e *Engine) handleUseItem(c *entity.Character, a PlayerAction) {
	if a.ItemIndex < 0 || a.ItemIndex >= len(c.Inventory) {
		return
	}
	item := &c.Inventory[a.ItemIndex]
	def := gamedata.Items[item.DefID]

	if !def.Consumable {
		e.pushEvent(a.PlayerKey, "Can't use that item.")
		return
	}

	if def.HealAmount > 0 {
		c.Heal(def.HealAmount)
		e.pushEvent(a.PlayerKey, formatHealMsg(def.Name, def.HealAmount))
	}

	c.RemoveItem(item.DefID, 1)
}

func (e *Engine) handleEquip(c *entity.Character, a PlayerAction) {
	if a.ItemIndex < 0 || a.ItemIndex >= len(c.Inventory) {
		return
	}
	item := c.Inventory[a.ItemIndex]
	def := gamedata.Items[item.DefID]

	switch def.Slot {
	case gamedata.SlotWeapon:
		if c.Weapon != nil {
			c.AddItem(c.Weapon.DefID, 1)
		}
		c.Weapon = &entity.Item{DefID: item.DefID, Quantity: 1}
		c.RemoveItem(item.DefID, 1)
		e.pushEvent(a.PlayerKey, "Equipped "+def.Name)
	case gamedata.SlotArmor:
		if c.Armor != nil {
			c.AddItem(c.Armor.DefID, 1)
		}
		c.Armor = &entity.Item{DefID: item.DefID, Quantity: 1}
		c.RemoveItem(item.DefID, 1)
		e.pushEvent(a.PlayerKey, "Equipped "+def.Name)
	case gamedata.SlotAccessory:
		if c.Accessory != nil {
			c.AddItem(c.Accessory.DefID, 1)
		}
		c.Accessory = &entity.Item{DefID: item.DefID, Quantity: 1}
		c.RemoveItem(item.DefID, 1)
		e.pushEvent(a.PlayerKey, "Equipped "+def.Name)
	default:
		e.pushEvent(a.PlayerKey, "Can't equip that item.")
	}
}

func (e *Engine) getPlanetForChar(c *entity.Character) *PlanetInstance {
	e.planetsMu.RLock()
	defer e.planetsMu.RUnlock()

	if c.Location == "sgc" {
		return e.sgc
	}
	return e.planets[c.Location]
}

// tickPlanets runs AI and updates for all active planet instances.
func (e *Engine) tickPlanets() {
	e.planetsMu.RLock()
	defer e.planetsMu.RUnlock()

	for _, pi := range e.planets {
		e.tickPlanet(pi)
		pi.PublishSnapshot(e.tick)
	}
}

func (e *Engine) tickPlanet(pi *PlanetInstance) {
	if pi.IsEmpty() {
		pi.emptyTicks++
		return
	}

	e.charactersMu.RLock()
	defer e.charactersMu.RUnlock()

	// Tick projectiles
	e.tickProjectiles(pi)

	// Enemy AI
	for _, enemy := range pi.Enemies {
		if enemy.IsDead() {
			continue
		}

		// Tick stun
		if enemy.IsStunned() {
			enemy.TickStun()
			continue
		}

		// Decrement shot cooldown
		if enemy.ShotCD > 0 {
			enemy.ShotCD--
		}

		enemy.TickCD--
		if enemy.TickCD > 0 {
			continue
		}
		enemy.TickCD = enemy.Def().Speed

		e.tickEnemyAI(pi, enemy)
	}
}

func (e *Engine) tickEnemyAI(pi *PlanetInstance, enemy *entity.Enemy) {
	def := enemy.Def()
	aggroRange := def.AggroRange

	// Find nearest player
	var nearestKey string
	nearestDist := 9999
	for key, pos := range pi.Players {
		d := enemy.Pos.ManhattanDist(pos)
		if d < nearestDist {
			nearestDist = d
			nearestKey = key
		}
	}

	// State transitions based on detection
	if nearestDist <= aggroRange && nearestKey != "" {
		targetPos := pi.Players[nearestKey]
		enemy.Target = targetPos
		enemy.LastKnown = targetPos

		// Check if can see the player (LOS)
		hasLOS := pi.Map.HasLOS(enemy.Pos, targetPos)

		if hasLOS && def.IsRanged && nearestDist <= def.Range && nearestDist > 1 {
			// In weapon range with LOS — attack from distance
			enemy.State = entity.AIStateAttack
		} else if nearestDist <= aggroRange {
			enemy.State = entity.AIStateChase
		}
	} else if enemy.State == entity.AIStateChase && nearestDist > e.cfg.ChaseRange {
		// Lost target, enter alert briefly then patrol
		enemy.State = entity.AIStateAlert
		enemy.AlertTicks = 5
	}

	// Check flee condition
	if enemy.ShouldFlee() && enemy.State != entity.AIStateFlee {
		enemy.State = entity.AIStateFlee
	}

	switch enemy.State {
	case entity.AIStateIdle:
		// Stationary guard — only aggro if player enters range
		if nearestDist <= aggroRange && nearestKey != "" {
			enemy.State = entity.AIStateChase
		}
	case entity.AIStatePatrol:
		e.enemyPatrol(pi, enemy)
	case entity.AIStateAlert:
		enemy.AlertTicks--
		if enemy.AlertTicks <= 0 {
			enemy.State = entity.AIStatePatrol
		}
	case entity.AIStateChase:
		e.enemyChase(pi, enemy, nearestKey)
	case entity.AIStateAttack:
		e.enemyRangedAttack(pi, enemy, nearestKey)
	case entity.AIStateFlee:
		e.enemyFlee(pi, enemy, nearestKey)
	}
}

func (e *Engine) enemyPatrol(pi *PlanetInstance, enemy *entity.Enemy) {
	// Try current direction, or pick a new one
	newPos := enemy.Pos.Add(enemy.PatrolDir)
	if pi.Map.IsWalkable(newPos) && pi.EnemyAt(newPos) == nil {
		enemy.Pos = newPos
		return
	}

	// Pick random direction
	dirs := []core.Pos{core.DirUp, core.DirDown, core.DirLeft, core.DirRight}
	dir := dirs[pi.rng.Intn(4)]
	newPos = enemy.Pos.Add(dir)
	if pi.Map.IsWalkable(newPos) && pi.EnemyAt(newPos) == nil {
		enemy.PatrolDir = dir
		enemy.Pos = newPos
	}
}

func (e *Engine) enemyChase(pi *PlanetInstance, enemy *entity.Enemy, targetKey string) {
	if targetKey == "" {
		return
	}

	targetPos, ok := pi.Players[targetKey]
	if !ok {
		enemy.State = entity.AIStatePatrol
		return
	}

	// Adjacent = bump-attack the player
	if enemy.Pos.ManhattanDist(targetPos) == 1 {
		c := e.characters[targetKey]
		if c == nil || c.IsDead() {
			return
		}
		dmg := combat.CalcDamage(enemy.AttackPower(), c.DefensePower())
		died := c.TakeDamage(dmg)
		e.pushEvent(targetKey, formatEnemyAttackMsg(enemy.Def().Name, dmg))
		if died {
			e.pushEvent(targetKey, "You have been killed! Respawning at SGC...")
			e.emitGameEvent(chat.GameEvent{
				Type:        chat.GamePlayerKilled,
				Fingerprint: targetKey,
				Callsign:    c.CallSign,
				PlanetName:  pi.Name,
			})
			c.Respawn()
			pi.RemovePlayer(targetKey)
			e.sgc.Players[targetKey] = c.Pos
		}
		return
	}

	// Move toward player (simple greedy pathfinding)
	dx := 0
	dy := 0
	if targetPos.X > enemy.Pos.X {
		dx = 1
	} else if targetPos.X < enemy.Pos.X {
		dx = -1
	}
	if targetPos.Y > enemy.Pos.Y {
		dy = 1
	} else if targetPos.Y < enemy.Pos.Y {
		dy = -1
	}

	// Try primary axis first, then secondary
	candidates := []core.Pos{}
	if dx != 0 {
		candidates = append(candidates, core.Pos{X: dx, Y: 0})
	}
	if dy != 0 {
		candidates = append(candidates, core.Pos{X: 0, Y: dy})
	}

	for _, d := range candidates {
		newPos := enemy.Pos.Add(d)
		if pi.Map.IsWalkable(newPos) && pi.EnemyAt(newPos) == nil && pi.PlayerAt(newPos) == "" {
			enemy.Pos = newPos
			return
		}
	}
}

func (e *Engine) enemyRangedAttack(pi *PlanetInstance, enemy *entity.Enemy, targetKey string) {
	if targetKey == "" || !enemy.CanShoot() {
		// Can't shoot, fall back to chase
		e.enemyChase(pi, enemy, targetKey)
		return
	}

	targetPos, ok := pi.Players[targetKey]
	if !ok {
		enemy.State = entity.AIStatePatrol
		return
	}

	// Check LOS
	if !pi.Map.HasLOS(enemy.Pos, targetPos) {
		// Lost LOS, chase
		enemy.State = entity.AIStateChase
		return
	}

	def := enemy.Def()
	dist := enemy.Pos.ManhattanDist(targetPos)
	if dist > def.Range {
		enemy.State = entity.AIStateChase
		return
	}

	// Fire a projectile
	proj := combat.NewProjectile(
		0, // ID assigned by SpawnProjectile
		itoa(int(enemy.ID)),
		false, // not a player
		enemy.Pos,
		targetPos,
		def.Attack,
		def.ProjGlyph,
		def.ProjColor,
		2, // enemy projectile speed
	)
	pi.SpawnProjectile(proj)
	enemy.ShotCD = def.Speed // cooldown = speed stat
}

func (e *Engine) enemyFlee(pi *PlanetInstance, enemy *entity.Enemy, nearestKey string) {
	if nearestKey == "" {
		enemy.State = entity.AIStatePatrol
		return
	}

	targetPos, ok := pi.Players[nearestKey]
	if !ok {
		enemy.State = entity.AIStatePatrol
		return
	}

	// Move away from player
	dx := 0
	dy := 0
	if targetPos.X > enemy.Pos.X {
		dx = -1
	} else if targetPos.X < enemy.Pos.X {
		dx = 1
	}
	if targetPos.Y > enemy.Pos.Y {
		dy = -1
	} else if targetPos.Y < enemy.Pos.Y {
		dy = 1
	}

	candidates := []core.Pos{}
	if dx != 0 {
		candidates = append(candidates, core.Pos{X: dx, Y: 0})
	}
	if dy != 0 {
		candidates = append(candidates, core.Pos{X: 0, Y: dy})
	}

	for _, d := range candidates {
		newPos := enemy.Pos.Add(d)
		if pi.Map.IsWalkable(newPos) && pi.EnemyAt(newPos) == nil && pi.PlayerAt(newPos) == "" {
			enemy.Pos = newPos
			return
		}
	}
}

func (e *Engine) tickProjectiles(pi *PlanetInstance) {
	for id, proj := range pi.Projectiles {
		expired := proj.Advance()
		if expired {
			delete(pi.Projectiles, id)
			continue
		}

		// Check collision with walls
		if pi.Map.IsOpaque(proj.Pos) {
			delete(pi.Projectiles, id)
			continue
		}

		if proj.IsPlayer {
			// Player projectile — check hit on enemy
			if enemy := pi.EnemyAt(proj.Pos); enemy != nil {
				killed := enemy.TakeDamage(proj.Damage)
				e.pushEvent(proj.OwnerKey, formatAttackMsg(enemy.Def().Name, proj.Damage, killed))

				if killed {
					c := e.characters[proj.OwnerKey]
					if c != nil {
						xp := combat.XPForKill(enemy.Def().XP, pi.Threat)
						leveled := c.GainXP(xp, e.cfg.XPPerLevel)
						e.pushEvent(proj.OwnerKey, formatKillMsg(enemy.Def().Name, xp))
						if leveled {
							e.pushEvent(proj.OwnerKey, formatLevelMsg(c.Level))
							e.emitGameEvent(chat.GameEvent{
								Type:        chat.GamePlayerLevelUp,
								Fingerprint: proj.OwnerKey,
								Callsign:    c.CallSign,
								Extra:       itoa(c.Level),
							})
						}
						// Loot drop
						if pi.rng.Float64() < e.cfg.LootChance {
							lootID := entity.RollLoot(enemy.Def().LootTable, pi.rng)
							if lootID != "" {
								pi.SpawnGroundItem(lootID, 1, enemy.Pos)
								e.pushEvent(proj.OwnerKey, formatDropMsg(gamedata.Items[lootID].Name))
							}
						}
					}
				}
				delete(pi.Projectiles, id)
				continue
			}
		} else {
			// Enemy projectile — check hit on player
			if key := pi.PlayerAt(proj.Pos); key != "" {
				c := e.characters[key]
				if c != nil && !c.IsDead() {
					dmg := combat.CalcDamage(proj.Damage, c.DefensePower())
					died := c.TakeDamage(dmg)
					e.pushEvent(key, "Incoming fire! " + itoa(dmg) + " damage!")
					if died {
						e.pushEvent(key, "You have been killed! Respawning at SGC...")
						e.emitGameEvent(chat.GameEvent{
							Type:        chat.GamePlayerKilled,
							Fingerprint: key,
							Callsign:    c.CallSign,
							PlanetName:  pi.Name,
						})
						c.Respawn()
						pi.RemovePlayer(key)
						e.sgc.Players[key] = c.Pos
					}
				}
				delete(pi.Projectiles, id)
				continue
			}
		}
	}
}

func (e *Engine) handleFire(c *entity.Character, a PlayerAction) {
	if c.IsDead() || c.Weapon == nil {
		return
	}

	pi := e.getPlanetForChar(c)
	if pi == nil {
		return
	}

	wDef := gamedata.Items[c.Weapon.DefID]
	if wDef.WType != gamedata.WeaponRanged {
		e.pushEvent(a.PlayerKey, "You can't fire that weapon!")
		return
	}

	// Check range
	dist := c.Pos.ManhattanDist(a.FireTarget)
	if dist > wDef.Range {
		e.pushEvent(a.PlayerKey, "Target out of range!")
		return
	}

	// Check LOS
	if !pi.Map.HasLOS(c.Pos, a.FireTarget) {
		e.pushEvent(a.PlayerKey, "No line of sight!")
		return
	}

	// Create projectile
	speed := wDef.ProjSpeed
	if speed == 0 {
		speed = 2
	}
	glyph := wDef.Glyph
	if glyph == 0 {
		glyph = '-'
	}
	color := wDef.ProjColor
	if color == "" {
		color = "#FFCC44"
	}

	proj := combat.NewProjectile(
		0,
		a.PlayerKey,
		true,
		c.Pos,
		a.FireTarget,
		c.AttackPower(),
		glyph,
		color,
		speed,
	)
	pi.SpawnProjectile(proj)
}

func (e *Engine) handleReload(c *entity.Character, a PlayerAction) {
	// Reload is a no-op for now (ammo system is tracked but not consumed yet in MVP)
	if c.Weapon == nil {
		return
	}
	e.pushEvent(a.PlayerKey, "Weapon reloaded.")
}

func (e *Engine) unloadEmptyPlanets() {
	e.planetsMu.Lock()
	defer e.planetsMu.Unlock()

	for code, pi := range e.planets {
		if pi.Persistent || !pi.IsEmpty() {
			continue
		}
		if pi.emptyTicks > e.cfg.TickRate*e.cfg.UnloadTimeout {
			delete(e.planets, code)
			log.Info("planet unloaded", "address", code, "name", pi.Name)
		}
	}
}

// Format helpers for event messages.
func formatAttackMsg(name string, dmg int, killed bool) string {
	if killed {
		return "You strike " + name + " for " + itoa(dmg) + " damage — killing blow!"
	}
	return "You strike " + name + " for " + itoa(dmg) + " damage."
}

func formatKillMsg(name string, xp int) string {
	return name + " defeated! +" + itoa(xp) + " XP"
}

func formatLevelMsg(level int) string {
	return "LEVEL UP! You are now level " + itoa(level) + "!"
}

func formatDropMsg(itemName string) string {
	return itemName + " dropped on the ground."
}

func formatEnemyAttackMsg(name string, dmg int) string {
	return name + " hits you for " + itoa(dmg) + " damage!"
}

func formatHealMsg(itemName string, amount int) string {
	return "Used " + itemName + " — restored " + itoa(amount) + " HP."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
