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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/exception"
	commonHttp "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock response structure for testing
type MockAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// createTestBaseClient creates a base client for testing with mock servers
func createTestBaseClient(pgHostURL, oauthHostURL string, retryConfig *commonHttp.RetryConfig) (*BaseClient, error) {
	return NewBaseClientWithRetry(
		"test-client-id",
		"test-client-secret",
		1,
		types.Env{
			PgHostURL:     pgHostURL,
			OAuthHostURL:  oauthHostURL,
			EventsHostURL: "https://events.test.com",
		},
		false, // don't publish events in tests
		retryConfig,
	)
}

// Test NewBaseClient
func TestNewBaseClient(t *testing.T) {
	client, err := NewBaseClient(
		"test-client-id",
		"test-client-secret",
		1,
		types.Test,
		false,
	)

	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.HttpClient)
	assert.NotNil(t, client.TokenService)
	assert.NotNil(t, client.CredentialConfig)
	assert.NotNil(t, client.EventPublisher)
	assert.NotNil(t, client.EventPublisherFactory)
	assert.False(t, client.ShouldPublishEvents)
	assert.Nil(t, client.RetryConfig, "RetryConfig should be nil when using NewBaseClient")

	// Verify credential config
	assert.Equal(t, "test-client-id", client.CredentialConfig.ClientId)
	assert.Equal(t, "test-client-secret", client.CredentialConfig.ClientSecret)
	assert.Equal(t, 1, client.CredentialConfig.ClientVersion)
}

// Test NewBaseClientWithRetry
func TestNewBaseClientWithRetry(t *testing.T) {
	retryConfig := commonHttp.DefaultRetryConfig()

	client, err := NewBaseClientWithRetry(
		"test-client-id",
		"test-client-secret",
		2,
		types.Test,
		true, // enable event publishing
		retryConfig,
	)

	require.NoError(t, err)
	require.NotNil(t, client)

	assert.NotNil(t, client.HttpClient)
	assert.NotNil(t, client.TokenService)
	assert.NotNil(t, client.CredentialConfig)
	assert.NotNil(t, client.EventPublisher)
	assert.True(t, client.ShouldPublishEvents)
	assert.NotNil(t, client.RetryConfig, "RetryConfig should be set")
	assert.Equal(t, retryConfig, client.RetryConfig)

	// Verify credential config
	assert.Equal(t, "test-client-id", client.CredentialConfig.ClientId)
	assert.Equal(t, "test-client-secret", client.CredentialConfig.ClientSecret)
	assert.Equal(t, 2, client.CredentialConfig.ClientVersion)
}

// Test NewBaseClientWithRetry with nil retry config
func TestNewBaseClientWithRetry_NilRetryConfig(t *testing.T) {
	client, err := NewBaseClientWithRetry(
		"test-client-id",
		"test-client-secret",
		1,
		types.Test,
		false,
		nil, // nil retry config
	)

	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Nil(t, client.RetryConfig, "RetryConfig should remain nil")
}

// Test AddAuthHeader
func TestAddAuthHeader(t *testing.T) {
	// Create mock OAuth server
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	client, err := createTestBaseClient("https://api.test.com", oauthServer.URL, nil)
	require.NoError(t, err)

	// Test adding auth header
	initialHeaders := []*commonHttp.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
	}

	headersWithAuth, err := client.AddAuthHeader(context.Background(), initialHeaders)

	require.NoError(t, err)
	assert.Len(t, headersWithAuth, 2)
	assert.Equal(t, "Content-Type", headersWithAuth[0].Key)
	assert.Equal(t, "application/json", headersWithAuth[0].Value)
	assert.Equal(t, "Authorization", headersWithAuth[1].Key)
	assert.Equal(t, "Bearer test-token-123", headersWithAuth[1].Value)
}

// Test AddAuthHeader when token fetch fails
func TestAddAuthHeader_TokenFetchFails(t *testing.T) {
	// Create mock OAuth server that always fails
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer oauthServer.Close()

	client, err := createTestBaseClient("https://api.test.com", oauthServer.URL, nil)
	require.NoError(t, err)

	initialHeaders := []*commonHttp.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
	}

	// Should return error when token fetch fails
	headersWithAuth, err := client.AddAuthHeader(context.Background(), initialHeaders)

	assert.Error(t, err, "Should return error when token fetch fails")
	assert.Len(t, headersWithAuth, 1, "Should return original headers when token fetch fails")
	assert.Equal(t, "Content-Type", headersWithAuth[0].Key)
}

