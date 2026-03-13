package store

// UpgradeGateLink sets the level of a gate link.
func (s *PlayerStore) UpgradeGateLink(fromID, toID, level int, playerFP string) error {
	_, err := s.db.Exec(`
		INSERT INTO gate_links (from_id, to_id, level, upgraded_by, upgraded_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(from_id, to_id) DO UPDATE SET
			level = excluded.level,
			upgraded_by = excluded.upgraded_by,
			upgraded_at = CURRENT_TIMESTAMP
	`, fromID, toID, level, playerFP)
	return err
}

// GetGateLinkLevels returns all persisted gate link levels.
func (s *PlayerStore) GetGateLinkLevels() (map[[2]int]int, error) {
	rows, err := s.db.Query("SELECT from_id, to_id, level FROM gate_links WHERE level > 0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[[2]int]int)
	for rows.Next() {
		var fromID, toID, level int
		if err := rows.Scan(&fromID, &toID, &level); err != nil {
			return nil, err
		}
		result[[2]int{fromID, toID}] = level
	}
	return result, rows.Err()
}

// RecordTransfer logs a resource transfer.
func (s *PlayerStore) RecordTransfer(senderFP string, planetID, bonusType, amount int) error {
	_, err := s.db.Exec(`
		INSERT INTO resource_transfers (sender_fp, target_planet_id, bonus_type, amount)
		VALUES (?, ?, ?, ?)
	`, senderFP, planetID, bonusType, amount)
	return err
}
