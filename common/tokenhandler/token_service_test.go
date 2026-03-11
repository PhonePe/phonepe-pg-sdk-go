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

package tokenhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/configs"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock event publisher that does nothing
type mockEventPublisher struct{}

func (m *mockEventPublisher) SetAuthTokenSupplier(authTokenSupplier func(context.Context) (string, error)) {
}

func (m *mockEventPublisher) StartPublishingEvents(authTokenSupplier func(context.Context) (string, error)) {
}

func (m *mockEventPublisher) Send(baseEvent *models.BaseEvent) {}

func (m *mockEventPublisher) Run() {}

func (m *mockEventPublisher) Stop() {}

// createMockOAuthServer creates a mock OAuth server for testing
func createMockOAuthServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// createTestTokenService creates a token service for testing
func createTestTokenService(oauthURL string) *TokenService {
	credConfig := &configs.CredentialConfig{
		ClientId:      "test-client-id",
		ClientSecret:  "test-client-secret",
		ClientVersion: 1,
	}

	env := types.Env{
		OAuthHostURL:  oauthURL,
		PgHostURL:     "https://api.test.com",
		EventsHostURL: "https://events.test.com",
	}

	eventPublisher := &mockEventPublisher{}
	testLogger := logger.NewLogger(logger.DisabledLogConfig())

	return NewTokenService(http.DefaultClient, credConfig, env, eventPublisher, testLogger)
}

// createValidOAuthResponse creates a valid OAuth response
func createValidOAuthResponse(issuedAt, expiresAt int64) *OAuthResponse {
	return &OAuthResponse{
		AccessToken:          "test-access-token-12345",
		EncryptedAccessToken: "encrypted-token",
		RefreshToken:         "refresh-token",
		ExpiresIn:            3600,
		IssuedAt:             issuedAt,
		ExpiresAt:            expiresAt,
		SessionExpiresAt:     expiresAt + 3600,
		TokenType:            "Bearer",
	}
}

// Test NewTokenService
func TestNewTokenService(t *testing.T) {
	credConfig := &configs.CredentialConfig{
		ClientId:      "test-client-id",
		ClientSecret:  "test-client-secret",
		ClientVersion: 1,
	}

	env := types.Env{
		OAuthHostURL:  "https://oauth.test.com",
		PgHostURL:     "https://api.test.com",
		EventsHostURL: "https://events.test.com",
	}

	eventPublisher := &mockEventPublisher{}
	testLogger := logger.NewLogger(logger.DisabledLogConfig())

	ts := NewTokenService(http.DefaultClient, credConfig, env, eventPublisher, testLogger)

	assert.NotNil(t, ts)
	assert.Equal(t, credConfig, ts.CredentialConfig)
	assert.Equal(t, env, ts.Env)
	assert.Equal(t, http.DefaultClient, ts.HttpClient)
	assert.Equal(t, eventPublisher, ts.EventPublisher)
	assert.Nil(t, ts.oAuthResponse, "Initial OAuth response should be nil")
}

// Test prepareRequestHeaders
func TestPrepareRequestHeaders(t *testing.T) {
	ts := createTestTokenService("https://oauth.test.com")
	headers := ts.prepareRequestHeaders()

	assert.Len(t, headers, 2)
	assert.Equal(t, "Content-Type", headers[0].Key)
	assert.Equal(t, "application/x-www-form-urlencoded", headers[0].Value)
	assert.Equal(t, "Accept", headers[1].Key)
	assert.Equal(t, "application/json", headers[1].Value)
}

// Test formatCachedToken
func TestFormatCachedToken(t *testing.T) {
	ts := createTestTokenService("https://oauth.test.com")
	ts.oAuthResponse = &OAuthResponse{
		AccessToken: "test-token-123",
		TokenType:   "Bearer",
	}

	formatted := ts.formatCachedToken()
	assert.Equal(t, "Bearer test-token-123", formatted)
}

