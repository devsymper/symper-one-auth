package sms_provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supabase/auth/internal/conf"
)

func TestNewTingtingProvider(t *testing.T) {
	config := conf.TingtingProviderConfiguration{
		ApiKey:    "test-api-key",
		ApiSender: "TING TING",
	}

	provider, err := NewTingtingProvider(config)
	require.NoError(t, err)
	assert.NotNil(t, provider)

	tingtingProvider, ok := provider.(*TingtingProvider)
	require.True(t, ok)
	assert.Equal(t, "https://v1.tingting.im/api/sms", tingtingProvider.APIPath)
	assert.Equal(t, "test-api-key", tingtingProvider.Config.ApiKey)
	assert.Equal(t, "TING TING", tingtingProvider.Config.ApiSender)
}

func TestTingtingProviderValidation(t *testing.T) {
	// Test missing API key
	config := conf.TingtingProviderConfiguration{
		ApiSender: "TING TING",
	}
	_, err := NewTingtingProvider(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing Tingting API key")

	// Test missing sender
	config = conf.TingtingProviderConfiguration{
		ApiKey: "test-api-key",
	}
	_, err = NewTingtingProvider(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing Tingting sender")
}

func TestTingtingProviderSendMessage(t *testing.T) {
	config := conf.TingtingProviderConfiguration{
		ApiKey:    "test-api-key",
		ApiSender: "TING TING",
	}

	provider, err := NewTingtingProvider(config)
	require.NoError(t, err)

	// Test unsupported channel
	_, err = provider.SendMessage("+1234567890", "Test message", "whatsapp", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "channel type \"whatsapp\" is not supported for Tingting")

	// Test VerifyOTP not supported
	err = provider.VerifyOTP("+1234567890", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VerifyOTP is not supported for Tingting provider")
}
