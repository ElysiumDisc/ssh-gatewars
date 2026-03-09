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
	Name       string  `json:"name"`
	Players    int     `json:"players"`
	Systems    int     `json:"systems"`
	Population float64 `json:"population"`
	Naquadah   float64 `json:"naquadah"`
}

type campaignStats struct {
	State  string `json:"state"`
	Winner string `json:"winner"`
}

type statsResponse struct {
	Factions []factionStats `json:"factions"`
	Systems  int            `json:"systems"`
	Colonies int            `json:"colonies"`
	Campaign campaignStats  `json:"campaign"`
	Tick     uint64         `json:"tick"`
}

func (s *StatsServer) handleStats(w http.ResponseWriter, r *http.Request) {
	snap := s.engine.Snapshot()

	resp := statsResponse{
		Factions: make([]factionStats, faction.Count),
		Systems:  len(snap.Systems),
		Colonies: len(snap.Colonies),
		Tick:     snap.Tick,
	}

	for i := 0; i < faction.Count; i++ {
		resp.Factions[i] = factionStats{
			Name:       faction.Factions[i].Name,
			Players:    snap.PlayerCounts[i],
			Systems:    snap.Factions[i].SystemCount,
			Population: snap.Factions[i].Population,
			Naquadah:   snap.Factions[i].Naquadah,
		}
	}

	state := "active"
	winner := ""
	if snap.Campaign.State == simulation.CampaignWon {
		state = "won"
		if snap.Campaign.Winner >= 0 && snap.Campaign.Winner < faction.Count {
			winner = faction.Factions[snap.Campaign.Winner].Name
		}
	}
	resp.Campaign = campaignStats{State: state, Winner: winner}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
}
