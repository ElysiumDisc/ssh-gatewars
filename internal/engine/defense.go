package engine

import (
	"math"
	"math/rand"
	"sort"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/game"
)

// DefenseInstance manages the real-time defense of a single planet.
type DefenseInstance struct {
	Planet      *game.Planet
	Chairs      []*game.Chair
	Drones      []*game.Drone
	Replicators []*game.Replicator
	cfg         core.GameConfig

	// State
	HoldTicks    int // ticks held so far
	HoldRequired int // ticks needed to liberate
	WaveNum      int
	WaveTimer    int // ticks until next wave
	NextID       int // entity ID counter
	TotalKills   int
	ZPMEarned    int
	rng          *rand.Rand
	Liberated    bool
	Failed       bool // all chairs destroyed

	// Network bonuses (set by engine from gate link levels)
	NetDamageBoost    float64 // +% damage for drones
	NetSpawnReduction float64 // -% replicator spawns
	NetShieldRegen    float64 // fraction of max shield restored per tick

	// Resource transfer bonuses
	BonusDrones   int // extra drone slots from transfers
	BonusDroneTTL int // ticks remaining for bonus drones

	// Faction ZPM multiplier (set from first deploying faction)
	ZPMEarnMult float64

	// Passive ability timer (per-chair)
	passiveTimers map[string]int // playerFP → ticks until next passive
}

// NewDefenseInstance creates a defense session for a planet.
func NewDefenseInstance(planet *game.Planet, cfg core.GameConfig) *DefenseInstance {
	return &DefenseInstance{
		Planet:        planet,
		Chairs:        make([]*game.Chair, 0),
		Drones:        make([]*game.Drone, 0),
		Replicators:   make([]*game.Replicator, 0),
		cfg:           cfg,
		HoldTicks:     0,
		HoldRequired:  0,
		WaveNum:       0,
		WaveTimer:     cfg.TickRate * 3, // first wave after 3 seconds
		NextID:        1,
		rng:           rand.New(rand.NewSource(planet.Seed)),
		passiveTimers: make(map[string]int),
	}
}

// AddChair places a player's chair and recalculates hold time.
func (d *DefenseInstance) AddChair(fp, callsign string, level int, tier game.DroneTier, faction game.Faction) {
	// Position chairs near center, spread out
	idx := len(d.Chairs)
	spacing := 2.0
	offsetX := float64(idx-(len(d.Chairs)/2)) * spacing
	pos := core.Vec2{X: offsetX, Y: 0}

	chair := game.NewChair(fp, callsign, pos, level, tier, faction)
	d.Chairs = append(d.Chairs, chair)
	d.recalcHoldTime()

	// Set ZPM earn multiplier from faction (use highest among defenders)
	fDef := game.FactionDefs[faction]
	if fDef.ZPMEarnMult > d.ZPMEarnMult {
		d.ZPMEarnMult = fDef.ZPMEarnMult
	}
}

// RemoveChair removes a player's chair.
func (d *DefenseInstance) RemoveChair(fp string) {
	for i, c := range d.Chairs {
		if c.PlayerFP == fp {
			d.Chairs = append(d.Chairs[:i], d.Chairs[i+1:]...)
			break
		}
	}
	d.recalcHoldTime()
}

// SetChairTactic changes a chair's targeting tactic.
func (d *DefenseInstance) SetChairTactic(fp string, tactic game.DroneTactic) {
	for _, c := range d.Chairs {
		if c.PlayerFP == fp {
			c.Tactic = tactic
			return
		}
	}
}

func (d *DefenseInstance) recalcHoldTime() {
	n := len(d.Chairs)
	if n == 0 {
		d.HoldRequired = 0
		return
	}
	// 5 minutes per player (at 10Hz that's 3000 ticks per player)
	d.HoldRequired = n * d.cfg.HoldTimeSec * d.cfg.TickRate
}

