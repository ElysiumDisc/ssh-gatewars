package engine

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/game"
)

// GameEvent is sent from the engine to the chat hub.
type GameEvent struct {
	Type       GameEventType
	PlayerFP   string
	Callsign   string
	PlanetName string
	Extra      string
}

type GameEventType int

const (
	GamePlayerDeploy    GameEventType = iota // player deployed chair
	GamePlayerRetreat                        // player left planet
	GamePlanetLiberated                      // planet freed!
	GamePlanetFailed                         // all chairs destroyed
	GamePlayerConnect
	GamePlayerDisconnect
	GameSurgeStart  // replicator surge begins
	GameSurgeEnd    // replicator surge ends
	GameMilestone   // galaxy liberation milestone
	GameGalaxyReset // new game+ cycle
)

// Engine manages the game simulation: galaxy, planet defenses, tick loop.
type Engine struct {
	cfg        core.GameConfig
	Galaxy     *game.Galaxy
	instances  map[int]*DefenseInstance // planet ID → instance
	mu         sync.RWMutex
	GameEvents chan GameEvent

	// Atomic snapshot for lock-free TUI reads
	galaxySnap atomic.Value // *GalaxySnapshot

	// Surge system
	surgeTimer    int // ticks until next surge rotation
	surgePlanetID int // currently surging planet (-1 = none)
	rng           *rand.Rand

	// Milestone tracking
	lastMilestonePct int // last announced milestone (0, 25, 50, 75, 100)
	cycle            int // new game+ cycle number
}

// GalaxySnapshot is a read-only view of the galaxy for TUI rendering.
type GalaxySnapshot struct {
	Planets   []PlanetSnap
	Cycle     int // new game+ cycle
	FreePct   int // percentage of galaxy liberated
	SurgeID   int // planet ID currently surging (-1 = none)
}

type PlanetSnap struct {
	ID            int
	Name          string
	Status        game.PlanetStatus
	InvasionLevel int
	DefenderCount int
	Surging       bool
	BountyZPM     int
}

// NewEngine creates the game engine.
func NewEngine(cfg core.GameConfig) *Engine {
	seed := cfg.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	galaxy := game.NewGalaxy(seed, cfg.NumPlanets)

	// Calculate initial bounties
	for _, p := range galaxy.Planets {
		p.BountyZPM = p.InvasionLevel * 10
	}

	e := &Engine{
		cfg:           cfg,
		Galaxy:        galaxy,
		instances:     make(map[int]*DefenseInstance),
		GameEvents:    make(chan GameEvent, 100),
		surgePlanetID: -1,
		surgeTimer:    cfg.SurgeIntervalSec * cfg.TickRate, // first surge after interval
		rng:           rand.New(rand.NewSource(seed)),
		cycle:         1,
	}

	e.updateGalaxySnapshot()
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
			e.tick()
		}
	}
}

func (e *Engine) tick() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for planetID, inst := range e.instances {
		inst.Tick()

		if inst.Liberated {
			e.Galaxy.FreePlanet(planetID)
			e.emitEvent(GameEvent{
				Type:       GamePlanetLiberated,
				PlanetName: inst.Planet.Name,
			})
			delete(e.instances, planetID)
			log.Info("planet liberated", "name", inst.Planet.Name)
		} else if inst.Failed {
			e.emitEvent(GameEvent{
				Type:       GamePlanetFailed,
				PlanetName: inst.Planet.Name,
			})
			delete(e.instances, planetID)
			log.Info("planet defense failed", "name", inst.Planet.Name)
		} else if len(inst.Chairs) == 0 {
			delete(e.instances, planetID)
		}
	}

	// Surge rotation
	e.tickSurge()

	// Milestone checks
	e.tickMilestones()

	e.updateGalaxySnapshot()
}

func (e *Engine) tickSurge() {
	e.surgeTimer--
	if e.surgeTimer > 0 {
		return
	}
	e.surgeTimer = e.cfg.SurgeIntervalSec * e.cfg.TickRate

	// Clear current surge
	if e.surgePlanetID >= 0 && e.surgePlanetID < len(e.Galaxy.Planets) {
		old := e.Galaxy.Planets[e.surgePlanetID]
		old.Surging = false
		e.emitEvent(GameEvent{
			Type:       GameSurgeEnd,
			PlanetName: old.Name,
		})
	}
	e.surgePlanetID = -1

	// Pick a new invaded planet to surge
	var candidates []int
	for _, p := range e.Galaxy.Planets {
		if p.Status == game.PlanetInvaded {
			candidates = append(candidates, p.ID)
		}
	}
	if len(candidates) == 0 {
		return // all free or contested, no surge
	}

	pick := candidates[e.rng.Intn(len(candidates))]
	e.surgePlanetID = pick
	planet := e.Galaxy.Planets[pick]
	planet.Surging = true

	e.emitEvent(GameEvent{
		Type:       GameSurgeStart,
		PlanetName: planet.Name,
	})
	log.Info("replicator surge", "planet", planet.Name)
}

