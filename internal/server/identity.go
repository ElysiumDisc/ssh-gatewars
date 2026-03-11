package server

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/charmbracelet/ssh"
)

// SessionKey extracts a stable identifier from an SSH session.
// Uses the public key fingerprint if available, otherwise falls back
// to a session-based anonymous ID.
func SessionKey(s ssh.Session) string {
	if s.PublicKey() != nil {
		hash := sha256.Sum256(s.PublicKey().Marshal())
		return "SHA256:" + base64.StdEncoding.EncodeToString(hash[:])
	}
	// Anonymous fallback: use remote address + session ID.
	return fmt.Sprintf("anon:%s:%s", s.RemoteAddr().String(), s.Context().SessionID())
}

// DisplayName extracts a display name from the SSH session.
// Uses the SSH username, falling back to "guest".
func DisplayName(s ssh.Session) string {
	if u := s.User(); u != "" {
		return u
	}
	return "guest"
}
