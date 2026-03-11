package store

import "database/sql"

// Team represents a team row.
type Team struct {
	ID       int64
	Name     string
	LeaderFP string
}

// CreateTeam creates a new SG team. Returns the team ID.
func (s *PlayerStore) CreateTeam(name, leaderFP string) (int64, error) {
	res, err := s.db.Exec(
		"INSERT INTO teams (name, leader_fp) VALUES (?, ?)", name, leaderFP)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()

	// Add leader as member
	s.db.Exec("INSERT INTO team_members (team_id, player_fp) VALUES (?, ?)", id, leaderFP)
	return id, nil
}

// DisbandTeam removes a team and all its members.
func (s *PlayerStore) DisbandTeam(teamID int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec("DELETE FROM team_members WHERE team_id = ?", teamID)
	tx.Exec("DELETE FROM teams WHERE id = ?", teamID)
	return tx.Commit()
}

// AddTeamMember adds a player to a team.
func (s *PlayerStore) AddTeamMember(teamID int64, playerFP string) error {
	_, err := s.db.Exec(
		"INSERT OR IGNORE INTO team_members (team_id, player_fp) VALUES (?, ?)",
		teamID, playerFP)
	return err
}

// RemoveTeamMember removes a player from a team.
func (s *PlayerStore) RemoveTeamMember(teamID int64, playerFP string) error {
	_, err := s.db.Exec(
		"DELETE FROM team_members WHERE team_id = ? AND player_fp = ?",
		teamID, playerFP)
	return err
}

// GetTeamByPlayer returns the team a player belongs to, or nil.
func (s *PlayerStore) GetTeamByPlayer(playerFP string) (*Team, error) {
	row := s.db.QueryRow(`
		SELECT t.id, t.name, t.leader_fp
		FROM teams t
		JOIN team_members tm ON tm.team_id = t.id
		WHERE tm.player_fp = ?
	`, playerFP)
	var t Team
	err := row.Scan(&t.ID, &t.Name, &t.LeaderFP)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetTeamMembers returns fingerprints of all members of a team.
func (s *PlayerStore) GetTeamMembers(teamID int64) ([]string, error) {
	rows, err := s.db.Query(
		"SELECT player_fp FROM team_members WHERE team_id = ?", teamID)
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

// GetTeamByName looks up a team by name.
func (s *PlayerStore) GetTeamByName(name string) (*Team, error) {
	row := s.db.QueryRow("SELECT id, name, leader_fp FROM teams WHERE name = ?", name)
	var t Team
	err := row.Scan(&t.ID, &t.Name, &t.LeaderFP)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}
