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

// Snapshot is a read-only copy of simulation state for rendering.
type Snapshot struct {
	Ships         []Ship
	Explosions    []Explosion
	Notifications []Notification
	Territory     *TerritoryMap
	PlayerCounts  [faction.Count]int
	ShipCounts    [faction.Count]int
	KillCounts    [faction.Count]int
	DeathCounts   [faction.Count]int
	Tick          uint64
}

// Engine runs the core simulation loop.
type Engine struct {
	mu sync.RWMutex

	ships      []*Ship
	explosions []Explosion
	nextID     uint64
	tick       uint64

	territory *TerritoryMap
	spatial   SpatialHash

	playerCounts [faction.Count]int
	killCounts   [faction.Count]int
	deathCounts  [faction.Count]int

	notifications []Notification

	// Unique player tracking (SSH key -> faction)
	players map[string]int

	rng *rand.Rand
}

// NewEngine creates a simulation engine.
func NewEngine() *Engine {
	e := &Engine{
		territory: NewTerritoryMap(),
		players:   make(map[string]int),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
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

	// Heading toward center with random spread
	center := Vec2{WorldW / 2, WorldH / 2}
	dir := center.Sub(gate).Normalize()

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

			// 4. Combat
			e.resolveCombat(dt)

			// 5. Process explosions
			e.processExplosions()

			// 6. Remove dead ships
			e.removeDeadShips()

			// 7. Territory (every 10 ticks = 1 second)
			territoryTimer++
			if territoryTimer >= TickRate {
				territoryTimer = 0
				e.territory = CalculateTerritory(e.ships)
			}

			// 8. Expire notifications
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
	for _, s := range e.ships {
		if s.State != Alive {
			continue
		}

		enemies := e.spatial.QueryRadius(s.X, s.Y, AttackRange)
		var target *Ship
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

		// Goa'uld shield matrix: rear ships take 50% less
		if target.Faction == faction.Goauld {
			if !e.isFrontRow(target) {
				rawDamage *= 0.5
			}
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

	snap := Snapshot{
		Ships:         ships,
		Explosions:    explosions,
		Notifications: notifications,
		Territory:     e.territory,
		PlayerCounts:  e.playerCounts,
		ShipCounts:    shipCounts,
		KillCounts:    e.killCounts,
		DeathCounts:   e.deathCounts,
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
	}
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
