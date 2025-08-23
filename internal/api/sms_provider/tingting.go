package sms_provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/supabase/auth/internal/conf"
	"github.com/supabase/auth/internal/utilities"
)

type TingtingProvider struct {
	Config  *conf.TingtingProviderConfiguration
	APIPath string
}

type TingtingRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
	Sender  string `json:"sender"`
}

type TingtingResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    struct {
		MessageID string `json:"message_id,omitempty"`
	} `json:"data,omitempty"`
}

// Creates a SmsProvider with the Tingting Config
func NewTingtingProvider(config conf.TingtingProviderConfiguration) (SmsProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	apiPath := config.BaseURL + "/api/sms"
	return &TingtingProvider{
		Config:  &config,
		APIPath: apiPath,
	}, nil
}

func (t *TingtingProvider) SendMessage(phone, message, channel, otp string) (string, error) {
	switch channel {
	case SMSProvider:
		return t.SendSms(phone, message)
	default:
		return "", fmt.Errorf("channel type %q is not supported for Tingting", channel)
	}
}

// Send an SMS containing the OTP with Tingting's API
func (t *TingtingProvider) SendSms(phone string, message string) (string, error) {
	payload := TingtingRequest{
		To:      phone,
		Content: message,
		Sender:  t.Config.Sender,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: defaultTimeout}
	req, err := http.NewRequest("POST", t.APIPath, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", t.Config.ApiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer utilities.SafeClose(resp.Body)

	var tingtingResp TingtingResponse
	if err := json.NewDecoder(resp.Body).Decode(&tingtingResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if tingtingResp.Status != "success" {
		return "", fmt.Errorf("tingting API error: %s", tingtingResp.Message)
	}

	return tingtingResp.Data.MessageID, nil
}

// VerifyOTP is not implemented for Tingting as it's a simple SMS provider
func (t *TingtingProvider) VerifyOTP(phone, token string) error {
	return fmt.Errorf("VerifyOTP is not supported for Tingting provider")
}
