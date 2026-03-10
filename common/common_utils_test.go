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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCallbackValid(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		authorization string
		expected      bool
	}{
		{
			name:          "valid authorization",
			username:      "testuser",
			password:      "testpass",
			authorization: CalculateSha256("testuser", "testpass"),
			expected:      true,
		},
		{
			name:          "invalid authorization - wrong password",
			username:      "testuser",
			password:      "testpass",
			authorization: CalculateSha256("testuser", "wrongpass"),
			expected:      false,
		},
		{
			name:          "invalid authorization - wrong username",
			username:      "testuser",
			password:      "testpass",
			authorization: CalculateSha256("wronguser", "testpass"),
			expected:      false,
		},
		{
			name:          "invalid authorization - empty string",
			username:      "testuser",
			password:      "testpass",
			authorization: "",
			expected:      false,
		},
		{
			name:          "invalid authorization - malformed hash",
			username:      "testuser",
			password:      "testpass",
			authorization: "invalid_hash_value",
			expected:      false,
		},
		{
			name:          "empty username and password",
			username:      "",
			password:      "",
			authorization: CalculateSha256("", ""),
			expected:      true,
		},
		{
			name:          "special characters in credentials",
			username:      "user@test.com",
			password:      "p@$$w0rd!#%",
			authorization: CalculateSha256("user@test.com", "p@$$w0rd!#%"),
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCallbackValid(tt.username, tt.password, tt.authorization)
			assert.Equal(t, tt.expected, result, "IsCallbackValid returned unexpected result")
		})
	}
}

func TestCalculateSha256(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string // We'll calculate this dynamically
	}{
		{
			name: "single string",
			args: []interface{}{"test"},
		},
		{
			name: "multiple strings",
			args: []interface{}{"user", "pass"},
		},
		{
			name: "mixed types",
			args: []interface{}{"client", 123, true},
		},
		{
			name: "empty args",
			args: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateSha256(tt.args...)
			assert.NotEmpty(t, result, "CalculateSha256 should return non-empty hash")
			assert.Len(t, result, 64, "SHA256 hash should be 64 characters (hex)")

			// Test consistency - same input should give same output
			result2 := CalculateSha256(tt.args...)
			assert.Equal(t, result, result2, "CalculateSha256 should be deterministic")
		})
	}
}

func TestCalculateSha256Consistency(t *testing.T) {
	// Test that SHA256 is consistent
	hash1 := CalculateSha256("user", "pass")
	hash2 := CalculateSha256("user", "pass")

	assert.Equal(t, hash1, hash2, "Same inputs should produce same hash")

	// Different inputs should produce different hash
	hash3 := CalculateSha256("user", "different")
	assert.NotEqual(t, hash1, hash3, "Different inputs should produce different hashes")
}
