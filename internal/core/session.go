package core

import "time"

// SessionInfo holds per-connection metadata for an SSH session.
type SessionInfo struct {
	SSHKey      string
	DisplayName string
	ConnectedAt time.Time
	LastInput   time.Time
	TermWidth   int
	TermHeight  int
}

// NewSessionInfo creates a SessionInfo for a new SSH connection.
func NewSessionInfo(sshKey, displayName string, width, height int) *SessionInfo {
	now := time.Now()
	return &SessionInfo{
		SSHKey:      sshKey,
		DisplayName: displayName,
		ConnectedAt: now,
		LastInput:   now,
		TermWidth:   width,
		TermHeight:  height,
	}
}
