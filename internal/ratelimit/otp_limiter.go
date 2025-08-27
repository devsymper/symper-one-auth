package ratelimit

import (
	"sync"
	"time"
)

// OTPSendLimiter manages rate limiting for OTP sends per phone number
// It enforces both interval-based (e.g., 1 minute between sends) and daily limits
type OTPSendLimiter struct {
	mu         sync.RWMutex
	trackers   map[string]*OTPSendTracker
	interval   time.Duration
	dailyLimit int
	cleanupTTL time.Duration
}

// OTPSendTracker tracks OTP send attempts for a specific phone number
type OTPSendTracker struct {
	mu          sync.Mutex
	lastSend    time.Time
	dailyReset  time.Time
	dailyCount  int
	lastAttempt time.Time
}

// NewOTPSendLimiter creates a new OTP send rate limiter
func NewOTPSendLimiter(interval time.Duration, dailyLimit int) *OTPSendLimiter {
	return &OTPSendLimiter{
		trackers:   make(map[string]*OTPSendTracker),
		interval:   interval,
		dailyLimit: dailyLimit,
		cleanupTTL: 48 * time.Hour, // Clean up after 2 days
	}
}

// CanSend checks if an OTP can be sent to the given phone number
func (o *OTPSendLimiter) CanSend(phoneNumber string) bool {
	return o.CanSendAt(phoneNumber, time.Now())
}

// CanSendAt checks if an OTP can be sent to the given phone number at the specified time
func (o *OTPSendLimiter) CanSendAt(phoneNumber string, at time.Time) bool {
	o.mu.Lock()
	tracker, exists := o.trackers[phoneNumber]
	if !exists {
		tracker = &OTPSendTracker{
			dailyReset:  at.Truncate(24 * time.Hour),
			lastAttempt: at,
		}
		o.trackers[phoneNumber] = tracker
	}
	o.mu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update last attempt time
	tracker.lastAttempt = at

	// Check if we need to reset daily counter
	if at.Truncate(24 * time.Hour).After(tracker.dailyReset) {
		tracker.dailyReset = at.Truncate(24 * time.Hour)
		tracker.dailyCount = 0
	}

	// Check daily limit first
	if tracker.dailyCount >= o.dailyLimit {
		return false
	}

	// Check interval limit (time since last send)
	if !tracker.lastSend.IsZero() && at.Sub(tracker.lastSend) < o.interval {
		return false
	}

	return true
}

// RecordSend records a successful OTP send for the given phone number
func (o *OTPSendLimiter) RecordSend(phoneNumber string) {
	o.RecordSendAt(phoneNumber, time.Now())
}

// RecordSendAt records a successful OTP send at the specified time
func (o *OTPSendLimiter) RecordSendAt(phoneNumber string, at time.Time) {
	o.mu.Lock()
	tracker, exists := o.trackers[phoneNumber]
	if !exists {
		tracker = &OTPSendTracker{
			dailyReset:  at.Truncate(24 * time.Hour),
			lastAttempt: at,
		}
		o.trackers[phoneNumber] = tracker
	}
	o.mu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update last attempt time
	tracker.lastAttempt = at

	// Check if we need to reset daily counter
	if at.Truncate(24 * time.Hour).After(tracker.dailyReset) {
		tracker.dailyReset = at.Truncate(24 * time.Hour)
		tracker.dailyCount = 0
	}

	// Record the send
	tracker.lastSend = at
	tracker.dailyCount++
}

