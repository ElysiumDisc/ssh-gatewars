package simulation

import (
	"time"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/gamedata"
)

// Notification is a broadcast message.
type Notification struct {
	Message   string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Snapshot is a read-only copy of the simulation state.
type Snapshot struct {
	Systems       []SystemSnapshot
	Gates         [][2]int
	Colonies      map[int]ColonySnapshot
	Fleets        []FleetSnapshot
	Factions      [faction.Count]FactionSnapshot
	Campaign      CampaignSnapshot
	Diplomacy     DiplomacySnapshot
	Notifications []Notification
	PlayerCounts  [faction.Count]int
	Tick          uint64
	Paused        bool
}

// SystemSnapshot is a read-only view of a star system.
type SystemSnapshot struct {
	ID         int
	Name       string
	StarType   int
	MapX, MapY float64
	HasPlanet  bool
	PlanetType int
	PlanetSize int
	Minerals   int
	Special    int
	Owner      int
}

// ColonySnapshot is a read-only view of a colony.
type ColonySnapshot struct {
	SystemID       int
	Faction        int
	Population     float64
	MaxPop         int
	Factories      int
	MaxFactory     int
	Waste          float64
	SliderShip     int
	SliderDefense  int
	SliderIndustry int
	SliderEcology  int
	SliderResearch int
	MissileBases   int
	ShieldLevel    int
	HasStarbase    bool
	BuildQueue     []string
	BuildProgress  float64
	TotalOutput    float64
	ShipOutput     float64
	DefenseOutput  float64
	IndustryOutput float64
	EcologyOutput  float64
	ResearchOutput float64
}

// FleetSnapshot is a read-only view of a fleet.
type FleetSnapshot struct {
	ID         uint64
	Faction    int
	SystemID   int
	Ships      map[int]int
	State      int
	FromSystem int
	ToSystem   int
	Progress   float64
}

// FactionSnapshot is a read-only view of faction state.
type FactionSnapshot struct {
	Naquadah    float64
	TechTiers   [gamedata.TreeCount]int
	TechAlloc   [gamedata.TreeCount]int
	TechRP      [gamedata.TreeCount]float64
	SystemCount int
	Population  float64
	TotalProd   float64
}

// CampaignSnapshot is a read-only view of campaign state.
type CampaignSnapshot struct {
	State     int
	StartedAt time.Time
	Winner    int
}

// DiplomacySnapshot is a read-only view of diplomacy state.
type DiplomacySnapshot struct {
	Relations [faction.Count][faction.Count]int
}