// Tick advances the defense simulation by one step.
func (d *DefenseInstance) Tick() {
	if d.Liberated || d.Failed {
		return
	}
	if len(d.Chairs) == 0 {
		return
	}

	d.HoldTicks++

	// Check liberation
	if d.HoldRequired > 0 && d.HoldTicks >= d.HoldRequired {
		d.Liberated = true
		return
	}

	// Network bonus: shield regen
	if d.NetShieldRegen > 0 {
		for _, c := range d.Chairs {
			if c.Alive() && c.ShieldHP < c.MaxShield {
				regen := int(d.NetShieldRegen * float64(c.MaxShield))
				if regen < 1 {
					regen = 1
				}
				c.ShieldHP += regen
				if c.ShieldHP > c.MaxShield {
					c.ShieldHP = c.MaxShield
				}
			}
		}
	}

	// Bonus drone TTL countdown
	if d.BonusDroneTTL > 0 {
		d.BonusDroneTTL--
		if d.BonusDroneTTL == 0 {
			d.BonusDrones = 0
		}
	}

	// Spawn waves
	d.WaveTimer--
	if d.WaveTimer <= 0 {
		d.spawnWave()
		d.WaveNum++
		// Next wave comes faster as difficulty increases
		interval := 50 - d.WaveNum*2
		if interval < 15 {
			interval = 15
		}
		d.WaveTimer = interval
	}

	// Chair auto-fire with salvo support
	for _, chair := range d.Chairs {
		if !chair.Alive() {
			continue
		}
		chair.FireTimer--
		if chair.FireTimer <= 0 {
			d.fireFromChair(chair)
			chair.FireTimer = chair.EffectiveFireRate(d.cfg.DroneFireRate)
		}
	}

	// Move drones
	d.moveDrones()

	// Move replicators
	d.moveReplicators()

	// Collisions: drones vs replicators
	d.checkCollisions()

	// Replicators reaching chairs
	d.checkBreaches()

	// Faction passive abilities
	d.tickPassives()

	// Clean dead entities
	d.cleanup()

	// Check if all chairs destroyed
	allDead := true
	for _, c := range d.Chairs {
		if c.Alive() {
			allDead = false
			break
		}
	}
	if allDead {
		d.Failed = true
	}
}

func (d *DefenseInstance) spawnWave() {
	count := 3 + d.WaveNum*2 + d.Planet.InvasionLevel
	if count > 30 {
		count = 30
	}

	// Surge: double spawn count
	if d.Planet.Surging {
		count *= 2
	}

	// Network bonus: spawn reduction
	if d.NetSpawnReduction > 0 {
		reduction := int(float64(count) * d.NetSpawnReduction)
		count -= reduction
		if count < 1 {
			count = 1
		}
	}

	for i := 0; i < count; i++ {
		// Spawn at random angle on the spawn radius
		angle := d.rng.Float64() * 2 * math.Pi
		pos := core.Vec2{
			X: math.Cos(angle) * d.cfg.SpawnRadius,
			Y: math.Sin(angle) * d.cfg.SpawnRadius,
		}

		// Pick type based on wave
		rtype := game.ReplicatorBasic
		roll := d.rng.Float64()
		if d.WaveNum >= 5 && roll < 0.1 {
			rtype = game.ReplicatorQueen
		} else if d.WaveNum >= 2 && roll < 0.3 {
			rtype = game.ReplicatorArmored
		}

		def := game.ReplicatorDefs[rtype]
		// Velocity points toward center
		dir := core.Vec2{X: -pos.X, Y: -pos.Y}
		dist := math.Sqrt(dir.LenSq())
		if dist > 0 {
			speed := d.cfg.ReplicatorSpeed * def.Speed
			dir = core.Vec2{X: dir.X / dist * speed, Y: dir.Y / dist * speed}
		}

		d.Replicators = append(d.Replicators, &game.Replicator{
			ID:    d.NextID,
			Type:  rtype,
			Pos:   pos,
			Vel:   dir,
			HP:    def.HP,
			MaxHP: def.HP,
			Alive: true,
		})
		d.NextID++
	}
}

