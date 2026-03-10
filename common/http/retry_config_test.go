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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	assert.NotNil(t, config, "DefaultRetryConfig should not return nil")
	assert.True(t, config.Enabled, "Default config should be enabled")
	assert.Equal(t, 3, config.MaxAttempts, "Default should have 3 max attempts")
	assert.Equal(t, 100*time.Millisecond, config.InitialDelay, "Default initial delay should be 100ms")
	assert.Equal(t, 5*time.Second, config.MaxDelay, "Default max delay should be 5s")
	assert.Equal(t, 2.0, config.Multiplier, "Default multiplier should be 2.0")
	assert.ElementsMatch(t, []int{429, 500, 502, 503, 504}, config.RetryableStatusCodes,
		"Default should retry on 429, 500, 502, 503, 504")
}

func TestNoRetryConfig(t *testing.T) {
	config := NoRetryConfig()

	assert.NotNil(t, config, "NoRetryConfig should not return nil")
	assert.False(t, config.Enabled, "NoRetry config should be disabled")
}

func TestConservativeRetryConfig(t *testing.T) {
	config := ConservativeRetryConfig()

	assert.NotNil(t, config, "ConservativeRetryConfig should not return nil")
	assert.True(t, config.Enabled, "Conservative config should be enabled")
	assert.Equal(t, 2, config.MaxAttempts, "Conservative should have 2 max attempts")
	assert.Equal(t, 50*time.Millisecond, config.InitialDelay, "Conservative initial delay should be 50ms")
	assert.Equal(t, 1*time.Second, config.MaxDelay, "Conservative max delay should be 1s")
	assert.Equal(t, 2.0, config.Multiplier, "Conservative multiplier should be 2.0")
	assert.ElementsMatch(t, []int{503}, config.RetryableStatusCodes,
		"Conservative should only retry on 503")
}

func TestAggressiveRetryConfig(t *testing.T) {
	config := AggressiveRetryConfig()

	assert.NotNil(t, config, "AggressiveRetryConfig should not return nil")
	assert.True(t, config.Enabled, "Aggressive config should be enabled")
	assert.Equal(t, 5, config.MaxAttempts, "Aggressive should have 5 max attempts")
	assert.Equal(t, 200*time.Millisecond, config.InitialDelay, "Aggressive initial delay should be 200ms")
	assert.Equal(t, 10*time.Second, config.MaxDelay, "Aggressive max delay should be 10s")
	assert.Equal(t, 2.0, config.Multiplier, "Aggressive multiplier should be 2.0")
	assert.ElementsMatch(t, []int{408, 429, 500, 502, 503, 504}, config.RetryableStatusCodes,
		"Aggressive should retry on 408, 429, 500, 502, 503, 504")
}

func TestCustomRetryConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *RetryConfig
	}{
		{
			name: "custom with 4 attempts",
			config: &RetryConfig{
				Enabled:              true,
				MaxAttempts:          4,
				InitialDelay:         250 * time.Millisecond,
				MaxDelay:             30 * time.Second,
				Multiplier:           3.0,
				RetryableStatusCodes: []int{429, 503},
			},
		},
		{
			name: "custom disabled",
			config: &RetryConfig{
				Enabled:     false,
				MaxAttempts: 1,
			},
		},
		{
			name: "custom with single retry code",
			config: &RetryConfig{
				Enabled:              true,
				MaxAttempts:          2,
				InitialDelay:         100 * time.Millisecond,
				MaxDelay:             1 * time.Second,
				Multiplier:           2.0,
				RetryableStatusCodes: []int{500},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config, "Custom config should not be nil")
			if tt.config.Enabled {
				assert.Greater(t, tt.config.MaxAttempts, 0, "MaxAttempts should be positive when enabled")
				assert.NotEmpty(t, tt.config.RetryableStatusCodes, "RetryableStatusCodes should not be empty when enabled")
			}
		})
	}
}

func TestRetryConfigComparison(t *testing.T) {
	tests := []struct {
		name     string
		config1  *RetryConfig
		config2  *RetryConfig
		expected string
	}{
		{
			name:     "Default vs Conservative - attempts",
			config1:  DefaultRetryConfig(),
			config2:  ConservativeRetryConfig(),
			expected: "default has more attempts",
		},
		{
			name:     "Default vs Aggressive - attempts",
			config1:  DefaultRetryConfig(),
			config2:  AggressiveRetryConfig(),
			expected: "aggressive has more attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expected == "default has more attempts" {
				assert.Greater(t, tt.config1.MaxAttempts, tt.config2.MaxAttempts)
			} else if tt.expected == "aggressive has more attempts" {
				assert.Greater(t, tt.config2.MaxAttempts, tt.config1.MaxAttempts)
			}
		})
	}
}
