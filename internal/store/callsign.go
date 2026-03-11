package store

import "database/sql"

// SetCallsign sets or updates a player's callsign.
func (s *PlayerStore) SetCallsign(fingerprint, callsign string) error {
	_, err := s.db.Exec(`
		INSERT INTO callsigns (fingerprint, callsign, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(fingerprint) DO UPDATE SET
			callsign = excluded.callsign,
			updated_at = CURRENT_TIMESTAMP
	`, fingerprint, callsign)
	return err
}

// GetCallsign returns a player's callsign, or "" if not set.
func (s *PlayerStore) GetCallsign(fingerprint string) (string, error) {
	row := s.db.QueryRow("SELECT callsign FROM callsigns WHERE fingerprint = ?", fingerprint)
	var cs string
	err := row.Scan(&cs)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return cs, err
}

// IsCallsignTaken checks if a callsign is in use (case-insensitive).
func (s *PlayerStore) IsCallsignTaken(callsign string) (bool, error) {
	row := s.db.QueryRow(
		"SELECT COUNT(*) FROM callsigns WHERE LOWER(callsign) = LOWER(?)", callsign)
	var count int
	err := row.Scan(&count)
	return count > 0, err
}

// LookupFingerprint finds the fingerprint for a callsign (case-insensitive).
func (s *PlayerStore) LookupFingerprint(callsign string) (string, error) {
	row := s.db.QueryRow(
		"SELECT fingerprint FROM callsigns WHERE LOWER(callsign) = LOWER(?)", callsign)
	var fp string
	err := row.Scan(&fp)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return fp, err
}

// AddMute records that muterFP muted mutedFP.
func (s *PlayerStore) AddMute(muterFP, mutedFP string) error {
	_, err := s.db.Exec(
		"INSERT OR IGNORE INTO mutes (muter_fp, muted_fp) VALUES (?, ?)",
		muterFP, mutedFP)
	return err
}

// RemoveMute removes a mute.
func (s *PlayerStore) RemoveMute(muterFP, mutedFP string) error {
	_, err := s.db.Exec(
		"DELETE FROM mutes WHERE muter_fp = ? AND muted_fp = ?",
		muterFP, mutedFP)
	return err
}

// GetMutes returns all fingerprints muted by the given player.
func (s *PlayerStore) GetMutes(muterFP string) ([]string, error) {
	rows, err := s.db.Query("SELECT muted_fp FROM mutes WHERE muter_fp = ?", muterFP)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fps []string
	for rows.Next() {
		var fp string
		if rows.Scan(&fp) == nil {
			fps = append(fps, fp)
		}
	}
	return fps, nil
}
