package ratelimit

import (
	"sync"
	"time"

	"github.com/supabase/auth/internal/conf"
)

// PerIdentifierLimiter tracks rate limits for multiple identifiers (e.g., phone numbers)
// and supports multiple time windows per identifier
type PerIdentifierLimiter struct {
	mu         sync.RWMutex
	limiters   map[string]*IdentifierTracker
	config     conf.Rate
	cleanupTTL time.Duration
}

// IdentifierTracker tracks rate limiting data for a specific identifier
type IdentifierTracker struct {
	mu          sync.Mutex
	lastReset   time.Time
	count       int
	dailyReset  time.Time
	dailyCount  int
	lastAttempt time.Time
}

// NewPerIdentifierLimiter creates a new per-identifier rate limiter
func NewPerIdentifierLimiter(config conf.Rate) *PerIdentifierLimiter {
	return &PerIdentifierLimiter{
		limiters:   make(map[string]*IdentifierTracker),
		config:     config,
		cleanupTTL: 24 * time.Hour, // Clean up old trackers after 24 hours
	}
}

// Allow checks if an action is allowed for the given identifier
func (p *PerIdentifierLimiter) Allow(identifier string) bool {
	return p.AllowAt(identifier, time.Now())
}

// AllowAt checks if an action is allowed for the given identifier at the specified time
func (p *PerIdentifierLimiter) AllowAt(identifier string, at time.Time) bool {
	p.mu.Lock()
	tracker, exists := p.limiters[identifier]
	if !exists {
		tracker = &IdentifierTracker{
			lastReset:   at,
			dailyReset:  at.Truncate(24 * time.Hour),
			lastAttempt: at,
		}
		p.limiters[identifier] = tracker
	}
	p.mu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update last attempt time
	tracker.lastAttempt = at

	// Check if we need to reset the time window
	since := at.Sub(tracker.lastReset)
	if since >= p.config.OverTime {
		// Reset the counter for the current time window
		intervals := int64(since / p.config.OverTime)
		tracker.lastReset = tracker.lastReset.Add(time.Duration(intervals) * p.config.OverTime)
		tracker.count = 0
	}

	// Check if we need to reset daily counter
	if at.Truncate(24 * time.Hour).After(tracker.dailyReset) {
		tracker.dailyReset = at.Truncate(24 * time.Hour)
		tracker.dailyCount = 0
	}

	// Check current time window limit
	if tracker.count >= int(p.config.Events) {
		return false
	}

	// Allow the action and increment counter
	tracker.count++
	return true
}

// AllowDaily checks if an action is allowed within the daily limit for the given identifier
func (p *PerIdentifierLimiter) AllowDaily(identifier string, dailyLimit int) bool {
	return p.AllowDailyAt(identifier, dailyLimit, time.Now())
}

// AllowDailyAt checks if an action is allowed within the daily limit at the specified time
func (p *PerIdentifierLimiter) AllowDailyAt(identifier string, dailyLimit int, at time.Time) bool {
	p.mu.Lock()
	tracker, exists := p.limiters[identifier]
	if !exists {
		tracker = &IdentifierTracker{
			lastReset:   at,
			dailyReset:  at.Truncate(24 * time.Hour),
			lastAttempt: at,
		}
		p.limiters[identifier] = tracker
	}
	p.mu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update last attempt time
	tracker.lastAttempt = at

	// Check if we need to reset daily counter
	if at.Truncate(24 * time.Hour).After(tracker.dailyReset) {
		tracker.dailyReset = at.Truncate(24 * time.Hour)
		tracker.dailyCount = 0
	}

	// Check daily limit
	if tracker.dailyCount >= dailyLimit {
		return false
	}

	// Allow the action and increment daily counter
	tracker.dailyCount++
	return true
}

// Reset resets the counters for a specific identifier
func (p *PerIdentifierLimiter) Reset(identifier string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if tracker, exists := p.limiters[identifier]; exists {
		tracker.mu.Lock()
		tracker.count = 0
		tracker.lastReset = time.Now()
		tracker.mu.Unlock()
	}
}

// ResetAll resets all counters for a specific identifier (including daily)
func (p *PerIdentifierLimiter) ResetAll(identifier string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if tracker, exists := p.limiters[identifier]; exists {
		tracker.mu.Lock()
		tracker.count = 0
		tracker.dailyCount = 0
		tracker.lastReset = time.Now()
		tracker.dailyReset = time.Now().Truncate(24 * time.Hour)
		tracker.mu.Unlock()
	}
}

// Cleanup removes old trackers that haven't been used recently
func (p *PerIdentifierLimiter) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for identifier, tracker := range p.limiters {
		tracker.mu.Lock()
		if now.Sub(tracker.lastAttempt) > p.cleanupTTL {
			delete(p.limiters, identifier)
		}
		tracker.mu.Unlock()
	}
}

// GetCount returns the current count for an identifier (for testing/debugging)
func (p *PerIdentifierLimiter) GetCount(identifier string) int {
	p.mu.RLock()
	tracker, exists := p.limiters[identifier]
	p.mu.RUnlock()

	if !exists {
		return 0
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	return tracker.count
}

// GetDailyCount returns the current daily count for an identifier (for testing/debugging)
func (p *PerIdentifierLimiter) GetDailyCount(identifier string) int {
	p.mu.RLock()
	tracker, exists := p.limiters[identifier]
	p.mu.RUnlock()

	if !exists {
		return 0
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	return tracker.dailyCount
}