func (d *DefenseInstance) fireFromChair(chair *game.Chair) {
	// Count active drones for THIS chair (per-owner)
	activeDrones := 0
	for _, dr := range d.Drones {
		if dr.Alive && dr.OwnerFP == chair.PlayerFP {
			activeDrones++
		}
	}

	salvo := chair.SalvoCount()
	maxDrones := chair.MaxDrones + d.BonusDrones
	available := maxDrones - activeDrones
	if available <= 0 {
		return
	}
	if salvo > available {
		salvo = available
	}

	// Select targets based on tactic
	targets := d.selectTargets(chair, salvo)
	if len(targets) == 0 {
		return
	}

	tierCfg := game.DroneTiers[chair.DroneTier]
	dmg := chair.DroneDamage() // faction-adjusted damage
	// Network bonus: damage boost
	if d.NetDamageBoost > 0 {
		dmg = int(float64(dmg) * (1 + d.NetDamageBoost))
		if dmg < 1 {
			dmg = 1
		}
	}

	for _, target := range targets {
		dir := target.Pos.Sub(chair.Pos)
		dist := math.Sqrt(dir.LenSq())
		speed := d.cfg.DroneSpeed * tierCfg.Speed
		if dist > 0 {
			dir = core.Vec2{X: dir.X / dist * speed, Y: dir.Y / dist * speed}
		}

		d.Drones = append(d.Drones, &game.Drone{
			ID:       d.NextID,
			OwnerFP:  chair.PlayerFP,
			Pos:      chair.Pos,
			Vel:      dir,
			TargetID: target.ID,
			Tier:     chair.DroneTier,
			Damage:   dmg,
			Alive:    true,
		})
		d.NextID++
	}
}

// selectTargets picks targets based on the chair's tactic.
func (d *DefenseInstance) selectTargets(chair *game.Chair, count int) []*game.Replicator {
	alive := d.aliveReplicators()
	if len(alive) == 0 {
		return nil
	}

	switch chair.Tactic {
	case game.TacticFocus:
		// All drones focus the strongest enemy (queens first, then nearest)
		target := d.findStrongestTarget(chair.Pos, alive)
		if target == nil {
			return nil
		}
		result := make([]*game.Replicator, count)
		for i := range result {
			result[i] = target
		}
		return result

	case game.TacticPerimeter:
		// Prioritize enemies closest to center (deepest penetration)
		return d.findClosestToCenter(alive, count)

	default: // TacticSpread
		// Target nearest distinct enemies
		return d.findNearestTargets(chair.Pos, alive, count)
	}
}

func (d *DefenseInstance) aliveReplicators() []*game.Replicator {
	var alive []*game.Replicator
	for _, r := range d.Replicators {
		if r.Alive {
			alive = append(alive, r)
		}
	}
	return alive
}

func (d *DefenseInstance) findStrongestTarget(from core.Vec2, alive []*game.Replicator) *game.Replicator {
	if len(alive) == 0 {
		return nil
	}
	// Sort by threat: queen > armored > basic, then by distance
	sort.Slice(alive, func(i, j int) bool {
		if alive[i].Type != alive[j].Type {
			return alive[i].Type > alive[j].Type // queen(2) > armored(1) > basic(0)
		}
		return from.Dist(alive[i].Pos) < from.Dist(alive[j].Pos)
	})
	return alive[0]
}

func (d *DefenseInstance) findClosestToCenter(alive []*game.Replicator, count int) []*game.Replicator {
	center := core.Vec2{}
	sort.Slice(alive, func(i, j int) bool {
		return center.Dist(alive[i].Pos) < center.Dist(alive[j].Pos)
	})
	if count > len(alive) {
		count = len(alive)
	}
	return alive[:count]
}

func (d *DefenseInstance) findNearestTargets(from core.Vec2, alive []*game.Replicator, count int) []*game.Replicator {
	sort.Slice(alive, func(i, j int) bool {
		return from.Dist(alive[i].Pos) < from.Dist(alive[j].Pos)
	})
	if count > len(alive) {
		count = len(alive)
	}
	return alive[:count]
}

