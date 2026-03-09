package simulation

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/gamedata"
)

const TickRate = 10

// FactionState holds per-faction mutable state.
type FactionState struct {
	Naquadah float64
	Tech     *FactionTech
	Designs  []*ShipDesign
}

// Engine runs the core simulation loop.
type Engine struct {
	mu sync.RWMutex

	galaxy    *Galaxy
	colonies  map[int]*Colony
	factions  [faction.Count]*FactionState
	fleets    []*Fleet
	combats   []*Combat
	diplomacy *DiplomacyState
	campaign  *CampaignState
	actions   *ActionQueue

	notifications []Notification

	players      map[string]int
	playerCounts [faction.Count]int

	tick uint64
	rng  *rand.Rand
}

// NewEngine creates a simulation engine with a fresh galaxy.
func NewEngine(seed int64, systemCount int) *Engine {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	e := &Engine{
		colonies: make(map[int]*Colony),
		players:  make(map[string]int),
		actions:  NewActionQueue(),
		rng:      rand.New(rand.NewSource(seed)),
	}

	e.galaxy = NewGalaxy(seed, systemCount)
	e.campaign = NewCampaign(seed, systemCount)
	e.diplomacy = NewDiplomacyState()

	for i := range e.factions {
		e.factions[i] = &FactionState{
			Naquadah: 200.0,
			Tech:     NewFactionTech(),
		}
	}

	// Create homeworld colonies (systems 0-4)
	for i := 0; i < faction.Count; i++ {
		sys := e.galaxy.Systems[i]
		if sys.Planet != nil {
			maxPop := gamedata.MaxPop(sys.Planet.Type, sys.Planet.Size)
			col := NewColony(sys.ID, i, maxPop)
			col.UpdateMaxFactory(faction.Factions[i].FactoryCapMod)
			e.colonies[sys.ID] = col
		}
	}

	return e
}

// Run starts the tick loop. Call in a goroutine.
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second / TickRate)
	defer ticker.Stop()

	economyTimer := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.mu.Lock()

			totalPlayers := 0
			for _, c := range e.playerCounts {
				totalPlayers += c
			}
			if totalPlayers == 0 {
				e.mu.Unlock()
				continue
			}

			if e.campaign.State != CampaignActive {
				e.mu.Unlock()
				continue
			}

			e.tick++

			e.processActions()

			economyTimer++
			if economyTimer >= TickRate {
				economyTimer = 0
				e.tickPopulation()
				e.tickEconomy()
			}

			e.checkVictory()
			e.expireNotifications()

			e.mu.Unlock()
		}
	}
}

func (e *Engine) processActions() {
	for _, a := range e.actions.Drain() {
		switch a.Type {
		case ActionSetSliders:
			e.handleSetSliders(a)
		case ActionSetTechAlloc:
			e.handleSetTechAlloc(a)
		}
	}
}

func (e *Engine) handleSetSliders(a PlayerAction) {
	col, ok := e.colonies[a.SystemID]
	if !ok || col.Faction != a.Faction {
		return
	}
	sum := 0
	for _, v := range a.Sliders {
		if v < 0 {
			return
		}
		sum += v
	}
	if sum != 100 {
		return
	}
	col.SliderShip = a.Sliders[0]
	col.SliderDefense = a.Sliders[1]
	col.SliderIndustry = a.Sliders[2]
	col.SliderEcology = a.Sliders[3]
	col.SliderResearch = a.Sliders[4]
}

func (e *Engine) handleSetTechAlloc(a PlayerAction) {
	if a.Faction < 0 || a.Faction >= faction.Count {
		return
	}
	sum := 0
	for _, v := range a.TechAlloc {
		if v < 0 {
			return
		}
		sum += v
	}
	if sum != 100 {
		return
	}
	for i := range e.factions[a.Faction].Tech.Allocation {
		e.factions[a.Faction].Tech.Allocation[i] = a.TechAlloc[i]
	}
}