// Test isCachedTokenValid
func TestIsCachedTokenValid(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name          string
		oAuthResponse *OAuthResponse
		expectedValid bool
	}{
		{
			name:          "nil response",
			oAuthResponse: nil,
			expectedValid: false,
		},
		{
			name: "valid token - just issued",
			oAuthResponse: &OAuthResponse{
				IssuedAt:  now,
				ExpiresAt: now + 3600, // expires in 1 hour
			},
			expectedValid: true,
		},
		{
			name: "valid token - before half life",
			oAuthResponse: &OAuthResponse{
				IssuedAt:  now - 1000,
				ExpiresAt: now + 3000, // total 4000s, half life at 2000s, current at 1000s
			},
			expectedValid: true,
		},
		{
			name: "expired token - past half life",
			oAuthResponse: &OAuthResponse{
				IssuedAt:  now - 2000,
				ExpiresAt: now + 1000, // total 3000s, half life at 1500s, current at 2000s
			},
			expectedValid: false,
		},
		{
			name: "expired token - fully expired",
			oAuthResponse: &OAuthResponse{
				IssuedAt:  now - 4000,
				ExpiresAt: now - 1000,
			},
			expectedValid: false,
		},
		{
			name: "token at exact half life boundary",
			oAuthResponse: &OAuthResponse{
				IssuedAt:  now - 1800,
				ExpiresAt: now + 1800, // total 3600s, half life at 1800s
			},
			expectedValid: false, // Should be false as current >= reloadTime
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := createTestTokenService("https://oauth.test.com")
			ts.oAuthResponse = tt.oAuthResponse
			isValid := ts.isCachedTokenValid()
			assert.Equal(t, tt.expectedValid, isValid)
		})
	}
}

// Test prepareFormBody
func TestPrepareFormBody(t *testing.T) {
	credConfig := &configs.CredentialConfig{
		ClientId:      "test-client-id",
		ClientSecret:  "test-client-secret",
		ClientVersion: 1,
	}

	ts := createTestTokenService("https://oauth.test.com")
	formBody := ts.prepareFormBody(credConfig)

	assert.Equal(t, "test-client-id", formBody.Get("client_id"))
	assert.Equal(t, "test-client-secret", formBody.Get("client_secret"))
	assert.Equal(t, "client_credentials", formBody.Get("grant_type"))
	assert.Equal(t, "1", formBody.Get("client_version"))
}

// Test fetchTokenFromPhonePe - Success
func TestFetchTokenFromPhonePe_Success(t *testing.T) {
	now := time.Now().Unix()
	expectedResponse := createValidOAuthResponse(now, now+3600)

	// Create mock server
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		assert.Equal(t, "POST", r.Method)

		// Verify headers
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		// Verify form data
		err := r.ParseForm()
		require.NoError(t, err)
		assert.Equal(t, "test-client-id", r.FormValue("client_id"))
		assert.Equal(t, "test-client-secret", r.FormValue("client_secret"))
		assert.Equal(t, "client_credentials", r.FormValue("grant_type"))
		assert.Equal(t, "1", r.FormValue("client_version"))

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	response, err := ts.fetchTokenFromPhonePe(context.Background())

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, expectedResponse.AccessToken, response.AccessToken)
	assert.Equal(t, expectedResponse.TokenType, response.TokenType)
	assert.Equal(t, expectedResponse.ExpiresIn, response.ExpiresIn)
}

// Test fetchTokenFromPhonePe - Error
func TestFetchTokenFromPhonePe_Error(t *testing.T) {
	// Create mock server that returns error
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_client"}`))
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	response, err := ts.fetchTokenFromPhonePe(context.Background())

	assert.Error(t, err)
	assert.Nil(t, response)
}

// Test GetAuthToken - First Time Fetch
func TestGetAuthToken_FirstTimeFetch(t *testing.T) {
	now := time.Now().Unix()
	expectedResponse := createValidOAuthResponse(now, now+3600)

	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	token, err := ts.GetAuthToken(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "Bearer test-access-token-12345", token)
	assert.NotNil(t, ts.oAuthResponse)
	assert.Equal(t, expectedResponse.AccessToken, ts.oAuthResponse.AccessToken)
}

