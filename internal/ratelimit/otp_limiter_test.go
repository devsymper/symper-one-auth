package ratelimit

import (
	"testing"
	"time"
)

func TestOTPSendLimiter(t *testing.T) {
	// Test interval rate limiting (1 minute between sends)
	limiter := NewOTPSendLimiter(1*time.Minute, 5)
	phoneNumber := "+1234567890"

	// First send should be allowed
	if !limiter.CanSend(phoneNumber) {
		t.Error("First send should be allowed")
	}

	// Record the send
	limiter.RecordSend(phoneNumber)

	// Immediate second send should be blocked
	if limiter.CanSend(phoneNumber) {
		t.Error("Second send within interval should be blocked")
	}

	// Send after 1 minute should be allowed
	if !limiter.CanSendAt(phoneNumber, time.Now().Add(1*time.Minute+1*time.Second)) {
		t.Error("Send after interval should be allowed")
	}
}

func TestOTPSendLimiterDailyLimit(t *testing.T) {
	// Test daily limit (5 sends per day)
	limiter := NewOTPSendLimiter(1*time.Second, 3) // Low daily limit for testing
	phoneNumber := "+1234567890"
	now := time.Now()

	// Send 3 times (daily limit)
	for i := 0; i < 3; i++ {
		if !limiter.CanSendAt(phoneNumber, now.Add(time.Duration(i+1)*time.Minute)) {
			t.Errorf("Send %d should be allowed", i+1)
		}
		limiter.RecordSendAt(phoneNumber, now.Add(time.Duration(i+1)*time.Minute))
	}

	// 4th send should be blocked
	if limiter.CanSendAt(phoneNumber, now.Add(4*time.Minute)) {
		t.Error("4th send should be blocked due to daily limit")
	}

	// Send next day should be allowed
	nextDay := now.Add(24 * time.Hour)
	if !limiter.CanSendAt(phoneNumber, nextDay) {
		t.Error("Send next day should be allowed")
	}
}

func TestOTPVerifyFailureLimiter(t *testing.T) {
	// Test verification failure rate limiting (3 failures per 5 minutes)
	limiter := NewOTPVerifyFailureLimiter(3, 5*time.Minute)
	phoneNumber := "+1234567890"
	now := time.Now()

	// First 3 attempts should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.CanAttemptAt(phoneNumber, now.Add(time.Duration(i)*time.Second)) {
			t.Errorf("Attempt %d should be allowed", i+1)
		}
		limiter.RecordFailureAt(phoneNumber, now.Add(time.Duration(i)*time.Second))
	}

	// 4th attempt should be blocked
	if limiter.CanAttemptAt(phoneNumber, now.Add(3*time.Second)) {
		t.Error("4th attempt should be blocked")
	}

	// Attempt after interval should be allowed
	if !limiter.CanAttemptAt(phoneNumber, now.Add(5*time.Minute+1*time.Second)) {
		t.Error("Attempt after interval should be allowed")
	}
}

func TestOTPVerifyFailureLimiterReset(t *testing.T) {
	// Test reset functionality
	limiter := NewOTPVerifyFailureLimiter(3, 5*time.Minute)
	phoneNumber := "+1234567890"
	now := time.Now()

	// Record 3 failures
	for i := 0; i < 3; i++ {
		limiter.RecordFailureAt(phoneNumber, now.Add(time.Duration(i)*time.Second))
	}

	// Should be blocked
	if limiter.CanAttemptAt(phoneNumber, now.Add(3*time.Second)) {
		t.Error("Should be blocked after 3 failures")
	}

	// Reset the counter
	limiter.Reset(phoneNumber)

	// Should be allowed again
	if !limiter.CanAttemptAt(phoneNumber, now.Add(3*time.Second)) {
		t.Error("Should be allowed after reset")
	}
}

func TestMultiplePhoneNumbers(t *testing.T) {
	// Test that rate limiting is per phone number
	limiter := NewOTPSendLimiter(1*time.Minute, 5)
	phone1 := "+1234567890"
	phone2 := "+0987654321"

	// Send to phone1
	limiter.RecordSend(phone1)

	// phone1 should be blocked
	if limiter.CanSend(phone1) {
		t.Error("phone1 should be blocked")
	}

	// phone2 should still be allowed
	if !limiter.CanSend(phone2) {
		t.Error("phone2 should be allowed")
	}
}
