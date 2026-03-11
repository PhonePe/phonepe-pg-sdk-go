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

import "time"

// RetryConfig defines the retry behavior for API calls
type RetryConfig struct {
	// Enabled determines if retry logic is active
	Enabled bool

	// MaxAttempts is the maximum number of retry attempts (including the initial attempt)
	MaxAttempts int

	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases with each retry (exponential backoff)
	Multiplier float64

	// RetryableStatusCodes are HTTP status codes that should trigger a retry
	RetryableStatusCodes []int
}

// DefaultRetryConfig returns a sensible default retry configuration
// - 3 total attempts (1 initial + 2 retries)
// - Exponential backoff starting at 100ms
// - Retries on 5xx errors and 429 (rate limiting)
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		Enabled:              true,
		MaxAttempts:          3,
		InitialDelay:         100 * time.Millisecond,
		MaxDelay:             5 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// NoRetryConfig returns a configuration with retry disabled
func NoRetryConfig() *RetryConfig {
	return &RetryConfig{
		Enabled: false,
	}
}

// ConservativeRetryConfig returns a conservative retry configuration
// - 2 total attempts (1 initial + 1 retry)
// - Shorter delays
// - Only retries on 503 (service unavailable)
func ConservativeRetryConfig() *RetryConfig {
	return &RetryConfig{
		Enabled:              true,
		MaxAttempts:          2,
		InitialDelay:         50 * time.Millisecond,
		MaxDelay:             1 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{503},
	}
}

// AggressiveRetryConfig returns an aggressive retry configuration
// - 5 total attempts (1 initial + 4 retries)
// - Longer maximum delay
// - Retries on more error codes
func AggressiveRetryConfig() *RetryConfig {
	return &RetryConfig{
		Enabled:              true,
		MaxAttempts:          5,
		InitialDelay:         200 * time.Millisecond,
		MaxDelay:             10 * time.Second,
		Multiplier:           2.0,
		RetryableStatusCodes: []int{408, 429, 500, 502, 503, 504},
	}
}