// Test RequestViaAuthRefresh - Success
func TestRequestViaAuthRefresh_Success(t *testing.T) {
	// Create mock OAuth server
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	// Create mock API server
	apiCallCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCallCount++

		// Verify auth header is present
		authHeader := r.Header.Get("Authorization")
		assert.NotEmpty(t, authHeader, "Auth header should be present")
		assert.Equal(t, "Bearer test-token-123", authHeader)

		response := MockAPIResponse{
			Success: true,
			Message: "Success",
			Data:    "test-data",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer apiServer.Close()

	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, nil)
	require.NoError(t, err)

	// Make request
	var response MockAPIResponse
	headers := []*commonHttp.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
	}

	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		nil,
		&response,
		headers,
	)

	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Success", response.Message)
	assert.Equal(t, "test-data", response.Data)
	assert.Equal(t, 1, apiCallCount, "API should be called once")
}

// Test RequestViaAuthRefresh - With POST data
func TestRequestViaAuthRefresh_WithPostData(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	var receivedData map[string]string
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's POST
		assert.Equal(t, "POST", r.Method)

		// Read request body
		json.NewDecoder(r.Body).Decode(&receivedData)

		response := MockAPIResponse{
			Success: true,
			Message: "Received",
			Data:    "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer apiServer.Close()

	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, nil)
	require.NoError(t, err)

	// Make POST request with data
	requestData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.POST,
		requestData,
		"/test-endpoint",
		nil,
		&response,
		[]*commonHttp.HttpHeaderPair{},
	)

	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "value1", receivedData["key1"])
	assert.Equal(t, "value2", receivedData["key2"])
}

// Test RequestViaAuthRefresh - 401 triggers token refresh
func TestRequestViaAuthRefresh_401TriggersRefresh(t *testing.T) {
	tokenRefreshCount := 0
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenRefreshCount++
		response := map[string]interface{}{
			"access_token": "refreshed-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 401
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unauthorized",
		})
	}))
	defer apiServer.Close()

	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, nil)
	require.NoError(t, err)

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		nil,
		&response,
		[]*commonHttp.HttpHeaderPair{},
	)

	// Should get UnauthorizedAccess error
	assert.Error(t, err)
	assert.IsType(t, &exception.UnauthorizedAccess{}, err)

	// Token refresh should have been triggered (2 calls: initial + force refresh)
	assert.GreaterOrEqual(t, tokenRefreshCount, 2, "Token should be refreshed after 401")
}

// Test RequestViaAuthRefresh - With retry config (success after retries)
func TestRequestViaAuthRefresh_WithRetry_SuccessAfterRetries(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	apiCallCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCallCount++

		// Fail first 2 times, succeed on 3rd
		if apiCallCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Server error"}`))
			return
		}

		response := MockAPIResponse{
			Success: true,
			Message: "Success after retries",
			Data:    "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer apiServer.Close()

	retryConfig := commonHttp.DefaultRetryConfig()
	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, retryConfig)
	require.NoError(t, err)

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		nil,
		&response,
		[]*commonHttp.HttpHeaderPair{},
	)

	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Success after retries", response.Message)
	assert.Equal(t, 3, apiCallCount, "Should retry and succeed on 3rd attempt")
}

// Test RequestViaAuthRefresh - With retry disabled
func TestRequestViaAuthRefresh_NoRetryWhenDisabled(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	apiCallCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCallCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Server error"}`))
	}))
	defer apiServer.Close()

	// Create client without retry config
	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, nil)
	require.NoError(t, err)

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		nil,
		&response,
		[]*commonHttp.HttpHeaderPair{},
	)

	assert.Error(t, err)
	assert.Equal(t, 1, apiCallCount, "Should NOT retry when retry config is nil")
}

// Test RequestViaAuthRefresh - With query parameters
func TestRequestViaAuthRefresh_WithQueryParams(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	var receivedQuery map[string]string
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract query parameters
		query := r.URL.Query()
		receivedQuery = make(map[string]string)
		for key, values := range query {
			if len(values) > 0 {
				receivedQuery[key] = values[0]
			}
		}

		response := MockAPIResponse{
			Success: true,
			Message: "Success",
			Data:    "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer apiServer.Close()

	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, nil)
	require.NoError(t, err)

	queryParams := map[string]string{
		"param1": "value1",
		"param2": "value2",
	}

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		queryParams,
		&response,
		[]*commonHttp.HttpHeaderPair{},
	)

	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "value1", receivedQuery["param1"])
	assert.Equal(t, "value2", receivedQuery["param2"])
}

