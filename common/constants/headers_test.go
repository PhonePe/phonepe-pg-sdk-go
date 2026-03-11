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

package constants

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test SDK Version Format - should follow semantic versioning
func TestSDKVersionFormat(t *testing.T) {
	version := SDKVersion

	// Validate version is not empty
	require.NotEmpty(t, version, "SDK version should not be empty")

	// Validate semantic versioning format (X.Y.Z)
	semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)

	assert.True(t, semverPattern.MatchString(version),
		"SDK version should follow semantic versioning format (X.Y.Z). Got: %s", version)

	// Split and validate components (major.minor.patch)
	parts := strings.Split(version, ".")

	require.Equal(t, 3, len(parts),
		"SDK version should have exactly 3 parts (major.minor.patch). Got: %s", version)

	// Each part should be numeric
	for i, part := range parts {
		matched, err := regexp.MatchString(`^\d+$`, part)
		require.NoError(t, err)
		assert.True(t, matched,
			"Version part %d should be numeric. Got: %s in version %s", i, part, version)
	}
}

// Test All Header Constants Are Not Null/Empty
func TestAllHeaderConstantsAreNotNullOrEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		required bool
	}{
		{"APIVersion", APIVersion, true},
		{"SubscriptionAPIVersion", SubscriptionAPIVersion, true},
		{"Integration", Integration, true},
		{"SDKVersion", SDKVersion, true},
		{"SDKType", SDKType, true},
		{"Source", Source, true},
		{"SourceVersion", SourceVersion, true},
		{"SourcePlatform", SourcePlatform, true},
		{"SourcePlatformVersion", SourcePlatformVersion, true},
		{"OAuthAuthorization", OAuthAuthorization, true},
		{"ContentType", ContentType, true},
		{"Accept", Accept, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.required {
				assert.NotEmpty(t, tt.value,
					"%s constant should not be empty", tt.name)
			}
		})
	}
}

// Test Header Values Don't Contain Problematic Characters
func TestHeaderValuesDoNotContainProblematicCharacters(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"APIVersion", APIVersion},
		{"SubscriptionAPIVersion", SubscriptionAPIVersion},
		{"SDKVersion", SDKVersion},
		{"SDKType", SDKType},
		{"Integration", Integration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not contain whitespace
			assert.False(t, strings.Contains(tt.value, " "),
				"%s should not contain whitespace. Got: %s", tt.name, tt.value)

			// Should not contain newlines or tabs
			assert.False(t, strings.Contains(tt.value, "\n"),
				"%s should not contain newlines. Got: %s", tt.name, tt.value)
			assert.False(t, strings.Contains(tt.value, "\r"),
				"%s should not contain carriage returns. Got: %s", tt.name, tt.value)
			assert.False(t, strings.Contains(tt.value, "\t"),
				"%s should not contain tabs. Got: %s", tt.name, tt.value)

			// Should not contain null bytes
			assert.False(t, strings.Contains(tt.value, "\x00"),
				"%s should not contain null bytes. Got: %s", tt.name, tt.value)
		})
	}
}

// Test SDK Version Components Can Be Parsed
func TestSDKVersionComponents(t *testing.T) {
	version := SDKVersion

	// Split by dots
	parts := strings.Split(version, ".")

	// Should have exactly 3 parts for semantic version
	assert.Equal(t, 3, len(parts),
		"SDK version should have exactly 3 parts (major.minor.patch). Got: %s with %d parts",
		version, len(parts))

	// Validate each part is numeric
	for i, part := range parts {
		// Check if the part is numeric
		matched, err := regexp.MatchString(`^\d+$`, part)
		require.NoError(t, err)
		assert.True(t, matched,
			"Version part %d (%s) should be numeric in version %s", i, part, version)

		// Additional check: part should not start with zero unless it is zero
		if len(part) > 1 {
			assert.NotEqual(t, '0', part[0],
				"Version part %d (%s) should not have leading zeros (except if it's just '0')",
				i, part)
		}
	}
}

// Test Header Key Format - should follow HTTP header naming conventions
func TestHeaderKeyFormat(t *testing.T) {
	headerKeys := []struct {
		name string
		key  string
	}{
		{"Source", Source},
		{"SourceVersion", SourceVersion},
		{"SourcePlatform", SourcePlatform},
		{"SourcePlatformVersion", SourcePlatformVersion},
		{"OAuthAuthorization", OAuthAuthorization},
		{"ContentType", ContentType},
		{"Accept", Accept},
	}

	for _, hk := range headerKeys {
		t.Run(hk.name, func(t *testing.T) {
			// Header keys should not be empty
			assert.NotEmpty(t, hk.key, "%s header key should not be empty", hk.name)

			// Header keys should not contain whitespace
			assert.False(t, strings.Contains(hk.key, " "),
				"%s header key should not contain whitespace. Got: %s", hk.name, hk.key)

			// Header keys should not contain control characters
			for _, char := range hk.key {
				assert.False(t, char < 32 || char == 127,
					"%s header key should not contain control characters. Got: %s", hk.name, hk.key)
			}
		})
	}
}

// Test API Version Value
func TestAPIVersionValue(t *testing.T) {
	// API Version should be V2 (matching Java SDK)
	assert.Equal(t, "V2", APIVersion, "APIVersion should be 'V2'")
}

// Test SDK Type Value
func TestSDKTypeValue(t *testing.T) {
	// SDK Type should be BACKEND_GO_SDK
	assert.Equal(t, "BACKEND_GO_SDK", SDKType, "SDKType should be 'BACKEND_GO_SDK'")

	// Should not be the old Java SDK value
	assert.NotEqual(t, "JAVA_SDK", SDKType, "SDKType should not be 'JAVA_SDK'")
}

// Test Content Type and Accept Headers
func TestContentTypeAndAcceptHeaders(t *testing.T) {
	// ContentType should be standard HTTP header name
	assert.Equal(t, "Content-Type", ContentType,
		"ContentType header should be 'Content-Type'")

	// Accept should be standard HTTP header name
	assert.Equal(t, "Accept", Accept,
		"Accept header should be 'Accept'")

	// OAuthAuthorization should be standard HTTP header name
	assert.Equal(t, "Authorization", OAuthAuthorization,
		"OAuthAuthorization header should be 'Authorization'")
}

// Test Header Prefix Consistency
func TestHeaderPrefixConsistency(t *testing.T) {
	// Custom headers should use x- prefix (lowercase)
	customHeaders := []struct {
		name  string
		value string
	}{
		{"Source", Source},
		{"SourceVersion", SourceVersion},
		{"SourcePlatform", SourcePlatform},
		{"SourcePlatformVersion", SourcePlatformVersion},
	}

	for _, ch := range customHeaders {
		t.Run(ch.name, func(t *testing.T) {
			assert.True(t, strings.HasPrefix(ch.value, "x-"),
				"%s header should start with 'x-' prefix. Got: %s", ch.name, ch.value)

			// Should be lowercase
			assert.Equal(t, strings.ToLower(ch.value), ch.value,
				"%s header should be lowercase. Got: %s", ch.name, ch.value)
		})
	}
}