// Test GetAuthToken - Return Cached Token
func TestGetAuthToken_ReturnCachedToken(t *testing.T) {
	now := time.Now().Unix()
	cachedResponse := createValidOAuthResponse(now, now+3600)

	// Create a server that should NOT be called
	callCount := 0
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cachedResponse)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	ts.oAuthResponse = cachedResponse // Pre-populate with valid cached token

	token, err := ts.GetAuthToken(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "Bearer test-access-token-12345", token)
	assert.Equal(t, 0, callCount, "Server should not be called when token is cached")
}

// Test GetAuthToken - Refresh Expired Token
func TestGetAuthToken_RefreshExpiredToken(t *testing.T) {
	now := time.Now().Unix()

	// Old expired token
	oldResponse := createValidOAuthResponse(now-4000, now-1000)

	// New token to be fetched
	newResponse := createValidOAuthResponse(now, now+3600)
	newResponse.AccessToken = "new-access-token-67890"

	callCount := 0
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(newResponse)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	ts.oAuthResponse = oldResponse // Pre-populate with expired token

	token, err := ts.GetAuthToken(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "Bearer new-access-token-67890", token)
	assert.Equal(t, 1, callCount, "Server should be called once to refresh token")
	assert.Equal(t, newResponse.AccessToken, ts.oAuthResponse.AccessToken)
}

// Test GetAuthToken - Fallback to Cached on Error
func TestGetAuthToken_FallbackToCachedOnError(t *testing.T) {
	now := time.Now().Unix()

	// Expired cached token
	cachedResponse := createValidOAuthResponse(now-2000, now-100)

	callCount := 0
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Return error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server_error"}`))
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	ts.oAuthResponse = cachedResponse // Pre-populate with expired token

	token, err := ts.GetAuthToken(context.Background())

	// Should NOT return error, should fallback to cached token
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-access-token-12345", token)
	assert.Equal(t, 1, callCount, "Server should be called once (failed)")
}

// Test GetAuthToken - Error When No Cache
func TestGetAuthToken_ErrorWhenNoCache(t *testing.T) {
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)
	// No cached token

	token, err := ts.GetAuthToken(context.Background())

	assert.Error(t, err)
	assert.Equal(t, "", token)
	assert.Nil(t, ts.oAuthResponse)
}

// Test ForceRefreshToken - Success
func TestForceRefreshToken_Success(t *testing.T) {
	now := time.Now().Unix()
	newResponse := createValidOAuthResponse(now, now+3600)
	newResponse.AccessToken = "force-refreshed-token"

	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(newResponse)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)

	// Set an old token
	ts.oAuthResponse = createValidOAuthResponse(now-1000, now-100)

	ts.ForceRefreshToken()

	// Give it a moment to complete (it runs in a goroutine-like pattern)
	time.Sleep(100 * time.Millisecond)

	assert.NotNil(t, ts.oAuthResponse)
	assert.Equal(t, "force-refreshed-token", ts.oAuthResponse.AccessToken)
}

// Test ForceRefreshToken - Error (should not panic)
func TestForceRefreshToken_Error(t *testing.T) {
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server_error"}`))
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)

	// This should not panic
	assert.NotPanics(t, func() {
		ts.ForceRefreshToken()
		time.Sleep(100 * time.Millisecond)
	})
}

// Test Concurrent Token Access
func TestGetAuthToken_ConcurrentAccess(t *testing.T) {
	now := time.Now().Unix()
	response := createValidOAuthResponse(now, now+3600)

	callCount := 0
	var mu sync.Mutex

	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		// Simulate slow response
		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)

	// Launch 10 concurrent requests
	var wg sync.WaitGroup
	tokens := make([]string, 10)
	errors := make([]error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			token, err := ts.GetAuthToken(context.Background())
			tokens[index] = token
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All should succeed
	for i := 0; i < 10; i++ {
		assert.NoError(t, errors[i])
		assert.Equal(t, "Bearer test-access-token-12345", tokens[i])
	}

	// Due to mutex, only one request should have been made
	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, callCount, "Only one request should be made due to mutex")
}