func (e *Engine) tickMilestones() {
	total := len(e.Galaxy.Planets)
	freed := e.Galaxy.FreeCount()
	pct := 0
	if total > 0 {
		pct = freed * 100 / total
	}

	// Check milestones: 25, 50, 75, 100
	milestones := []int{25, 50, 75, 100}
	for _, m := range milestones {
		if pct >= m && e.lastMilestonePct < m {
			e.lastMilestonePct = m
			if m == 100 {
				e.emitEvent(GameEvent{
					Type:  GameGalaxyReset,
					Extra: "The galaxy is free! A new threat emerges...",
				})
				e.resetGalaxy()
			} else {
				e.emitEvent(GameEvent{
					Type:  GameMilestone,
					Extra: fmt.Sprintf("Galaxy %d%% liberated!", m),
				})
			}
		}
	}
}

// resetGalaxy performs a New Game+ — regenerate all planets with higher difficulty.
func (e *Engine) resetGalaxy() {
	e.cycle++
	seed := e.rng.Int63()
	e.Galaxy = game.NewGalaxy(seed, e.cfg.NumPlanets)

	// Scale difficulty by cycle
	for _, p := range e.Galaxy.Planets {
		p.InvasionLevel = p.InvasionLevel + e.cycle - 1
		if p.InvasionLevel > 15 {
			p.InvasionLevel = 15
		}
		p.BountyZPM = p.InvasionLevel * 10 * e.cycle
	}

	// Clear state
	e.instances = make(map[int]*DefenseInstance)
	e.surgePlanetID = -1
	e.surgeTimer = e.cfg.SurgeIntervalSec * e.cfg.TickRate
	e.lastMilestonePct = 0

	log.Info("galaxy reset", "cycle", e.cycle)
}

// DeployChair places a player on a planet.
func (e *Engine) DeployChair(planetID int, fp, callsign string, level int, tier game.DroneTier, faction game.Faction) *DefenseInstance {
	e.mu.Lock()
	defer e.mu.Unlock()

	planet := e.Galaxy.Planets[planetID]
	planet.Status = game.PlanetContested

	inst, ok := e.instances[planetID]
	if !ok {
		inst = NewDefenseInstance(planet, e.cfg)
		e.instances[planetID] = inst
	}

	inst.AddChair(fp, callsign, level, tier, faction)
	planet.DefenderCount = len(inst.Chairs)

	e.emitEvent(GameEvent{
		Type:       GamePlayerDeploy,
		PlayerFP:   fp,
		Callsign:   callsign,
		PlanetName: planet.Name,
	})

	return inst
}

// RetreatChair removes a player from a planet.
func (e *Engine) RetreatChair(planetID int, fp string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if planetID < 0 || planetID >= len(e.Galaxy.Planets) {
		return
	}

	inst, ok := e.instances[planetID]
	if !ok {
		return
	}

	inst.RemoveChair(fp)
	planet := e.Galaxy.Planets[planetID]
	planet.DefenderCount = len(inst.Chairs)

	if len(inst.Chairs) == 0 {
		planet.Status = game.PlanetInvaded
		delete(e.instances, planetID)
	}
}

// SetChairTactic changes a player's targeting tactic on a planet.
func (e *Engine) SetChairTactic(planetID int, fp string, tactic game.DroneTactic) {
	e.mu.Lock()
	defer e.mu.Unlock()

	inst, ok := e.instances[planetID]
	if !ok {
		return
	}
	inst.SetChairTactic(fp, tactic)
}

// GetDefenseSnapshot returns a snapshot of a planet's defense state.
func (e *Engine) GetDefenseSnapshot(planetID int) *DefenseSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	inst, ok := e.instances[planetID]
	if !ok {
		return nil
	}
	snap := inst.Snapshot()
	return &snap
}

// GetGalaxySnapshot returns the latest galaxy snapshot (lock-free).
func (e *Engine) GetGalaxySnapshot() *GalaxySnapshot {
	v := e.galaxySnap.Load()
	if v == nil {
		return nil
	}
	return v.(*GalaxySnapshot)
}

func (e *Engine) updateGalaxySnapshot() {
	total := len(e.Galaxy.Planets)
	freed := e.Galaxy.FreeCount()
	freePct := 0
	if total > 0 {
		freePct = freed * 100 / total
	}

	snap := &GalaxySnapshot{
		Planets: make([]PlanetSnap, total),
		Cycle:   e.cycle,
		FreePct: freePct,
		SurgeID: e.surgePlanetID,
	}
	for i, p := range e.Galaxy.Planets {
		snap.Planets[i] = PlanetSnap{
			ID:            p.ID,
			Name:          p.Name,
			Status:        p.Status,
			InvasionLevel: p.InvasionLevel,
			DefenderCount: p.DefenderCount,
			Surging:       p.Surging,
			BountyZPM:     p.BountyZPM,
		}
	}
	e.galaxySnap.Store(snap)
}

func (e *Engine) emitEvent(ev GameEvent) {
	select {
	case e.GameEvents <- ev:
	default:
	}
}

// OnlinePlayerCount returns total chairs across all instances.
func (e *Engine) OnlinePlayerCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	count := 0
	for _, inst := range e.instances {
		count += len(inst.Chairs)
	}
	return count
}
