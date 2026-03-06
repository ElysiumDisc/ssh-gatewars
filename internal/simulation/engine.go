package simulation

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"ssh-gatewars/internal/faction"
)

const (
	WorldW   = 400
	WorldH   = 200
	TickRate = 10 // ticks per second

	// Spatial hash
	CellSize = 20
	GridW    = WorldW / CellSize  // 20
	GridH    = WorldH / CellSize  // 10

	// Combat
	AttackRange = 5.0

	// Explosions
	ExplodeFrames    = 4
	ExplodeTicksPerFrame = 4 // ~400ms at 10 ticks/sec

	MaxShips = 2000
)

// Gate positions in a pentagon arrangement.
var GatePositions [faction.Count]Vec2

func init() {
	cx, cy := float64(WorldW)/2, float64(WorldH)/2
	radius := float64(WorldH) * 0.4 // 80 units

	// Pentagon: Asgard top, Goa'uld upper-left, Jaffa upper-right,
	// Lucian lower-left, Tau'ri lower-right
	// Angles from top (90°), going clockwise
	angles := [faction.Count]float64{
		4 * math.Pi / 5,  // Tau'ri: lower-right (324° from top = ~-36° = 4π/5 from right)
		math.Pi,          // Goa'uld: left (180°)
		0,                // Jaffa: right (0°)
		3 * math.Pi / 5,  // Lucian: lower-left
		math.Pi / 2,      // Asgard: top (90°)
	}

	for i := 0; i < faction.Count; i++ {
		a := angles[i]
		GatePositions[i] = Vec2{
			X: cx + radius*math.Cos(a),
			Y: cy - radius*math.Sin(a), // screen Y is inverted
		}
	}
}

// Explosion represents a ship death animation.
type Explosion struct {
	X, Y      float64
	Frame     int
	TicksLeft int
	Faction   int
}