func (d *DefenseInstance) moveDrones() {
	// Build a set of which factions can retarget
	canRetarget := make(map[string]bool)
	for _, c := range d.Chairs {
		fDef := game.FactionDefs[c.Faction]
		canRetarget[c.PlayerFP] = fDef.DroneRetarget
	}

	for _, dr := range d.Drones {
		if !dr.Alive {
			continue
		}
		// Only re-aim if the owner's faction allows retargeting (Ancient=yes, Ori=no)
		if canRetarget[dr.OwnerFP] {
			for _, r := range d.Replicators {
				if r.ID == dr.TargetID && r.Alive {
					dir := r.Pos.Sub(dr.Pos)
					dist := math.Sqrt(dir.LenSq())
					tierCfg := game.DroneTiers[dr.Tier]
					speed := d.cfg.DroneSpeed * tierCfg.Speed
					if dist > 0.1 {
						dr.Vel = core.Vec2{X: dir.X / dist * speed, Y: dir.Y / dist * speed}
					}
					break
				}
			}
		}
		dr.Pos = dr.Pos.Add(dr.Vel)

		// Kill drones that fly too far
		if math.Sqrt(dr.Pos.LenSq()) > d.cfg.SpawnRadius*1.5 {
			dr.Alive = false
		}
	}
}

func (d *DefenseInstance) moveReplicators() {
	for _, r := range d.Replicators {
		if !r.Alive {
			continue
		}
		r.Pos = r.Pos.Add(r.Vel)
	}
}

func (d *DefenseInstance) checkCollisions() {
	hitRadius := 0.8

	zpmEarnMult := d.ZPMEarnMult
	if zpmEarnMult < 1.0 {
		zpmEarnMult = 1.0
	}

	for _, dr := range d.Drones {
		if !dr.Alive {
			continue
		}
		for _, r := range d.Replicators {
			if !r.Alive {
				continue
			}
			if dr.Pos.Dist(r.Pos) < hitRadius {
				r.HP -= dr.Damage
				zpmMult := 1
				if d.Planet.Surging {
					zpmMult = 2
				}
				if r.HP <= 0 {
					r.Alive = false
					def := game.ReplicatorDefs[r.Type]
					d.TotalKills++
					d.ZPMEarned += int(float64(def.ZPMDrop*zpmMult) * zpmEarnMult)
				}

				// Splash damage
				tierCfg := game.DroneTiers[dr.Tier]
				if tierCfg.Splash > 0 {
					for _, r2 := range d.Replicators {
						if !r2.Alive || r2.ID == r.ID {
							continue
						}
						if dr.Pos.Dist(r2.Pos) < tierCfg.Splash {
							r2.HP -= dr.Damage
							if r2.HP <= 0 {
								r2.Alive = false
								def := game.ReplicatorDefs[r2.Type]
								d.TotalKills++
								d.ZPMEarned += int(float64(def.ZPMDrop*zpmMult) * zpmEarnMult)
							}
						}
					}
				}

				// Piercing drones pass through, others die
				if !tierCfg.Pierce {
					dr.Alive = false
				}
				break
			}
		}
	}
}

func (d *DefenseInstance) checkBreaches() {
	breachRadius := 1.5

	for _, r := range d.Replicators {
		if !r.Alive {
			continue
		}
		for _, c := range d.Chairs {
			if !c.Alive() {
				continue
			}
			if r.Pos.Dist(c.Pos) < breachRadius {
				def := game.ReplicatorDefs[r.Type]
				c.TakeDamage(def.Damage)
				r.Alive = false
				break
			}
		}
	}
}

func (d *DefenseInstance) cleanup() {
	// Clean dead drones
	alive := d.Drones[:0]
	for _, dr := range d.Drones {
		if dr.Alive {
			alive = append(alive, dr)
		}
	}
	d.Drones = alive

	// Clean dead replicators
	aliveR := d.Replicators[:0]
	for _, r := range d.Replicators {
		if r.Alive {
			aliveR = append(aliveR, r)
		}
	}
	d.Replicators = aliveR
}

