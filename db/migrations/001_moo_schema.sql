-- SSH GateWars v1.0 — Master of Orion Clone Schema
-- Fresh schema for 4X strategy game

CREATE TABLE IF NOT EXISTS players (
    ssh_fingerprint TEXT PRIMARY KEY,
    faction         INTEGER NOT NULL DEFAULT -1,
    display_name    TEXT NOT NULL DEFAULT '',
    total_sessions  INTEGER NOT NULL DEFAULT 0,
    first_seen      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS campaigns (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    seed          INTEGER NOT NULL,
    system_count  INTEGER NOT NULL DEFAULT 50,
    state         TEXT NOT NULL DEFAULT 'active',
    winner_faction INTEGER NOT NULL DEFAULT -1,
    started_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at      DATETIME,
    tick          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS campaign_galaxy (
    campaign_id INTEGER PRIMARY KEY REFERENCES campaigns(id),
    galaxy_json TEXT NOT NULL,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS colonies (
    campaign_id    INTEGER NOT NULL REFERENCES campaigns(id),
    system_id      INTEGER NOT NULL,
    faction        INTEGER NOT NULL,
    population     REAL NOT NULL DEFAULT 1.0,
    factories      INTEGER NOT NULL DEFAULT 0,
    waste          REAL NOT NULL DEFAULT 0,
    slider_ship    INTEGER NOT NULL DEFAULT 0,
    slider_defense INTEGER NOT NULL DEFAULT 0,
    slider_industry INTEGER NOT NULL DEFAULT 30,
    slider_ecology INTEGER NOT NULL DEFAULT 20,
    slider_research INTEGER NOT NULL DEFAULT 50,
    missile_bases  INTEGER NOT NULL DEFAULT 0,
    shield_level   INTEGER NOT NULL DEFAULT 0,
    build_queue_json TEXT NOT NULL DEFAULT '[]',
    build_progress REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (campaign_id, system_id)
);

CREATE TABLE IF NOT EXISTS faction_tech (
    campaign_id    INTEGER NOT NULL REFERENCES campaigns(id),
    faction        INTEGER NOT NULL,
    allocation_json TEXT NOT NULL DEFAULT '[17,17,17,17,16,16]',
    tiers_json     TEXT NOT NULL DEFAULT '[0,0,0,0,0,0]',
    rp_json        TEXT NOT NULL DEFAULT '[0,0,0,0,0,0]',
    discovered_json TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (campaign_id, faction)
);

CREATE TABLE IF NOT EXISTS ship_designs (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id    INTEGER NOT NULL REFERENCES campaigns(id),
    faction        INTEGER NOT NULL,
    name           TEXT NOT NULL,
    hull           INTEGER NOT NULL,
    components_json TEXT NOT NULL,
    cost           REAL NOT NULL,
    active         INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS fleets (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id    INTEGER NOT NULL REFERENCES campaigns(id),
    faction        INTEGER NOT NULL,
    system_id      INTEGER NOT NULL,
    state          TEXT NOT NULL DEFAULT 'idle',
    ships_json     TEXT NOT NULL DEFAULT '{}',
    from_system    INTEGER NOT NULL DEFAULT -1,
    to_system      INTEGER NOT NULL DEFAULT -1,
    progress       REAL NOT NULL DEFAULT 0,
    travel_mode    TEXT NOT NULL DEFAULT 'gate'
);

CREATE TABLE IF NOT EXISTS diplomacy (
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id),
    faction_a   INTEGER NOT NULL,
    faction_b   INTEGER NOT NULL,
    status      TEXT NOT NULL DEFAULT 'none',
    trade_income REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (campaign_id, faction_a, faction_b)
);

CREATE TABLE IF NOT EXISTS player_contributions (
    campaign_id     INTEGER NOT NULL REFERENCES campaigns(id),
    ssh_fingerprint TEXT NOT NULL,
    faction         INTEGER NOT NULL,
    ships_built     INTEGER NOT NULL DEFAULT 0,
    colonies_founded INTEGER NOT NULL DEFAULT 0,
    battles_won     INTEGER NOT NULL DEFAULT 0,
    research_contributed REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (campaign_id, ssh_fingerprint)
);
