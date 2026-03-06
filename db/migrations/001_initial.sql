-- Players table: tracks SSH key fingerprints and their faction choice
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ssh_fingerprint TEXT UNIQUE NOT NULL,
    faction INTEGER NOT NULL DEFAULT -1,
    total_sessions INTEGER NOT NULL DEFAULT 0,
    first_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Faction lifetime stats
CREATE TABLE IF NOT EXISTS faction_stats (
    faction INTEGER PRIMARY KEY,
    total_kills INTEGER NOT NULL DEFAULT 0,
    total_deaths INTEGER NOT NULL DEFAULT 0,
    total_players INTEGER NOT NULL DEFAULT 0,
    peak_territory REAL NOT NULL DEFAULT 0
);

-- Initialize faction stats rows
INSERT OR IGNORE INTO faction_stats (faction) VALUES (0), (1), (2), (3), (4);
