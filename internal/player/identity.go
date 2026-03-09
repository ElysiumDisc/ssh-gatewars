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
    ssh_fingerprint TEXT PRIMARY KEY,
    faction         INTEGER NOT NULL DEFAULT -1,
    display_name    TEXT NOT NULL DEFAULT '',
    total_sessions  INTEGER NOT NULL DEFAULT 0,
    first_seen      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

// Store handles player persistence via SQLite.
type Store struct {
	db *sql.DB
}

// NewStore opens (or creates) the SQLite database.
func NewStore(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}

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

// GetPlayer looks up a player by SSH fingerprint.
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