// GetNextAllowedTime returns when the next OTP can be sent for the given phone number
func (o *OTPSendLimiter) GetNextAllowedTime(phoneNumber string) time.Time {
	o.mu.RLock()
	tracker, exists := o.trackers[phoneNumber]
	o.mu.RUnlock()

	if !exists {
		return time.Now() // Can send immediately if no record exists
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	if tracker.lastSend.IsZero() {
		return time.Now() // Can send immediately if never sent before
	}

	nextAllowed := tracker.lastSend.Add(o.interval)
	if nextAllowed.Before(time.Now()) {
		return time.Now() // Can send now
	}

	return nextAllowed
}

// Cleanup removes old trackers that haven't been used recently
func (o *OTPSendLimiter) Cleanup() {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now()
	for phoneNumber, tracker := range o.trackers {
		tracker.mu.Lock()
		if now.Sub(tracker.lastAttempt) > o.cleanupTTL {
			delete(o.trackers, phoneNumber)
		}
		tracker.mu.Unlock()
	}
}

// OTPVerifyFailureLimiter manages rate limiting for OTP verification failures per phone number
type OTPVerifyFailureLimiter struct {
	mu         sync.RWMutex
	trackers   map[string]*OTPVerifyTracker
	interval   time.Duration
	limit      int
	cleanupTTL time.Duration
}

// OTPVerifyTracker tracks OTP verification failures for a specific phone number
type OTPVerifyTracker struct {
	mu          sync.Mutex
	lastReset   time.Time
	count       int
	lastAttempt time.Time
}

// NewOTPVerifyFailureLimiter creates a new OTP verify failure rate limiter
func NewOTPVerifyFailureLimiter(limit int, interval time.Duration) *OTPVerifyFailureLimiter {
	return &OTPVerifyFailureLimiter{
		trackers:   make(map[string]*OTPVerifyTracker),
		interval:   interval,
		limit:      limit,
		cleanupTTL: 24 * time.Hour,
	}
}

// CanAttempt checks if an OTP verification attempt is allowed for the given phone number
func (o *OTPVerifyFailureLimiter) CanAttempt(phoneNumber string) bool {
	return o.CanAttemptAt(phoneNumber, time.Now())
}

// CanAttemptAt checks if an OTP verification attempt is allowed at the specified time
func (o *OTPVerifyFailureLimiter) CanAttemptAt(phoneNumber string, at time.Time) bool {
	o.mu.Lock()
	tracker, exists := o.trackers[phoneNumber]
	if !exists {
		tracker = &OTPVerifyTracker{
			lastReset:   at,
			lastAttempt: at,
		}
		o.trackers[phoneNumber] = tracker
	}
	o.mu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update last attempt time
	tracker.lastAttempt = at

	// Check if we need to reset the time window
	if at.Sub(tracker.lastReset) >= o.interval {
		// Reset the counter for the new time window
		intervals := int64(at.Sub(tracker.lastReset) / o.interval)
		tracker.lastReset = tracker.lastReset.Add(time.Duration(intervals) * o.interval)
		tracker.count = 0
	}

	// Check if we've exceeded the limit
	return tracker.count < o.limit
}

// RecordFailure records a failed OTP verification attempt
func (o *OTPVerifyFailureLimiter) RecordFailure(phoneNumber string) {
	o.RecordFailureAt(phoneNumber, time.Now())
}

// RecordFailureAt records a failed OTP verification attempt at the specified time
func (o *OTPVerifyFailureLimiter) RecordFailureAt(phoneNumber string, at time.Time) {
	o.mu.Lock()
	tracker, exists := o.trackers[phoneNumber]
	if !exists {
		tracker = &OTPVerifyTracker{
			lastReset:   at,
			lastAttempt: at,
		}
		o.trackers[phoneNumber] = tracker
	}
	o.mu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Update last attempt time
	tracker.lastAttempt = at

	// Check if we need to reset the time window
	if at.Sub(tracker.lastReset) >= o.interval {
		// Reset the counter for the new time window
		intervals := int64(at.Sub(tracker.lastReset) / o.interval)
		tracker.lastReset = tracker.lastReset.Add(time.Duration(intervals) * o.interval)
		tracker.count = 0
	}

	// Record the failure
	tracker.count++
}

// Reset resets the failure counter for a specific phone number (called when OTP resend succeeds)
func (o *OTPVerifyFailureLimiter) Reset(phoneNumber string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if tracker, exists := o.trackers[phoneNumber]; exists {
		tracker.mu.Lock()
		tracker.count = 0
		tracker.lastReset = time.Now()
		tracker.mu.Unlock()
	}
}

// GetRemainingAttempts returns the number of remaining verification attempts for the given phone number
func (o *OTPVerifyFailureLimiter) GetRemainingAttempts(phoneNumber string) int {
	return o.GetRemainingAttemptsAt(phoneNumber, time.Now())
}

// GetRemainingAttemptsAt returns the number of remaining verification attempts at the specified time
func (o *OTPVerifyFailureLimiter) GetRemainingAttemptsAt(phoneNumber string, at time.Time) int {
	o.mu.RLock()
	tracker, exists := o.trackers[phoneNumber]
	o.mu.RUnlock()

	if !exists {
		return o.limit // Full attempts available if no record exists
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	// Check if we need to reset the time window
	if at.Sub(tracker.lastReset) >= o.interval {
		// Would reset the counter for the new time window
		return o.limit
	}

	// Return remaining attempts in current window
	remaining := o.limit - tracker.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Cleanup removes old trackers that haven't been used recently
func (o *OTPVerifyFailureLimiter) Cleanup() {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now()
	for phoneNumber, tracker := range o.trackers {
		tracker.mu.Lock()
		if now.Sub(tracker.lastAttempt) > o.cleanupTTL {
			delete(o.trackers, phoneNumber)
		}
		tracker.mu.Unlock()
	}
}
