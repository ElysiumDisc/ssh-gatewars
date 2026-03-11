package server

import (
	"sync"
	"time"
)

// ConnLimiter enforces connection rate limits and per-key session caps.
type ConnLimiter struct {
	maxSessions int
	maxPerKey   int
	rateLimit   float64 // max new connections per second

	mu           sync.Mutex
	sessions     map[string]int // sshKey → active session count
	totalActive  int
	lastConnect  time.Time
	tokenBucket  float64
}

// NewConnLimiter creates a new connection limiter.
func NewConnLimiter(maxSessions, maxPerKey int, rateLimit float64) *ConnLimiter {
	return &ConnLimiter{
		maxSessions: maxSessions,
		maxPerKey:   maxPerKey,
		rateLimit:   rateLimit,
		sessions:    make(map[string]int),
		lastConnect: time.Now(),
		tokenBucket: rateLimit,
	}
}

// TryConnect attempts to register a new session. Returns true if allowed.
func (l *ConnLimiter) TryConnect(sshKey string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Refill token bucket.
	now := time.Now()
	elapsed := now.Sub(l.lastConnect).Seconds()
	l.lastConnect = now
	l.tokenBucket += elapsed * l.rateLimit
	if l.tokenBucket > l.rateLimit*2 {
		l.tokenBucket = l.rateLimit * 2
	}

	// Check rate limit.
	if l.tokenBucket < 1.0 {
		return false
	}

	// Check total session cap.
	if l.totalActive >= l.maxSessions {
		return false
	}

	// Check per-key cap.
	if l.sessions[sshKey] >= l.maxPerKey {
		return false
	}

	l.tokenBucket -= 1.0
	l.sessions[sshKey]++
	l.totalActive++
	return true
}

// Disconnect removes a session from tracking.
func (l *ConnLimiter) Disconnect(sshKey string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.sessions[sshKey] > 0 {
		l.sessions[sshKey]--
		if l.sessions[sshKey] == 0 {
			delete(l.sessions, sshKey)
		}
	}
	if l.totalActive > 0 {
		l.totalActive--
	}
}

// ActiveCount returns the total number of active sessions.
func (l *ConnLimiter) ActiveCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.totalActive
}