// Notification is a broadcast message to players.
type Notification struct {
	Faction   int
	Message   string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// BeamEffect represents a visual beam across the battlefield.
type BeamEffect struct {
	X1, Y1    float64 // start point
	X2, Y2    float64 // end point
	Faction   int
	TicksLeft int
}

// Snapshot is a read-only copy of simulation state for rendering.
type Snapshot struct {
	Ships         []Ship
	Explosions    []Explosion
	Beams         []BeamEffect
	Notifications []Notification
	Territory     *TerritoryMap
	PowerStatuses [faction.Count]PowerStatus
	PlayerCounts  [faction.Count]int
	ShipCounts    [faction.Count]int
	KillCounts    [faction.Count]int
	DeathCounts   [faction.Count]int
	FocusVotes    [faction.Count][faction.Count]int // [voter_faction][target_sector]
	Tick          uint64
}

// Engine runs the core simulation loop.
type Engine struct {
	mu sync.RWMutex

	ships      []*Ship
	explosions []Explosion
	beams      []BeamEffect
	nextID     uint64
	tick       uint64

	territory *TerritoryMap
	spatial   SpatialHash

	Powers       *PowerManager
	playerCounts [faction.Count]int
	killCounts   [faction.Count]int
	deathCounts  [faction.Count]int

	notifications []Notification

	// Unique player tracking (SSH key -> faction)
	players map[string]int

	// Sector focus votes per player (SSH key -> sector 0-4, -1 = none)
	focusVotes map[string]int

	rng *rand.Rand
}

// NewEngine creates a simulation engine.
func NewEngine() *Engine {
	e := &Engine{
		territory:  NewTerritoryMap(),
		players:    make(map[string]int),
		focusVotes: make(map[string]int),
		Powers:     NewPowerManager(),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	// Seed initial ships so the war is already happening
	e.seedInitialShips()
	return e
}

func (e *Engine) seedInitialShips() {
	for f := 0; f < faction.Count; f++ {
		count := 8
		if f == faction.Asgard {
			count = 3
		}
		for i := 0; i < count; i++ {
			e.spawnShip(f)
		}
	}
}

func (e *Engine) spawnShip(factionID int) {
	if len(e.ships) >= MaxShips {
		return
	}
	f := faction.Factions[factionID]
	gate := GatePositions[factionID]

	// Determine heading: bias toward focused sector if votes exist
	target := Vec2{WorldW / 2, WorldH / 2}
	votes := e.calculateFocusVotes()
	bestSector := -1
	bestVotes := 0
	for s := 0; s < faction.Count; s++ {
		if votes[factionID][s] > bestVotes {
			bestVotes = votes[factionID][s]
			bestSector = s
		}
	}
	if bestSector >= 0 {
		// Blend: 60% toward focused sector gate, 40% toward center
		sectorTarget := GatePositions[bestSector]
		target = Vec2{
			X: sectorTarget.X*0.6 + target.X*0.4,
			Y: sectorTarget.Y*0.6 + target.Y*0.4,
		}
	}
	dir := target.Sub(gate).Normalize()

	// Add +/-30 degree random spread
	angle := (e.rng.Float64() - 0.5) * math.Pi / 3
	cos, sin := math.Cos(angle), math.Sin(angle)
	dir = Vec2{
		X: dir.X*cos - dir.Y*sin,
		Y: dir.X*sin + dir.Y*cos,
	}

	// Small random offset from gate position
	offset := Vec2{
		X: (e.rng.Float64() - 0.5) * 10,
		Y: (e.rng.Float64() - 0.5) * 5,
	}

	e.nextID++
	ship := &Ship{
		ID:        e.nextID,
		Faction:   factionID,
		X:         gate.X + offset.X,
		Y:         gate.Y + offset.Y,
		VX:        dir.X * float64(f.BaseSpeed),
		VY:        dir.Y * float64(f.BaseSpeed),
		HP:        f.BaseHP,
		MaxHP:     f.BaseHP,
		Damage:    f.BaseDamage,
		Speed:     f.BaseSpeed,
		State:     Alive,
		SpawnTick: e.tick,
	}
	e.ships = append(e.ships, ship)
}

// Run starts the simulation tick loop. Call in a goroutine.
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second / TickRate)
	defer ticker.Stop()

	spawnTimers := [faction.Count]float64{}
	territoryTimer := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()
			dt := 1.0 / float64(TickRate)
			e.tick++

			// 1. Spawn
			for f := 0; f < faction.Count; f++ {
				spawnInterval := 3.0 / (1.0 + float64(e.playerCounts[f]))
				if spawnInterval < 0.2 {
					spawnInterval = 0.2
				}
				spawnInterval /= float64(faction.Factions[f].SpawnMult)

				spawnTimers[f] += dt
				if spawnTimers[f] >= spawnInterval {
					spawnTimers[f] = 0

					// Lucian: 25% double spawn
					count := 1
					if f == faction.Lucian && e.rng.Float64() < 0.25 {
						count = 2
					}
					for i := 0; i < count; i++ {
						e.spawnShip(f)
					}
				}
			}

			// 2. Move
			e.moveShips(dt)

			// 3. Rebuild spatial hash
			e.spatial.Clear()
			for _, s := range e.ships {
				if s.State == Alive {
					e.spatial.Insert(s)
				}
			}

			// 4. Apply active power effects to movement
			e.applyPowerMovementEffects()

			// 5. Combat (with power modifiers)
			e.resolveCombat(dt)

			// 6. Fire Asgard Ion Cannon if just activated
			e.processIonCannon()

			// 7. Process explosions + beam effects
			e.processExplosions()
			e.processBeams()

			// 8. Remove dead ships
			e.removeDeadShips()

			// 9. Territory (every 10 ticks = 1 second)
			territoryTimer++
			if territoryTimer >= TickRate {
				territoryTimer = 0
				e.territory = CalculateTerritory(e.ships)
			}

			// 10. Expire notifications
			now := time.Now()
			alive := e.notifications[:0]
			for _, n := range e.notifications {
				if now.Before(n.ExpiresAt) {
					alive = append(alive, n)
				}
			}
			e.notifications = alive

			e.mu.Unlock()
		}
	}
}