func (e *Engine) tickPopulation() {
	for _, col := range e.colonies {
		if col.Population >= float64(col.MaxPop) {
			col.Population = float64(col.MaxPop)
			continue
		}
		growthMod := faction.Factions[col.Faction].PopGrowthMod
		growth := PopGrowthRate * col.Population * (1.0 - col.Population/float64(col.MaxPop)) * growthMod
		col.Population += growth
		if col.Population > float64(col.MaxPop) {
			col.Population = float64(col.MaxPop)
		}
	}
}

func (e *Engine) tickEconomy() {
	for sysID, col := range e.colonies {
		sys := e.galaxy.Systems[sysID]
		if sys.Planet == nil {
			continue
		}

		mineralMult := gamedata.MineralMultiplier[sys.Planet.Minerals]
		prodMod := faction.Factions[col.Faction].ProductionMod
		online := e.playerCounts[col.Faction]

		output := col.TotalOutput(mineralMult, prodMod, online)

		indOut := output * float64(col.SliderIndustry) / 100.0
		defOut := output * float64(col.SliderDefense) / 100.0
		ecoOut := output * float64(col.SliderEcology) / 100.0
		resOut := output * float64(col.SliderResearch) / 100.0

		// Industry -> factories
		col.UpdateMaxFactory(faction.Factions[col.Faction].FactoryCapMod)
		if col.Factories < col.MaxFactory && indOut > 0 {
			col.IndustryAccum += indOut
			for col.IndustryAccum >= FactoryCostBase && col.Factories < col.MaxFactory {
				col.IndustryAccum -= FactoryCostBase
				col.Factories++
			}
		}

		// Defense -> missile bases
		if defOut > 0 {
			col.DefenseAccum += defOut
			for col.DefenseAccum >= MissileBaseCost {
				col.DefenseAccum -= MissileBaseCost
				col.MissileBases++
			}
		}

		// Ecology -> clean waste
		if ecoOut > 0 {
			col.Waste -= ecoOut
			if col.Waste < 0 {
				col.Waste = 0
			}
		}

		// Waste generation
		col.Waste += col.Population * WastePerPop

		// Research -> faction RP pool
		if resOut > 0 {
			resMod := faction.Factions[col.Faction].ResearchMod
			e.factions[col.Faction].Tech.AddResearch(resOut * resMod)
		}
	}
}

func (e *Engine) checkVictory() {
	var colCount [faction.Count]int
	totalCols := 0
	for _, col := range e.colonies {
		colCount[col.Faction]++
		totalCols++
	}
	if totalCols <= 1 {
		return
	}
	for i := 0; i < faction.Count; i++ {
		if colCount[i] == totalCols {
			e.campaign.End(i)
			e.addNotification(faction.Factions[i].Name + " has conquered the galaxy!")
			return
		}
	}
}

func (e *Engine) addNotification(msg string) {
	now := time.Now()
	e.notifications = append(e.notifications, Notification{
		Message:   msg,
		CreatedAt: now,
		ExpiresAt: now.Add(6 * time.Second),
	})
}

func (e *Engine) expireNotifications() {
	now := time.Now()
	alive := e.notifications[:0]
	for _, n := range e.notifications {
		if now.Before(n.ExpiresAt) {
			alive = append(alive, n)
		}
	}
	e.notifications = alive
}