// Test Token Caching Logic with Time Boundary
func TestTokenCachingWithTimeBoundary(t *testing.T) {
	// Test that token is considered valid only before half-life
	now := time.Now().Unix()

	ts := createTestTokenService("https://oauth.test.com")

	// Token issued now, expires in 1000 seconds
	// Half-life is at 500 seconds
	ts.oAuthResponse = &OAuthResponse{
		IssuedAt:    now,
		ExpiresAt:   now + 1000,
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}

	// Immediately after issue - should be valid
	assert.True(t, ts.isCachedTokenValid())

	// Simulate time passing by modifying issued time
	// Now at half-life exactly - should be invalid
	ts.oAuthResponse.IssuedAt = now - 500
	ts.oAuthResponse.ExpiresAt = now + 500
	assert.False(t, ts.isCachedTokenValid())

	// Past half-life - definitely invalid
	ts.oAuthResponse.IssuedAt = now - 600
	ts.oAuthResponse.ExpiresAt = now + 400
	assert.False(t, ts.isCachedTokenValid())
}

// Test prepareFormBody with Different Versions
func TestPrepareFormBody_DifferentVersions(t *testing.T) {
	tests := []struct {
		name          string
		clientVersion int
		expectedValue string
	}{
		{"version 1", 1, "1"},
		{"version 2", 2, "2"},
		{"version 10", 10, "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credConfig := &configs.CredentialConfig{
				ClientId:      "test-client-id",
				ClientSecret:  "test-client-secret",
				ClientVersion: tt.clientVersion,
			}

			ts := createTestTokenService("https://oauth.test.com")
			formBody := ts.prepareFormBody(credConfig)

			assert.Equal(t, tt.expectedValue, formBody.Get("client_version"))
		})
	}
}

// Test URL Encoding in Form Body
func TestPrepareFormBody_URLEncoding(t *testing.T) {
	credConfig := &configs.CredentialConfig{
		ClientId:      "client+id@test",
		ClientSecret:  "secret&key=value",
		ClientVersion: 1,
	}

	ts := createTestTokenService("https://oauth.test.com")
	formBody := ts.prepareFormBody(credConfig)

	// Values should be stored correctly (url.Values handles encoding)
	assert.Equal(t, "client+id@test", formBody.Get("client_id"))
	assert.Equal(t, "secret&key=value", formBody.Get("client_secret"))

	// Verify encoding when converted to string
	encoded := formBody.Encode()
	assert.Contains(t, encoded, url.QueryEscape("client+id@test"))
}

// Test Multiple Sequential Token Fetches
func TestMultipleSequentialTokenFetches(t *testing.T) {
	now := time.Now().Unix()

	fetchCount := 0
	server := createMockOAuthServer(t, func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		response := createValidOAuthResponse(now+int64(fetchCount*1000), now+int64(fetchCount*1000)+3600)
		response.AccessToken = fmt.Sprintf("token-%d", fetchCount)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})
	defer server.Close()

	ts := createTestTokenService(server.URL)

	// First fetch
	token1, err := ts.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Bearer token-1", token1)

	// Second call should use cache
	token2, err := ts.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Bearer token-1", token2) // Same token

	assert.Equal(t, 1, fetchCount, "Should only fetch once")

	// Expire the token by modifying the cached response
	ts.oAuthResponse.IssuedAt = now - 2000
	ts.oAuthResponse.ExpiresAt = now - 100

	// Third call should fetch new token
	token3, err := ts.GetAuthToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Bearer token-2", token3)

	assert.Equal(t, 2, fetchCount, "Should fetch twice")
}
