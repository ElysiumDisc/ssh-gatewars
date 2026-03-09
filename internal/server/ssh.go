package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/charmbracelet/ssh"

	"ssh-gatewars/internal/simulation"
)

// Server holds shared resources for all SSH sessions.
type Server struct {
	Engine *simulation.Engine
}

// SessionKey extracts a stable identifier from the SSH session.
func SessionKey(s ssh.Session) string {
	pk := s.PublicKey()
	if pk != nil {
		h := sha256.Sum256(pk.Marshal())
		return "SHA256:" + hex.EncodeToString(h[:16])
	}
	return fmt.Sprintf("anon:%s", s.Context().SessionID())
}