func (e *Engine) moveShips(dt float64) {
	for _, s := range e.ships {
		if s.State != Alive {
			continue
		}

		// Find nearest enemy for steering
		var nearest *Ship
		nearestDist := math.MaxFloat64
		for _, other := range e.ships {
			if other.State != Alive || other.Faction == s.Faction {
				continue
			}
			d := Dist(s, other)
			if d < nearestDist {
				nearestDist = d
				nearest = other
			}
		}

		if nearest != nil {
			desired := Vec2{nearest.X - s.X, nearest.Y - s.Y}.Normalize().Scale(float64(s.Speed))

			// Faction-specific steering
			steerRate := 0.1
			switch s.Faction {
			case faction.Goauld:
				steerRate = 0.05 // slow turning wall
			case faction.Asgard:
				steerRate = 0.15 // precise steering
			case faction.Jaffa:
				// Speed increases near enemies
				speedBonus := math.Max(0, (20-nearestDist)*0.02)
				s.Speed = faction.Factions[faction.Jaffa].BaseSpeed * float32(1.0+speedBonus)
				desired = desired.Normalize().Scale(float64(s.Speed))
			}

			// Lerp velocity toward desired
			s.VX += (desired.X - s.VX) * steerRate
			s.VY += (desired.Y - s.VY) * steerRate
		}

		// Lucian: random wobble
		if s.Faction == faction.Lucian {
			s.VX += (e.rng.Float64() - 0.5) * 0.6
			s.VY += (e.rng.Float64() - 0.5) * 0.3
		}

		// Tau'ri: formation tendency — steer toward nearby allies
		if s.Faction == faction.Tauri {
			allies := e.spatial.QueryRadius(s.X, s.Y, 15)
			cx, cy := 0.0, 0.0
			count := 0
			for _, a := range allies {
				if a.Faction == faction.Tauri && a.ID != s.ID {
					cx += a.X
					cy += a.Y
					count++
				}
			}
			if count > 0 {
				cx /= float64(count)
				cy /= float64(count)
				// Gentle pull toward centroid
				s.VX += (cx - s.X) * 0.005
				s.VY += (cy - s.Y) * 0.005
			}
		}

		// Asgard: wide spacing — steer away from nearby allies
		if s.Faction == faction.Asgard {
			allies := e.spatial.QueryRadius(s.X, s.Y, 12)
			for _, a := range allies {
				if a.Faction == faction.Asgard && a.ID != s.ID {
					d := Dist(s, a)
					if d < 8 && d > 0.1 {
						push := Vec2{s.X - a.X, s.Y - a.Y}.Normalize().Scale(0.3)
						s.VX += push.X
						s.VY += push.Y
					}
				}
			}
		}

		// Record trail (shift old positions, store current)
		if s.TrailLen < 3 {
			s.TrailLen++
		}
		s.Trail[2] = s.Trail[1]
		s.Trail[1] = s.Trail[0]
		s.Trail[0] = Vec2{s.X, s.Y}

		// Update position
		s.X += s.VX * dt
		s.Y += s.VY * dt

		// Soft bounce off world boundaries
		if s.X < 5 {
			s.X = 5
			s.VX = math.Abs(s.VX) * 0.5
		} else if s.X > WorldW-5 {
			s.X = WorldW - 5
			s.VX = -math.Abs(s.VX) * 0.5
		}
		if s.Y < 3 {
			s.Y = 3
			s.VY = math.Abs(s.VY) * 0.5
		} else if s.Y > WorldH-3 {
			s.Y = WorldH - 3
			s.VY = -math.Abs(s.VY) * 0.5
		}
	}
}

func (e *Engine) resolveCombat(dt float64) {
	// Pre-calculate Tau'ri coordinated strike target
	var tauriTarget *Ship
	if e.Powers.IsActive(faction.Tauri) {
		// Find Tau'ri centroid
		var cx, cy float64
		var tc int
		for _, s := range e.ships {
			if s.Faction == faction.Tauri && s.State == Alive {
				cx += s.X
				cy += s.Y
				tc++
			}
		}
		if tc > 0 {
			cx /= float64(tc)
			cy /= float64(tc)
			// Find nearest enemy to centroid
			minD := math.MaxFloat64
			for _, s := range e.ships {
				if s.Faction == faction.Tauri || s.State != Alive {
					continue
				}
				dx := s.X - cx
				dy := s.Y - cy
				d := dx*dx + dy*dy
				if d < minD {
					minD = d
					tauriTarget = s
				}
			}
		}
	}

	for _, s := range e.ships {
		if s.State != Alive {
			continue
		}

		var target *Ship

		// Tau'ri Coordinated Strike: all ships lock same target if in range
		if s.Faction == faction.Tauri && tauriTarget != nil {
			if Dist(s, tauriTarget) <= AttackRange*2 {
				target = tauriTarget
			}
		}

		if target == nil {
			enemies := e.spatial.QueryRadius(s.X, s.Y, AttackRange)
			minDist := math.MaxFloat64
			for _, other := range enemies {
				if other.Faction == s.Faction || other.State != Alive {
					continue
				}
				d := Dist(s, other)
				if d < minDist {
					minDist = d
					target = other
				}
			}
		}

		if target == nil {
			continue
		}

		rawDamage := float64(s.Damage) * dt

		// Tau'ri formation bonus: +6% per nearby ally, max +30%
		if s.Faction == faction.Tauri {
			allies := e.spatial.QueryRadius(s.X, s.Y, 3)
			allyCount := 0
			for _, a := range allies {
				if a.Faction == faction.Tauri && a.ID != s.ID && a.State == Alive {
					allyCount++
				}
			}
			bonus := math.Min(float64(allyCount)*0.06, 0.30)
			rawDamage *= 1.0 + bonus
		}

		// Goa'uld Bombardment: 2x damage while active
		if s.Faction == faction.Goauld && e.Powers.IsActive(faction.Goauld) {
			rawDamage *= 2.0
		}

		// Lucian Kassa Rush: +50% attack speed
		if s.Faction == faction.Lucian && e.Powers.IsActive(faction.Lucian) {
			rawDamage *= 1.5
		}

		// Goa'uld shield matrix: rear ships take 50% less
		if target.Faction == faction.Goauld {
			if !e.isFrontRow(target) {
				rawDamage *= 0.5
			}
		}

		// Jaffa Kree!: ignore incoming damage
		if target.Faction == faction.Jaffa && e.Powers.IsActive(faction.Jaffa) {
			continue // skip damage to Jaffa during Kree!
		}

		// Lucian Kassa Rush: take +25% more damage
		if target.Faction == faction.Lucian && e.Powers.IsActive(faction.Lucian) {
			rawDamage *= 1.25
		}

		// Underdog bonus
		rawDamage *= 1.0 + e.underdogBonus(s.Faction)

		target.HP -= float32(rawDamage)
		if target.HP <= 0 {
			target.State = Exploding
			target.HP = 0
			target.ExplodeFrame = 0
			target.ExplodeTicks = ExplodeTicksPerFrame
			e.killCounts[s.Faction]++
			e.deathCounts[target.Faction]++
		}
	}
}

