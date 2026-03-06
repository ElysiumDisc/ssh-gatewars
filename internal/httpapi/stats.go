package httpapi

import (
	"encoding/json"
	"net/http"

	"ssh-gatewars/internal/faction"
	"ssh-gatewars/internal/simulation"
)

// StatsServer serves a read-only JSON stats endpoint.
type StatsServer struct {
	engine *simulation.Engine
}

// NewStatsServer creates a stats HTTP server.
func NewStatsServer(engine *simulation.Engine) *StatsServer {
	return &StatsServer{engine: engine}
}

// Handler returns an http.Handler for the stats API.
func (s *StatsServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", s.handleStats)
	return mux
}

type factionStats struct {
	Name      string  `json:"name"`
	Players   int     `json:"players"`
	Ships     int     `json:"ships"`
	Territory float64 `json:"territory"`
	Kills     int     `json:"kills"`
	Deaths    int     `json:"deaths"`
}

type statsResponse struct {
	Factions []factionStats `json:"factions"`
	Tick     uint64         `json:"tick"`
}

func (s *StatsServer) handleStats(w http.ResponseWriter, r *http.Request) {
	snap := s.engine.Snapshot()

	resp := statsResponse{
		Factions: make([]factionStats, faction.Count),
		Tick:     snap.Tick,
	}

	for i := 0; i < faction.Count; i++ {
		territory := 20.0
		if snap.Territory != nil {
			territory = snap.Territory.Percents[i]
		}
		resp.Factions[i] = factionStats{
			Name:      faction.Factions[i].Name,
			Players:   snap.PlayerCounts[i],
			Ships:     snap.ShipCounts[i],
			Territory: territory,
			Kills:     snap.KillCounts[i],
			Deaths:    snap.DeathCounts[i],
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
}