// tickPassives fires faction-specific passive abilities for eligible chairs.
func (d *DefenseInstance) tickPassives() {
	for _, chair := range d.Chairs {
		if !chair.Alive() {
			continue
		}
		fDef := game.FactionDefs[chair.Faction]
		if chair.Level < fDef.PassiveUnlockLv {
			continue
		}
		if fDef.PassiveInterval <= 0 {
			continue
		}

		// Initialize timer if missing
		if _, ok := d.passiveTimers[chair.PlayerFP]; !ok {
			d.passiveTimers[chair.PlayerFP] = fDef.PassiveInterval
		}

		d.passiveTimers[chair.PlayerFP]--
		if d.passiveTimers[chair.PlayerFP] > 0 {
			continue
		}
		d.passiveTimers[chair.PlayerFP] = fDef.PassiveInterval

		// Ancient: Ascension Pulse — heal all friendly chairs
		if fDef.PassiveShieldPulse > 0 {
			for _, c := range d.Chairs {
				if c.Alive() && c.ShieldHP < c.MaxShield {
					c.ShieldHP += fDef.PassiveShieldPulse
					if c.ShieldHP > c.MaxShield {
						c.ShieldHP = c.MaxShield
					}
				}
			}
		}

		// Ori: Prior's Wrath — AOE damage to nearby replicators
		if fDef.PassiveWrathDamage > 0 && fDef.PassiveWrathRadius > 0 {
			zpmMult := 1
			if d.Planet.Surging {
				zpmMult = 2
			}
			earnMult := d.ZPMEarnMult
			if earnMult < 1.0 {
				earnMult = 1.0
			}
			for _, r := range d.Replicators {
				if !r.Alive {
					continue
				}
				if chair.Pos.Dist(r.Pos) <= fDef.PassiveWrathRadius {
					r.HP -= fDef.PassiveWrathDamage
					if r.HP <= 0 {
						r.Alive = false
						def := game.ReplicatorDefs[r.Type]
						d.TotalKills++
						d.ZPMEarned += int(float64(def.ZPMDrop*zpmMult) * earnMult)
					}
				}
			}
		}
	}
}

// ── Snapshots ─────────────────────────────────────────────────────────

// DefenseSnapshot is a read-only copy for rendering.
type DefenseSnapshot struct {
	PlanetName   string
	PlanetID     int
	Chairs       []ChairSnap
	Drones       []DroneSnap
	Replicators  []ReplicatorSnap
	HoldTicks    int
	HoldRequired int
	WaveNum      int
	TotalKills   int
	ZPMEarned    int
	Liberated    bool
	Failed       bool
	Surging      bool
	BountyZPM    int
}

type ChairSnap struct {
	PlayerFP  string
	Callsign  string
	Pos       core.Vec2
	Level     int
	ShieldHP  int
	MaxShield int
	DroneTier game.DroneTier
	Tactic    game.DroneTactic
	Faction   game.Faction
}

type DroneSnap struct {
	Pos  core.Vec2
	Tier game.DroneTier
}

type ReplicatorSnap struct {
	Pos  core.Vec2
	Type game.ReplicatorType
	HP   int
}

func (d *DefenseInstance) Snapshot() DefenseSnapshot {
	snap := DefenseSnapshot{
		PlanetName:   d.Planet.Name,
		PlanetID:     d.Planet.ID,
		HoldTicks:    d.HoldTicks,
		HoldRequired: d.HoldRequired,
		WaveNum:      d.WaveNum,
		TotalKills:   d.TotalKills,
		ZPMEarned:    d.ZPMEarned,
		Liberated:    d.Liberated,
		Failed:       d.Failed,
		Surging:      d.Planet.Surging,
		BountyZPM:    d.Planet.BountyZPM,
		Chairs:       make([]ChairSnap, len(d.Chairs)),
		Drones:       make([]DroneSnap, len(d.Drones)),
		Replicators:  make([]ReplicatorSnap, len(d.Replicators)),
	}

	for i, c := range d.Chairs {
		snap.Chairs[i] = ChairSnap{
			PlayerFP: c.PlayerFP, Callsign: c.Callsign,
			Pos: c.Pos, Level: c.Level,
			ShieldHP: c.ShieldHP, MaxShield: c.MaxShield,
			DroneTier: c.DroneTier, Tactic: c.Tactic,
			Faction: c.Faction,
		}
	}
	for i, dr := range d.Drones {
		snap.Drones[i] = DroneSnap{Pos: dr.Pos, Tier: dr.Tier}
	}
	for i, r := range d.Replicators {
		snap.Replicators[i] = ReplicatorSnap{Pos: r.Pos, Type: r.Type, HP: r.HP}
	}

	return snap
}