// isFrontRow checks if a Goa'uld ship has no allies closer to its nearest enemy.
func (e *Engine) isFrontRow(s *Ship) bool {
	// Find nearest enemy
	var nearestEnemy *Ship
	minDist := math.MaxFloat64
	for _, other := range e.ships {
		if other.Faction == s.Faction || other.State != Alive {
			continue
		}
		d := Dist(s, other)
		if d < minDist {
			minDist = d
			nearestEnemy = other
		}
	}
	if nearestEnemy == nil {
		return true
	}

	// Check if any friendly ship is closer to that enemy
	myDist := Dist(s, nearestEnemy)
	nearby := e.spatial.QueryRadius(s.X, s.Y, 15)
	for _, ally := range nearby {
		if ally.Faction == s.Faction && ally.ID != s.ID && ally.State == Alive {
			if Dist(ally, nearestEnemy) < myDist {
				return false // ally is closer → we're behind
			}
		}
	}
	return true
}

func (e *Engine) underdogBonus(factionID int) float64 {
	maxPlayers := 0
	for _, c := range e.playerCounts {
		if c > maxPlayers {
			maxPlayers = c
		}
	}
	if maxPlayers > 0 && e.playerCounts[factionID] < maxPlayers/2 {
		return 0.15
	}
	return 0
}

func (e *Engine) processExplosions() {
	for i := range e.ships {
		s := e.ships[i]
		if s.State != Exploding {
			continue
		}
		s.ExplodeTicks--
		if s.ExplodeTicks <= 0 {
			s.ExplodeFrame++
			if s.ExplodeFrame >= ExplodeFrames {
				s.State = Dead
			} else {
				s.ExplodeTicks = ExplodeTicksPerFrame
			}
		}
	}
	// Convert new deaths to explosion effects
	for _, s := range e.ships {
		if s.State == Exploding && s.ExplodeFrame == 0 && s.ExplodeTicks == ExplodeTicksPerFrame {
			e.explosions = append(e.explosions, Explosion{
				X: s.X, Y: s.Y,
				Frame:     0,
				TicksLeft: ExplodeTicksPerFrame * ExplodeFrames,
				Faction:   s.Faction,
			})
		}
	}
}

func (e *Engine) removeDeadShips() {
	alive := e.ships[:0]
	for _, s := range e.ships {
		if s.State != Dead {
			alive = append(alive, s)
		}
	}
	e.ships = alive

	// Update explosion effects
	remaining := e.explosions[:0]
	for i := range e.explosions {
		e.explosions[i].TicksLeft--
		e.explosions[i].Frame = (ExplodeFrames*ExplodeTicksPerFrame - e.explosions[i].TicksLeft) / ExplodeTicksPerFrame
		if e.explosions[i].TicksLeft > 0 {
			remaining = append(remaining, e.explosions[i])
		}
	}
	e.explosions = remaining
}

