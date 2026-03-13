package store

import "database/sql"

var migrations = []string{
	// 0: players table — persistent identity + progression
	`CREATE TABLE IF NOT EXISTS players (
		ssh_fingerprint TEXT PRIMARY KEY,
		display_name TEXT NOT NULL DEFAULT '',
		call_sign TEXT NOT NULL DEFAULT '',
		zpm_balance INTEGER NOT NULL DEFAULT 0,
		chair_level INTEGER NOT NULL DEFAULT 1,
		drone_tier INTEGER NOT NULL DEFAULT 0,
		planets_freed INTEGER NOT NULL DEFAULT 0,
		total_kills INTEGER NOT NULL DEFAULT 0,
		total_sessions INTEGER NOT NULL DEFAULT 0,
		first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 1: galaxy_planets — shared galaxy state
	`CREATE TABLE IF NOT EXISTS galaxy_planets (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		seed INTEGER NOT NULL,
		status INTEGER NOT NULL DEFAULT 0,
		invasion_level INTEGER NOT NULL DEFAULT 1,
		freed_at DATETIME
	)`,

	// 2: chat messages
	`CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		channel TEXT NOT NULL,
		sender_fp TEXT NOT NULL,
		sender_callsign TEXT NOT NULL,
		kind INTEGER NOT NULL DEFAULT 0,
		body TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_channel_time ON chat_messages(channel, created_at)`,

	// 3: teams
	`CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		leader_fp TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 4: team members
	`CREATE TABLE IF NOT EXISTS team_members (
		team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
		player_fp TEXT NOT NULL,
		joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (team_id, player_fp)
	)`,

	// 5: callsigns (unique mapping)
	`CREATE TABLE IF NOT EXISTS callsigns (
		fingerprint TEXT PRIMARY KEY,
		callsign TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 6: mutes
	`CREATE TABLE IF NOT EXISTS mutes (
		muter_fp TEXT NOT NULL,
		muted_fp TEXT NOT NULL,
		PRIMARY KEY (muter_fp, muted_fp)
	)`,

	// 7: schema version
	`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER NOT NULL
	)`,

	// 8: faction column on players (0=Ancient, 1=Ori)
	`ALTER TABLE players ADD COLUMN faction INTEGER NOT NULL DEFAULT 0`,
}

func runMigrations(db *sql.DB) error {
	for _, m := range migrations {
		_, err := db.Exec(m)
		if err != nil {
			// ALTER TABLE fails if column exists — ignore safely
			if isAlterColumnExists(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func isAlterColumnExists(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// SQLite error when column already exists
	return len(msg) > 0 && (contains(msg, "duplicate column") || contains(msg, "already exists"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
