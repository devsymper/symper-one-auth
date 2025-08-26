package api

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/api/apierrors"
	"github.com/supabase/auth/internal/ratelimit"
)

func TestOTPErrorCodeSeparation(t *testing.T) {
	// Test validateOtpWithDetailedErrors function with different scenarios
	api := &API{
		limiterOpts: &LimiterOptions{
			OtpVerifyFailLimiter: ratelimit.NewOTPVerifyFailureLimiter(3, 5*time.Minute),
		},
	}

	identifier := "test@example.com"
	tokenHash := "valid_token_hash"
	wrongTokenHash := "wrong_token_hash"
	expectedToken := "valid_token_hash"
	validSentAt := time.Now().Add(-30 * time.Second) // 30 seconds ago
	expiredSentAt := time.Now().Add(-2 * time.Hour)  // 2 hours ago
	otpExp := uint(3600)                             // 1 hour expiry

	t.Run("Valid OTP", func(t *testing.T) {
		err := api.validateOtpWithDetailedErrors(tokenHash, expectedToken, &validSentAt, otpExp, identifier)
		assert.NoError(t, err, "Valid OTP should not return error")
	})

	t.Run("OTP Code Mismatch", func(t *testing.T) {
		err := api.validateOtpWithDetailedErrors(wrongTokenHash, expectedToken, &validSentAt, otpExp, identifier)
		require.Error(t, err, "Wrong OTP code should return error")

		var apiErr *apierrors.HTTPError
		require.True(t, errors.As(err, &apiErr), "Should be an HTTP error")
		assert.Equal(t, apierrors.ErrorCodeOTPCodeMismatch, apiErr.ErrorCode, "Should return OTP code mismatch error")
		assert.Contains(t, apiErr.Message, "Invalid OTP code", "Should contain appropriate message")
	})

	t.Run("OTP Expired", func(t *testing.T) {
		err := api.validateOtpWithDetailedErrors(tokenHash, expectedToken, &expiredSentAt, otpExp, identifier)
		require.Error(t, err, "Expired OTP should return error")

		var apiErr *apierrors.HTTPError
		require.True(t, errors.As(err, &apiErr), "Should be an HTTP error")
		assert.Equal(t, apierrors.ErrorCodeOTPExpired, apiErr.ErrorCode, "Should return OTP expired error")
		assert.Contains(t, apiErr.Message, "OTP has expired", "Should contain appropriate message")
	})

	t.Run("Missing Token", func(t *testing.T) {
		err := api.validateOtpWithDetailedErrors(tokenHash, "", &validSentAt, otpExp, identifier)
		require.Error(t, err, "Missing expected token should return error")

		var apiErr *apierrors.HTTPError
		require.True(t, errors.As(err, &apiErr), "Should be an HTTP error")
		assert.Equal(t, apierrors.ErrorCodeOTPExpired, apiErr.ErrorCode, "Should return generic OTP expired error for missing token")
	})

	t.Run("Missing SentAt", func(t *testing.T) {
		err := api.validateOtpWithDetailedErrors(tokenHash, expectedToken, nil, otpExp, identifier)
		require.Error(t, err, "Missing sentAt should return error")

		var apiErr *apierrors.HTTPError
		require.True(t, errors.As(err, &apiErr), "Should be an HTTP error")
		assert.Equal(t, apierrors.ErrorCodeOTPExpired, apiErr.ErrorCode, "Should return generic OTP expired error for missing sentAt")
	})
}

func TestOTPSendRateLimitErrorCodes(t *testing.T) {
	api := &API{
		limiterOpts: &LimiterOptions{
			OtpSendLimiter: ratelimit.NewOTPSendLimiter(1*time.Minute, 3), // 1 minute interval, 3 per day
		},
	}

	phoneNumber := "+1234567890"

	t.Run("First Send Allowed", func(t *testing.T) {
		canSend := api.limiterOpts.OtpSendLimiter.CanSend(phoneNumber)
		assert.True(t, canSend, "First send should be allowed")
	})

	t.Run("Interval Rate Limit", func(t *testing.T) {
		// Record a send
		api.limiterOpts.OtpSendLimiter.RecordSend(phoneNumber)

		// Try to send again immediately - should be blocked by interval
		canSend := api.limiterOpts.OtpSendLimiter.CanSend(phoneNumber)
		assert.False(t, canSend, "Immediate second send should be blocked by interval")

		// In the actual API, this would trigger ErrorCodeOTPSendIntervalRateLimit
		// since waitTime > 0
	})

	t.Run("Daily Rate Limit", func(t *testing.T) {
		// Create a new limiter with low daily limit for testing
		limiter := ratelimit.NewOTPSendLimiter(1*time.Second, 1) // 1 second interval, 1 per day

		// Record a send
		limiter.RecordSend(phoneNumber)

		// Wait for interval to pass
		time.Sleep(2 * time.Second)

		// Try to send again - should be blocked by daily limit
		canSend := limiter.CanSend(phoneNumber)
		assert.False(t, canSend, "Should be blocked by daily limit")

		// In the actual API, this would trigger ErrorCodeOTPSendDailyRateLimit
		// since waitTime <= 0 but daily limit reached
	})
}

func TestErrorCodeConstants(t *testing.T) {
	// Verify all new error codes are defined
	assert.Equal(t, "otp_code_mismatch", apierrors.ErrorCodeOTPCodeMismatch, "OTP code mismatch error code should be defined")
	assert.Equal(t, "otp_send_interval_rate_limit", apierrors.ErrorCodeOTPSendIntervalRateLimit, "OTP send interval rate limit error code should be defined")
	assert.Equal(t, "otp_send_daily_rate_limit", apierrors.ErrorCodeOTPSendDailyRateLimit, "OTP send daily rate limit error code should be defined")
	assert.Equal(t, "otp_expired", apierrors.ErrorCodeOTPExpired, "OTP expired error code should be defined")
}