// Snapshot returns a read-only copy of the simulation state.
func (e *Engine) Snapshot() Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ships := make([]Ship, len(e.ships))
	for i, s := range e.ships {
		ships[i] = *s
	}

	explosions := make([]Explosion, len(e.explosions))
	copy(explosions, e.explosions)

	notifications := make([]Notification, len(e.notifications))
	copy(notifications, e.notifications)

	var shipCounts [faction.Count]int
	for _, s := range e.ships {
		if s.State == Alive {
			shipCounts[s.Faction]++
		}
	}

	beams := make([]BeamEffect, len(e.beams))
	copy(beams, e.beams)

	var powerStatuses [faction.Count]PowerStatus
	for i := 0; i < faction.Count; i++ {
		powerStatuses[i] = e.Powers.Status(i)
	}

	focusVotes := e.calculateFocusVotes()

	snap := Snapshot{
		Ships:         ships,
		Explosions:    explosions,
		Beams:         beams,
		Notifications: notifications,
		Territory:     e.territory,
		PowerStatuses: powerStatuses,
		PlayerCounts:  e.playerCounts,
		ShipCounts:    shipCounts,
		KillCounts:    e.killCounts,
		DeathCounts:   e.deathCounts,
		FocusVotes:    focusVotes,
		Tick:          e.tick,
	}
	return snap
}

// RegisterPlayer adds a player to a faction. Returns true if this is a new unique player.
func (e *Engine) RegisterPlayer(sshKey string, factionID int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.players[sshKey]; exists {
		return false // already registered (multiplex session)
	}
	e.players[sshKey] = factionID
	e.playerCounts[factionID]++
	return true
}

// UnregisterPlayer removes a player.
func (e *Engine) UnregisterPlayer(sshKey string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if f, exists := e.players[sshKey]; exists {
		e.playerCounts[f]--
		if e.playerCounts[f] < 0 {
			e.playerCounts[f] = 0
		}
		delete(e.players, sshKey)
		delete(e.focusVotes, sshKey)
	}
}

// SetFocusSector sets a player's sector focus vote (-1 to clear).
func (e *Engine) SetFocusSector(sshKey string, sector int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if sector < 0 || sector >= faction.Count {
		delete(e.focusVotes, sshKey)
	} else {
		e.focusVotes[sshKey] = sector
	}
}

// AddNotification broadcasts a message to a faction.
func (e *Engine) AddNotification(factionID int, msg string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	e.notifications = append(e.notifications, Notification{
		Faction:   factionID,
		Message:   msg,
		CreatedAt: now,
		ExpiresAt: now.Add(4 * time.Second),
	})
}

// calculateFocusVotes tallies sector votes per faction.
func (e *Engine) calculateFocusVotes() [faction.Count][faction.Count]int {
	var votes [faction.Count][faction.Count]int
	for key, sector := range e.focusVotes {
		if f, ok := e.players[key]; ok {
			votes[f][sector]++
		}
	}
	return votes
}

// applyPowerMovementEffects modifies ship movement based on active powers.
func (e *Engine) applyPowerMovementEffects() {
	// Goa'uld Bombardment: freeze all Goa'uld ships
	if e.Powers.IsActive(faction.Goauld) {
		for _, s := range e.ships {
			if s.Faction == faction.Goauld && s.State == Alive {
				s.VX = 0
				s.VY = 0
			}
		}
	}

	// Jaffa Kree!: max speed toward nearest enemy
	if e.Powers.IsActive(faction.Jaffa) {
		for _, s := range e.ships {
			if s.Faction == faction.Jaffa && s.State == Alive {
				// Double speed
				vel := Vec2{s.VX, s.VY}
				speed := vel.Len()
				if speed > 0.1 {
					maxSpeed := float64(faction.Factions[faction.Jaffa].BaseSpeed) * 2.5
					vel = vel.Normalize().Scale(maxSpeed)
					s.VX = vel.X
					s.VY = vel.Y
				}
				s.Boosted = true
			}
		}
	} else {
		for _, s := range e.ships {
			if s.Faction == faction.Jaffa {
				s.Boosted = false
			}
		}
	}
}

