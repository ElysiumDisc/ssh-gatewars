package server

import (
	"sync"
	"time"
)

// ConnLimiter enforces connection limits and rate limiting.
type ConnLimiter struct {
	mu              sync.Mutex
	maxSessions     int
	maxPerKey       int
	activeSessions  int
	sessionsPerKey  map[string]int

	// Token bucket for connection rate limiting
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// NewConnLimiter creates a connection limiter.
func NewConnLimiter(maxSessions, maxPerKey int, connectionsPerSec float64) *ConnLimiter {
	return &ConnLimiter{
		maxSessions:    maxSessions,
		maxPerKey:      maxPerKey,
		sessionsPerKey: make(map[string]int),
		tokens:         connectionsPerSec,
		maxTokens:      connectionsPerSec,
		refillRate:     connectionsPerSec,
		lastRefill:     time.Now(),
	}
}

// TryConnect attempts to establish a new connection. Returns true if allowed.
func (cl *ConnLimiter) TryConnect(sshKey string) bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(cl.lastRefill).Seconds()
	cl.tokens += elapsed * cl.refillRate
	if cl.tokens > cl.maxTokens {
		cl.tokens = cl.maxTokens
	}
	cl.lastRefill = now

	// Check rate limit
	if cl.tokens < 1 {
		return false
	}

	// Check total session cap
	if cl.activeSessions >= cl.maxSessions {
		return false
	}

	// Check per-key cap
	if cl.sessionsPerKey[sshKey] >= cl.maxPerKey {
		return false
	}

	cl.tokens--
	cl.activeSessions++
	cl.sessionsPerKey[sshKey]++
	return true
}

// Disconnect releases a connection slot.
func (cl *ConnLimiter) Disconnect(sshKey string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.activeSessions--
	if cl.activeSessions < 0 {
		cl.activeSessions = 0
	}
	cl.sessionsPerKey[sshKey]--
	if cl.sessionsPerKey[sshKey] <= 0 {
		delete(cl.sessionsPerKey, sshKey)
	}
}

// ActiveSessions returns the current session count.
func (cl *ConnLimiter) ActiveSessions() int {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.activeSessions
}
