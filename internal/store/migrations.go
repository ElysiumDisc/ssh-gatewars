package store

import "database/sql"

var migrations = []string{
	// 0: players table
	`CREATE TABLE IF NOT EXISTS players (
		ssh_fingerprint TEXT PRIMARY KEY,
		display_name TEXT NOT NULL DEFAULT '',
		call_sign TEXT NOT NULL DEFAULT '',
		total_sessions INTEGER NOT NULL DEFAULT 0,
		first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 1: characters table
	`CREATE TABLE IF NOT EXISTS characters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ssh_fingerprint TEXT NOT NULL REFERENCES players(ssh_fingerprint),
		hp INTEGER NOT NULL DEFAULT 20,
		max_hp INTEGER NOT NULL DEFAULT 20,
		level INTEGER NOT NULL DEFAULT 1,
		xp INTEGER NOT NULL DEFAULT 0,
		missions_completed INTEGER NOT NULL DEFAULT 0,
		deaths INTEGER NOT NULL DEFAULT 0,
		location TEXT NOT NULL DEFAULT 'sgc',
		pos_x INTEGER NOT NULL DEFAULT 7,
		pos_y INTEGER NOT NULL DEFAULT 4,
		weapon_id TEXT NOT NULL DEFAULT '',
		armor_id TEXT NOT NULL DEFAULT '',
		accessory_id TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 2: inventory
	`CREATE TABLE IF NOT EXISTS inventory (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		character_id INTEGER NOT NULL REFERENCES characters(id),
		item_id TEXT NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 1
	)`,

	// 3: discovered gate addresses
	`CREATE TABLE IF NOT EXISTS discovered_addresses (
		character_id INTEGER NOT NULL REFERENCES characters(id),
		address TEXT NOT NULL,
		discovered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (character_id, address)
	)`,

	// 4: schema version
	`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER NOT NULL
	)`,

	// 5: chat messages
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

	// 6: teams
	`CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		leader_fp TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 7: team members
	`CREATE TABLE IF NOT EXISTS team_members (
		team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
		player_fp TEXT NOT NULL,
		joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (team_id, player_fp)
	)`,

	// 8: callsigns (unique mapping)
	`CREATE TABLE IF NOT EXISTS callsigns (
		fingerprint TEXT PRIMARY KEY,
		callsign TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 9: mutes
	`CREATE TABLE IF NOT EXISTS mutes (
		muter_fp TEXT NOT NULL,
		muted_fp TEXT NOT NULL,
		PRIMARY KEY (muter_fp, muted_fp)
	)`,
}

func runMigrations(db *sql.DB) error {
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return err
		}
	}
	return nil
}