// processIonCannon fires the Asgard ion cannon beam when first activated.
func (e *Engine) processIonCannon() {
	if !e.Powers.IsActive(faction.Asgard) {
		return
	}

	// Only fire on the first tick of activation (check if beam already exists)
	for _, b := range e.beams {
		if b.Faction == faction.Asgard {
			return // already fired this activation
		}
	}

	// Find a random Asgard ship to fire from
	var firingShip *Ship
	for _, s := range e.ships {
		if s.Faction == faction.Asgard && s.State == Alive {
			firingShip = s
			break
		}
	}
	if firingShip == nil {
		return
	}

	// Find the nearest enemy cluster center
	var ex, ey float64
	var enemyCount int
	for _, s := range e.ships {
		if s.Faction != faction.Asgard && s.State == Alive {
			ex += s.X
			ey += s.Y
			enemyCount++
		}
	}
	if enemyCount == 0 {
		return
	}
	ex /= float64(enemyCount)
	ey /= float64(enemyCount)

	// Create beam from firing ship toward enemy cluster, extending through
	dir := Vec2{ex - firingShip.X, ey - firingShip.Y}.Normalize()
	endX := firingShip.X + dir.X*500 // extend far past target
	endY := firingShip.Y + dir.Y*500

	e.beams = append(e.beams, BeamEffect{
		X1: firingShip.X, Y1: firingShip.Y,
		X2: endX, Y2: endY,
		Faction:   faction.Asgard,
		TicksLeft: 10, // 1 second
	})

	// Damage all enemies within 3 tiles of the beam line
	beamDmg := float64(faction.Factions[faction.Asgard].BaseDamage) * 3
	for _, s := range e.ships {
		if s.Faction == faction.Asgard || s.State != Alive {
			continue
		}
		if distToLine(s.X, s.Y, firingShip.X, firingShip.Y, ex, ey) < 3.0 {
			s.HP -= float32(beamDmg)
			if s.HP <= 0 {
				s.State = Exploding
				s.HP = 0
				s.ExplodeFrame = 0
				s.ExplodeTicks = ExplodeTicksPerFrame
				e.killCounts[faction.Asgard]++
				e.deathCounts[s.Faction]++
			}
		}
	}
}

func (e *Engine) processBeams() {
	remaining := e.beams[:0]
	for i := range e.beams {
		e.beams[i].TicksLeft--
		if e.beams[i].TicksLeft > 0 {
			remaining = append(remaining, e.beams[i])
		}
	}
	e.beams = remaining
}

// distToLine returns the distance from point (px,py) to line segment (x1,y1)-(x2,y2).
func distToLine(px, py, x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	lenSq := dx*dx + dy*dy
	if lenSq < 0.0001 {
		return math.Sqrt((px-x1)*(px-x1) + (py-y1)*(py-y1))
	}
	t := ((px-x1)*dx + (py-y1)*dy) / lenSq
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	closestX := x1 + t*dx
	closestY := y1 + t*dy
	return math.Sqrt((px-closestX)*(px-closestX) + (py-closestY)*(py-closestY))
}

// SpatialHash is a grid-based spatial index for fast neighbor queries.
type SpatialHash struct {
	cells [GridW][GridH][]*Ship
}

func (sh *SpatialHash) Clear() {
	for x := range sh.cells {
		for y := range sh.cells[x] {
			sh.cells[x][y] = sh.cells[x][y][:0]
		}
	}
}

func (sh *SpatialHash) Insert(s *Ship) {
	cx := int(s.X) / CellSize
	cy := int(s.Y) / CellSize
	if cx < 0 {
		cx = 0
	} else if cx >= GridW {
		cx = GridW - 1
	}
	if cy < 0 {
		cy = 0
	} else if cy >= GridH {
		cy = GridH - 1
	}
	sh.cells[cx][cy] = append(sh.cells[cx][cy], s)
}

func (sh *SpatialHash) QueryRadius(x, y, radius float64) []*Ship {
	minCX := int((x - radius) / CellSize)
	maxCX := int((x + radius) / CellSize)
	minCY := int((y - radius) / CellSize)
	maxCY := int((y + radius) / CellSize)

	if minCX < 0 {
		minCX = 0
	}
	if maxCX >= GridW {
		maxCX = GridW - 1
	}
	if minCY < 0 {
		minCY = 0
	}
	if maxCY >= GridH {
		maxCY = GridH - 1
	}

	var result []*Ship
	r2 := radius * radius
	for cx := minCX; cx <= maxCX; cx++ {
		for cy := minCY; cy <= maxCY; cy++ {
			for _, s := range sh.cells[cx][cy] {
				dx := s.X - x
				dy := s.Y - y
				if dx*dx+dy*dy <= r2 {
					result = append(result, s)
				}
			}
		}
	}
	return result
}
