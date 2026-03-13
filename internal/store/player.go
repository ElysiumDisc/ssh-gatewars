package store

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// PlayerStore handles all SQLite persistence.
type PlayerStore struct {
	db *sql.DB
}

// NewPlayerStore opens the SQLite database and runs migrations.
func NewPlayerStore(path string) (*PlayerStore, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}
	return &PlayerStore{db: db}, nil
}

// Close closes the database.
func (s *PlayerStore) Close() error {
	return s.db.Close()
}

// PlayerRecord holds player progression data.
type PlayerRecord struct {
	SSHFingerprint string
	DisplayName    string
	CallSign       string
	ZPMBalance     int
	ChairLevel     int
	DroneTier      int
	Faction        int // 0=Ancient, 1=Ori
	PlanetsFreed   int
	TotalKills     int
	TotalSessions  int
}

// GetPlayer loads a player record by SSH fingerprint.
func (s *PlayerStore) GetPlayer(fingerprint string) (*PlayerRecord, error) {
	row := s.db.QueryRow(`
		SELECT ssh_fingerprint, display_name, call_sign,
		       zpm_balance, chair_level, drone_tier, faction,
		       planets_freed, total_kills, total_sessions
		FROM players WHERE ssh_fingerprint = ?
	`, fingerprint)
	var rec PlayerRecord
	err := row.Scan(
		&rec.SSHFingerprint, &rec.DisplayName, &rec.CallSign,
		&rec.ZPMBalance, &rec.ChairLevel, &rec.DroneTier, &rec.Faction,
		&rec.PlanetsFreed, &rec.TotalKills, &rec.TotalSessions,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// UpsertPlayer creates or updates a player record (session bump).
func (s *PlayerStore) UpsertPlayer(fingerprint, displayName, callSign string) error {
	_, err := s.db.Exec(`
		INSERT INTO players (ssh_fingerprint, display_name, call_sign, total_sessions, first_seen, last_seen)
		VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(ssh_fingerprint) DO UPDATE SET
			display_name = excluded.display_name,
			call_sign = CASE WHEN excluded.call_sign != '' THEN excluded.call_sign ELSE players.call_sign END,
			total_sessions = players.total_sessions + 1,
			last_seen = CURRENT_TIMESTAMP
	`, fingerprint, displayName, callSign)
	return err
}

// AddZPM adds ZPM to a player's balance.
func (s *PlayerStore) AddZPM(fingerprint string, amount int) error {
	_, err := s.db.Exec(
		"UPDATE players SET zpm_balance = zpm_balance + ? WHERE ssh_fingerprint = ?",
		amount, fingerprint)
	return err
}

// SpendZPM deducts ZPM. Returns false if insufficient.
func (s *PlayerStore) SpendZPM(fingerprint string, amount int) (bool, error) {
	res, err := s.db.Exec(
		"UPDATE players SET zpm_balance = zpm_balance - ? WHERE ssh_fingerprint = ? AND zpm_balance >= ?",
		amount, fingerprint, amount)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// UpgradeChair increments chair level.
func (s *PlayerStore) UpgradeChair(fingerprint string) error {
	_, err := s.db.Exec(
		"UPDATE players SET chair_level = chair_level + 1 WHERE ssh_fingerprint = ?",
		fingerprint)
	return err
}

// UpgradeDroneTier sets the drone tier.
func (s *PlayerStore) UpgradeDroneTier(fingerprint string, tier int) error {
	_, err := s.db.Exec(
		"UPDATE players SET drone_tier = ? WHERE ssh_fingerprint = ?",
		tier, fingerprint)
	return err
}

// RecordPlanetFreed increments planets_freed counter.
func (s *PlayerStore) RecordPlanetFreed(fingerprint string) error {
	_, err := s.db.Exec(
		"UPDATE players SET planets_freed = planets_freed + 1 WHERE ssh_fingerprint = ?",
		fingerprint)
	return err
}

// AddKills adds to total kill count.
func (s *PlayerStore) AddKills(fingerprint string, count int) error {
	_, err := s.db.Exec(
		"UPDATE players SET total_kills = total_kills + ? WHERE ssh_fingerprint = ?",
		count, fingerprint)
	return err
}

// SetFaction changes a player's faction.
func (s *PlayerStore) SetFaction(fingerprint string, faction int) error {
	_, err := s.db.Exec(
		"UPDATE players SET faction = ? WHERE ssh_fingerprint = ?",
		faction, fingerprint)
	return err
}

// ResetPlayer resets a player's progression (ZPM, chair level, drone tier) but keeps callsign and stats.
func (s *PlayerStore) ResetPlayer(fingerprint string) error {
	_, err := s.db.Exec(`
		UPDATE players SET
			zpm_balance = 0,
			chair_level = 1,
			drone_tier = 0,
			faction = 0
		WHERE ssh_fingerprint = ?
	`, fingerprint)
	return err
}

// SaveGalaxyPlanet persists a planet's state.
func (s *PlayerStore) SaveGalaxyPlanet(id int, name string, seed int64, status int, invasionLevel int) error {
	_, err := s.db.Exec(`
		INSERT INTO galaxy_planets (id, name, seed, status, invasion_level)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			invasion_level = excluded.invasion_level,
			freed_at = CASE WHEN excluded.status = 2 THEN CURRENT_TIMESTAMP ELSE galaxy_planets.freed_at END
	`, id, name, seed, status, invasionLevel)
	return err
}
