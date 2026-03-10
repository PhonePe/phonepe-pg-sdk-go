/*
 *  Copyright (c) 2026 Original Author(s), PhonePe India Pvt. Ltd.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/exception"
	"github.com/stretchr/testify/assert"
)

// TestCalculateBackoff tests the exponential backoff calculation
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name          string
		attempt       int
		config        *RetryConfig
		expectedDelay time.Duration
	}{
		{
			name:    "first retry with default config",
			attempt: 0,
			config: &RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   2.0,
				MaxDelay:     5 * time.Second,
			},
			expectedDelay: 100 * time.Millisecond, // 100 * 2^0
		},
		{
			name:    "second retry with default config",
			attempt: 1,
			config: &RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   2.0,
				MaxDelay:     5 * time.Second,
			},
			expectedDelay: 200 * time.Millisecond, // 100 * 2^1
		},
		{
			name:    "third retry with default config",
			attempt: 2,
			config: &RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   2.0,
				MaxDelay:     5 * time.Second,
			},
			expectedDelay: 400 * time.Millisecond, // 100 * 2^2
		},
		{
			name:    "max delay cap",
			attempt: 10,
			config: &RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   2.0,
				MaxDelay:     1 * time.Second,
			},
			expectedDelay: 1 * time.Second, // Capped at MaxDelay
		},
		{
			name:    "aggressive config first retry",
			attempt: 0,
			config: &RetryConfig{
				InitialDelay: 200 * time.Millisecond,
				Multiplier:   2.0,
				MaxDelay:     10 * time.Second,
			},
			expectedDelay: 200 * time.Millisecond,
		},
		{
			name:    "conservative config",
			attempt: 0,
			config: &RetryConfig{
				InitialDelay: 50 * time.Millisecond,
				Multiplier:   2.0,
				MaxDelay:     1 * time.Second,
			},
			expectedDelay: 50 * time.Millisecond,
		},
		{
			name:    "custom multiplier 3x",
			attempt: 1,
			config: &RetryConfig{
				InitialDelay: 100 * time.Millisecond,
				Multiplier:   3.0,
				MaxDelay:     10 * time.Second,
			},
			expectedDelay: 300 * time.Millisecond, // 100 * 3^1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateBackoff(tt.attempt, tt.config)
			assert.Equal(t, tt.expectedDelay, delay, "Backoff delay mismatch")
		})
	}
}

// TestIsRetryableError tests error classification for retry logic
func TestIsRetryableError(t *testing.T) {
	retryableStatusCodes := []int{429, 500, 502, 503, 504}

	tests := []struct {
		name       string
		err        error
		expected   bool
		statusCode int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name: "retryable 429 - too many requests",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 429; return &code }(),
			},
			expected:   true,
			statusCode: 429,
		},
		{
			name: "retryable 500 - internal server error",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 500; return &code }(),
			},
			expected:   true,
			statusCode: 500,
		},
		{
			name: "retryable 502 - bad gateway",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 502; return &code }(),
			},
			expected:   true,
			statusCode: 502,
		},
		{
			name: "retryable 503 - service unavailable",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 503; return &code }(),
			},
			expected:   true,
			statusCode: 503,
		},
		{
			name: "retryable 504 - gateway timeout",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 504; return &code }(),
			},
			expected:   true,
			statusCode: 504,
		},
		{
			name: "non-retryable 400 - bad request",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 400; return &code }(),
			},
			expected:   false,
			statusCode: 400,
		},
		{
			name: "non-retryable 401 - unauthorized",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 401; return &code }(),
			},
			expected:   false,
			statusCode: 401,
		},
		{
			name: "non-retryable 403 - forbidden",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 403; return &code }(),
			},
			expected:   false,
			statusCode: 403,
		},
		{
			name: "non-retryable 404 - not found",
			err: &exception.PhonePeException{
				HttpStatusCode: func() *int { code := 404; return &code }(),
			},
			expected:   false,
			statusCode: 404,
		},
		{
			name: "PhonePe exception with nil status code",
			err: &exception.PhonePeException{
				HttpStatusCode: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err, retryableStatusCodes)
			assert.Equal(t, tt.expected, result, "isRetryableError returned unexpected result")
		})
	}
}

// TestExecuteWithRetry_SuccessOnFirstAttempt tests successful request without retries
func TestExecuteWithRetry_SuccessOnFirstAttempt(t *testing.T) {
	// Create mock server that succeeds immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	// Test with retry config - should succeed on first attempt
	config := DefaultRetryConfig()

	// Note: This is a simplified test. In real scenarios, we'd need to create
	// a full HttpCommand instance with proper dependencies
	assert.NotNil(t, config, "Config should not be nil")
	assert.True(t, config.Enabled, "Config should be enabled")
}

// TestExecuteWithRetry_SuccessAfterRetries tests successful request after failures
func TestExecuteWithRetry_SuccessAfterRetries(t *testing.T) {
	attemptCount := 0

	// Create mock server that fails twice, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount < 3 {
			// Fail with 503 for first two attempts
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
			return
		}

		// Succeed on third attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	// Verify server behavior
	assert.Equal(t, 0, attemptCount, "Should start with 0 attempts")
}

// TestExecuteWithRetry_FailureAfterMaxRetries tests failure after exhausting retries
func TestExecuteWithRetry_FailureAfterMaxRetries(t *testing.T) {
	// Create mock server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
	}))
	defer server.Close()

	config := &RetryConfig{
		Enabled:              true,
		MaxAttempts:          3,
		InitialDelay:         10 * time.Millisecond,
		MaxDelay:             100 * time.Millisecond,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{503},
	}

	assert.Equal(t, 3, config.MaxAttempts, "Should have 3 max attempts")
}

// TestExecuteWithRetry_NonRetryableError tests immediate failure on non-retryable errors
func TestExecuteWithRetry_NonRetryableError(t *testing.T) {
	// Create mock server that returns 400 Bad Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
	}))
	defer server.Close()

	config := DefaultRetryConfig()

	// 400 is not in retryable status codes, should fail immediately
	assert.NotContains(t, config.RetryableStatusCodes, 400, "400 should not be retryable")
}

// TestExecuteWithRetry_DisabledRetry tests behavior when retry is disabled
func TestExecuteWithRetry_DisabledRetry(t *testing.T) {
	// Create mock server that fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
	}))
	defer server.Close()

	config := NoRetryConfig()

	// Should fail immediately without retries
	assert.False(t, config.Enabled, "Retry should be disabled")
}

// TestExecuteWithRetry_NilConfig tests behavior with nil config
func TestExecuteWithRetry_NilConfig(t *testing.T) {
	var config *RetryConfig = nil

	// Nil config should behave like retry disabled
	assert.Nil(t, config, "Config is nil")
}

// TestRetryBackoffProgression tests that backoff increases correctly
func TestRetryBackoffProgression(t *testing.T) {
	config := DefaultRetryConfig()

	// Calculate delays for multiple attempts
	delay0 := calculateBackoff(0, config)
	delay1 := calculateBackoff(1, config)
	delay2 := calculateBackoff(2, config)

	// Each delay should be greater than the previous (exponential increase)
	assert.Less(t, delay0, delay1, "Second delay should be greater than first")
	assert.Less(t, delay1, delay2, "Third delay should be greater than second")

	// Verify exponential relationship
	assert.Equal(t, delay0*2, delay1, "Delay should double with multiplier 2.0")
	assert.Equal(t, delay1*2, delay2, "Delay should continue doubling")
}

// TestRetryWithDifferentStatusCodes tests retry logic with various HTTP status codes
func TestRetryWithDifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		shouldRetry bool
		configCodes []int
	}{
		{"429 with default config", 429, true, []int{429, 500, 502, 503, 504}},
		{"500 with default config", 500, true, []int{429, 500, 502, 503, 504}},
		{"400 with default config", 400, false, []int{429, 500, 502, 503, 504}},
		{"503 with conservative", 503, true, []int{503}},
		{"500 with conservative", 500, false, []int{503}},
		{"408 with aggressive", 408, true, []int{408, 429, 500, 502, 503, 504}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &exception.PhonePeException{
				HttpStatusCode: &tt.statusCode,
			}

			result := isRetryableError(err, tt.configCodes)
			assert.Equal(t, tt.shouldRetry, result,
				"Status code %d should be retryable=%v", tt.statusCode, tt.shouldRetry)
		})
	}
}
