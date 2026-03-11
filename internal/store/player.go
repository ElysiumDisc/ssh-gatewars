package store

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"ssh-gatewars/internal/core"
	"ssh-gatewars/internal/entity"
	"ssh-gatewars/internal/gamedata"
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

// PlayerRecord holds basic player info for the SSH server.
type PlayerRecord struct {
	DisplayName string
	CallSign    string
}

// GetPlayer loads a player record by SSH fingerprint.
func (s *PlayerStore) GetPlayer(fingerprint string) (*PlayerRecord, error) {
	row := s.db.QueryRow(
		"SELECT display_name, call_sign FROM players WHERE ssh_fingerprint = ?",
		fingerprint,
	)
	var rec PlayerRecord
	err := row.Scan(&rec.DisplayName, &rec.CallSign)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// UpsertPlayer creates or updates a player record.
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

// LoadCharacter loads a character for a player. Returns nil if no character exists.
func (s *PlayerStore) LoadCharacter(fingerprint string, cfg core.GameConfig) (*entity.Character, error) {
	row := s.db.QueryRow(`
		SELECT c.id, p.display_name, p.call_sign,
		       c.hp, c.max_hp, c.level, c.xp,
		       c.missions_completed, c.deaths,
		       c.location, c.pos_x, c.pos_y,
		       c.weapon_id, c.armor_id, c.accessory_id
		FROM characters c
		JOIN players p ON p.ssh_fingerprint = c.ssh_fingerprint
		WHERE c.ssh_fingerprint = ?
		ORDER BY c.id DESC LIMIT 1
	`, fingerprint)

	var (
		id          int64
		displayName string
		callSign    string
		hp, maxHP   int
		level, xp   int
		missions    int
		deaths      int
		location    string
		posX, posY  int
		weaponID    string
		armorID     string
		accID       string
	)
	err := row.Scan(&id, &displayName, &callSign,
		&hp, &maxHP, &level, &xp,
		&missions, &deaths,
		&location, &posX, &posY,
		&weaponID, &armorID, &accID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	c := &entity.Character{
		ID:                id,
		SSHKey:            fingerprint,
		DisplayName:       displayName,
		CallSign:          callSign,
		HP:                hp,
		MaxHP:             maxHP,
		Level:             level,
		XP:                xp,
		MissionsCompleted: missions,
		Deaths:            deaths,
		Location:          location,
		Pos:               core.Pos{X: posX, Y: posY},
		MaxItems:          20,
		Inventory:         make([]entity.Item, 0),
	}

	// Load equipment
	if weaponID != "" {
		c.Weapon = &entity.Item{DefID: weaponID, Quantity: 1}
	}
	if armorID != "" {
		c.Armor = &entity.Item{DefID: armorID, Quantity: 1}
	}
	if accID != "" {
		c.Accessory = &entity.Item{DefID: accID, Quantity: 1}
	}

	// Load inventory
	rows, err := s.db.Query(
		"SELECT item_id, quantity FROM inventory WHERE character_id = ?", id)
	if err != nil {
		return c, nil // return char without inventory on error
	}
	defer rows.Close()
	for rows.Next() {
		var itemID string
		var qty int
		if err := rows.Scan(&itemID, &qty); err == nil {
			c.Inventory = append(c.Inventory, entity.Item{DefID: itemID, Quantity: qty})
		}
	}

	// Load discovered addresses
	addrRows, err := s.db.Query(
		"SELECT address FROM discovered_addresses WHERE character_id = ?", id)
	if err != nil {
		return c, nil
	}
	defer addrRows.Close()
	for addrRows.Next() {
		var addrStr string
		if err := addrRows.Scan(&addrStr); err == nil {
			if addr, ok := gamedata.ParseAddress(addrStr); ok {
				c.DiscoveredAddresses = append(c.DiscoveredAddresses, addr)
			}
		}
	}

	return c, nil
}

// SaveCharacter persists a character's current state.
func (s *PlayerStore) SaveCharacter(c *entity.Character) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	weaponID := ""
	if c.Weapon != nil {
		weaponID = c.Weapon.DefID
	}
	armorID := ""
	if c.Armor != nil {
		armorID = c.Armor.DefID
	}
	accID := ""
	if c.Accessory != nil {
		accID = c.Accessory.DefID
	}

	if c.ID == 0 {
		// Insert new character
		res, err := tx.Exec(`
			INSERT INTO characters (ssh_fingerprint, hp, max_hp, level, xp,
			    missions_completed, deaths, location, pos_x, pos_y,
			    weapon_id, armor_id, accessory_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, c.SSHKey, c.HP, c.MaxHP, c.Level, c.XP,
			c.MissionsCompleted, c.Deaths, c.Location, c.Pos.X, c.Pos.Y,
			weaponID, armorID, accID)
		if err != nil {
			return err
		}
		c.ID, _ = res.LastInsertId()
	} else {
		// Update existing character
		_, err := tx.Exec(`
			UPDATE characters SET hp=?, max_hp=?, level=?, xp=?,
			    missions_completed=?, deaths=?, location=?, pos_x=?, pos_y=?,
			    weapon_id=?, armor_id=?, accessory_id=?
			WHERE id=?
		`, c.HP, c.MaxHP, c.Level, c.XP,
			c.MissionsCompleted, c.Deaths, c.Location, c.Pos.X, c.Pos.Y,
			weaponID, armorID, accID, c.ID)
		if err != nil {
			return err
		}
	}

	// Replace inventory
	tx.Exec("DELETE FROM inventory WHERE character_id = ?", c.ID)
	for _, item := range c.Inventory {
		tx.Exec("INSERT INTO inventory (character_id, item_id, quantity) VALUES (?, ?, ?)",
			c.ID, item.DefID, item.Quantity)
	}

	// Replace discovered addresses
	tx.Exec("DELETE FROM discovered_addresses WHERE character_id = ?", c.ID)
	for _, addr := range c.DiscoveredAddresses {
		tx.Exec("INSERT INTO discovered_addresses (character_id, address) VALUES (?, ?)",
			c.ID, addr.Code())
	}

	return tx.Commit()
}