// Snapshot returns a read-only copy of the simulation state.
func (e *Engine) Snapshot() Snapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	systems := make([]SystemSnapshot, len(e.galaxy.Systems))
	for i, sys := range e.galaxy.Systems {
		ss := SystemSnapshot{
			ID:       sys.ID,
			Name:     sys.Name,
			StarType: sys.StarType,
			MapX:     sys.MapX,
			MapY:     sys.MapY,
			Special:  sys.Special,
			Owner:    -1,
		}
		if sys.Planet != nil {
			ss.HasPlanet = true
			ss.PlanetType = sys.Planet.Type
			ss.PlanetSize = sys.Planet.Size
			ss.Minerals = sys.Planet.Minerals
		}
		if col, ok := e.colonies[sys.ID]; ok {
			ss.Owner = col.Faction
		}
		systems[i] = ss
	}

	gates := make([][2]int, len(e.galaxy.Gates))
	for i, g := range e.galaxy.Gates {
		gates[i] = [2]int{g.From, g.To}
	}

	colonies := make(map[int]ColonySnapshot)
	for sysID, col := range e.colonies {
		sys := e.galaxy.Systems[sysID]
		mineralMult := 1.0
		if sys.Planet != nil {
			mineralMult = gamedata.MineralMultiplier[sys.Planet.Minerals]
		}
		prodMod := faction.Factions[col.Faction].ProductionMod
		online := e.playerCounts[col.Faction]
		totalOut := col.TotalOutput(mineralMult, prodMod, online)

		bq := make([]string, len(col.BuildQueue))
		for i, b := range col.BuildQueue {
			bq[i] = b.Name
		}

		colonies[sysID] = ColonySnapshot{
			SystemID:       col.SystemID,
			Faction:        col.Faction,
			Population:     col.Population,
			MaxPop:         col.MaxPop,
			Factories:      col.Factories,
			MaxFactory:     col.MaxFactory,
			Waste:          col.Waste,
			SliderShip:     col.SliderShip,
			SliderDefense:  col.SliderDefense,
			SliderIndustry: col.SliderIndustry,
			SliderEcology:  col.SliderEcology,
			SliderResearch: col.SliderResearch,
			MissileBases:   col.MissileBases,
			ShieldLevel:    col.ShieldLevel,
			HasStarbase:    col.HasStarbase,
			BuildQueue:     bq,
			BuildProgress:  col.BuildProgress,
			TotalOutput:    totalOut,
			ShipOutput:     totalOut * float64(col.SliderShip) / 100.0,
			DefenseOutput:  totalOut * float64(col.SliderDefense) / 100.0,
			IndustryOutput: totalOut * float64(col.SliderIndustry) / 100.0,
			EcologyOutput:  totalOut * float64(col.SliderEcology) / 100.0,
			ResearchOutput: totalOut * float64(col.SliderResearch) / 100.0,
		}
	}

	var factions [faction.Count]FactionSnapshot
	for i := 0; i < faction.Count; i++ {
		fs := e.factions[i]
		factions[i] = FactionSnapshot{
			Naquadah: fs.Naquadah,
		}
		copy(factions[i].TechTiers[:], fs.Tech.CurrentTier[:])
		copy(factions[i].TechAlloc[:], fs.Tech.Allocation[:])
		copy(factions[i].TechRP[:], fs.Tech.ResearchPts[:])

		for _, col := range e.colonies {
			if col.Faction == i {
				factions[i].SystemCount++
				factions[i].Population += col.Population
			}
		}
	}

	notifs := make([]Notification, len(e.notifications))
	copy(notifs, e.notifications)

	totalP := 0
	for _, c := range e.playerCounts {
		totalP += c
	}

	return Snapshot{
		Systems:       systems,
		Gates:         gates,
		Colonies:      colonies,
		Factions:      factions,
		Campaign:      CampaignSnapshot{State: e.campaign.State, StartedAt: e.campaign.StartedAt, Winner: e.campaign.Winner},
		Diplomacy:     DiplomacySnapshot{Relations: e.diplomacy.Relations},
		Notifications: notifs,
		PlayerCounts:  e.playerCounts,
		Tick:          e.tick,
		Paused:        totalP == 0,
	}
}

// RegisterPlayer adds a player to a faction.
func (e *Engine) RegisterPlayer(sshKey string, factionID int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, exists := e.players[sshKey]; exists {
		return false
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

// PlayerFaction returns the faction of a registered player, or -1.
func (e *Engine) PlayerFaction(sshKey string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if f, ok := e.players[sshKey]; ok {
		return f
	}
	return -1
}

// EnqueueAction adds a player action to the queue.
func (e *Engine) EnqueueAction(action PlayerAction) {
	e.actions.Enqueue(action)
}

// TotalPlayers returns total connected players.
func (e *Engine) TotalPlayers() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	total := 0
	for _, c := range e.playerCounts {
		total += c
	}
	return total
}

// GalaxySystemCount returns the number of star systems.
func (e *Engine) GalaxySystemCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.galaxy.Systems)
}