// Test RequestViaAuthRefresh - Non-retryable error
func TestRequestViaAuthRefresh_NonRetryableError(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	apiCallCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCallCount++
		// 400 Bad Request - non-retryable
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Bad request"}`))
	}))
	defer apiServer.Close()

	retryConfig := commonHttp.DefaultRetryConfig()
	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, retryConfig)
	require.NoError(t, err)

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		nil,
		&response,
		[]*commonHttp.HttpHeaderPair{},
	)

	assert.Error(t, err)
	assert.Equal(t, 1, apiCallCount, "Should NOT retry for 400 error")
}

// Test GetObjectMapper
func TestGetObjectMapper(t *testing.T) {
	client, err := NewBaseClient(
		"test-client-id",
		"test-client-secret",
		1,
		types.Test,
		false,
	)
	require.NoError(t, err)

	mapper := client.GetObjectMapper()
	assert.Nil(t, mapper, "GetObjectMapper returns nil in Go implementation")
}

// Test header copying doesn't mutate original
func TestRequestViaAuthRefresh_HeadersCopied(t *testing.T) {
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"access_token": "test-token-123",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"issued_at":    1234567890,
			"expires_at":   1234571490,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := MockAPIResponse{Success: true}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer apiServer.Close()

	client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, nil)
	require.NoError(t, err)

	originalHeaders := []*commonHttp.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Custom-Header", Value: "custom-value"},
	}
	originalLen := len(originalHeaders)

	var response MockAPIResponse
	err = client.RequestViaAuthRefresh(
		context.Background(),
		commonHttp.GET,
		nil,
		"/test-endpoint",
		nil,
		&response,
		originalHeaders,
	)

	require.NoError(t, err)
	// Original headers should not be modified
	assert.Len(t, originalHeaders, originalLen, "Original headers should not be mutated")
}

// Test different retry configurations
func TestRequestViaAuthRefresh_DifferentRetryConfigs(t *testing.T) {
	tests := []struct {
		name          string
		retryConfig   *commonHttp.RetryConfig
		failCount     int
		shouldSucceed bool
		expectedCalls int
	}{
		{
			name:          "No retry config",
			retryConfig:   nil,
			failCount:     2,
			shouldSucceed: false,
			expectedCalls: 1,
		},
		{
			name:          "Default retry - success",
			retryConfig:   commonHttp.DefaultRetryConfig(),
			failCount:     2,
			shouldSucceed: true,
			expectedCalls: 3,
		},
		{
			name:          "Conservative retry - fail",
			retryConfig:   commonHttp.ConservativeRetryConfig(),
			failCount:     5,
			shouldSucceed: false,
			expectedCalls: 2, // Conservative only retries once
		},
		{
			name:          "Aggressive retry - success",
			retryConfig:   commonHttp.AggressiveRetryConfig(),
			failCount:     4,
			shouldSucceed: true,
			expectedCalls: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"access_token": "test-token",
					"token_type":   "Bearer",
					"expires_in":   3600,
					"issued_at":    1234567890,
					"expires_at":   1234571490,
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer oauthServer.Close()

			apiCallCount := 0
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				apiCallCount++

				if apiCallCount <= tt.failCount {
					w.WriteHeader(http.StatusServiceUnavailable)
					w.Write([]byte(`{"message": "Service unavailable"}`))
					return
				}

				response := MockAPIResponse{Success: true}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}))
			defer apiServer.Close()

			client, err := createTestBaseClient(apiServer.URL, oauthServer.URL, tt.retryConfig)
			require.NoError(t, err)

			var response MockAPIResponse
			err = client.RequestViaAuthRefresh(
				context.Background(),
				commonHttp.GET,
				nil,
				"/test-endpoint",
				nil,
				&response,
				[]*commonHttp.HttpHeaderPair{},
			)

			if tt.shouldSucceed {
				assert.NoError(t, err)
				assert.True(t, response.Success)
			} else {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.expectedCalls, apiCallCount)
		})
	}
}

// Test multiple environments
func TestNewBaseClient_DifferentEnvironments(t *testing.T) {
	envs := []struct {
		name string
		env  types.Env
	}{
		{"Sandbox", types.Sandbox},
		{"Production", types.Production},
		{"Test", types.Test},
	}

	for _, tc := range envs {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewBaseClient(
				"client-id",
				"client-secret",
				1,
				tc.env,
				false,
			)

			require.NoError(t, err)
			require.NotNil(t, client)
			assert.Equal(t, tc.env, client.Env)
			assert.Equal(t, tc.env.PgHostURL, client.Env.PgHostURL)
			assert.Equal(t, tc.env.OAuthHostURL, client.Env.OAuthHostURL)
		})
	}
}
