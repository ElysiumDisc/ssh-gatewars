package player

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const migrationSQL = `
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ssh_fingerprint TEXT UNIQUE NOT NULL,
    faction INTEGER NOT NULL DEFAULT -1,
    total_sessions INTEGER NOT NULL DEFAULT 0,
    first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS faction_stats (
    faction INTEGER PRIMARY KEY,
    total_kills INTEGER NOT NULL DEFAULT 0,
    total_deaths INTEGER NOT NULL DEFAULT 0,
    total_players INTEGER NOT NULL DEFAULT 0,
    peak_territory REAL NOT NULL DEFAULT 0
);

INSERT OR IGNORE INTO faction_stats (faction) VALUES (0), (1), (2), (3), (4);
`

// Store handles player persistence via SQLite.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) the SQLite database at the given path.
func NewStore(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// WAL mode for concurrent reads
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

	// Run migration
	if _, err := db.Exec(migrationSQL); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// PlayerInfo holds saved player data.
type PlayerInfo struct {
	Faction       int
	TotalSessions int
}

// GetPlayer looks up a player by SSH fingerprint. Returns nil if not found.
func (s *Store) GetPlayer(fingerprint string) (*PlayerInfo, error) {
	row := s.db.QueryRow(
		"SELECT faction, total_sessions FROM players WHERE ssh_fingerprint = ?",
		fingerprint,
	)
	var info PlayerInfo
	err := row.Scan(&info.Faction, &info.TotalSessions)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// SavePlayer creates or updates a player record.
func (s *Store) SavePlayer(fingerprint string, factionID int) error {
	_, err := s.db.Exec(`
		INSERT INTO players (ssh_fingerprint, faction, total_sessions, last_seen)
		VALUES (?, ?, 1, ?)
		ON CONFLICT(ssh_fingerprint) DO UPDATE SET
			faction = excluded.faction,
			total_sessions = total_sessions + 1,
			last_seen = excluded.last_seen
	`, fingerprint, factionID, time.Now().UTC())
	return err
}

// UpdateFactionStats records kills/deaths for a faction.
func (s *Store) UpdateFactionStats(factionID int, kills, deaths int) error {
	_, err := s.db.Exec(`
		UPDATE faction_stats SET
			total_kills = total_kills + ?,
			total_deaths = total_deaths + ?
		WHERE faction = ?
	`, kills, deaths, factionID)
	return err
}
