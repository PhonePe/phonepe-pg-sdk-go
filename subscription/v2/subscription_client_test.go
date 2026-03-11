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

package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PhonePe/phonepe-pg-sdk-go/common"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
	httplib "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create mock OAuth server
func createMockOAuthServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-token-12345",
			"expires_in":   3600,
			"issued_at":    1609459200000,
		})
	}))
}

// Test client initialization
func TestGetInstance(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.BaseClient)
	assert.NotNil(t, client.headers)
	assert.Equal(t, 5, len(client.headers))
}

func TestGetInstanceWithRetry(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	retryConfig := httplib.DefaultRetryConfig()
	client, err := GetInstanceWithRetry("test-client-id", "test-secret", 1, env, false, retryConfig)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.BaseClient)
}

func TestPrepareHeaders(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	// Check headers
	headerMap := make(map[string]string)
	for _, h := range client.headers {
		headerMap[h.Key] = h.Value
	}

	assert.Equal(t, "application/json", headerMap["Content-Type"])
	assert.Equal(t, "INTEGRATION", headerMap["Source"])
	assert.Equal(t, constants.SubscriptionAPIVersion, headerMap["x-source-version"])
	assert.Equal(t, "BACKEND_GO_SDK", headerMap["x-Source-Platform"])
	assert.Equal(t, common.GetSDKVersion(), headerMap["x-Source-Platform-Version"])
}

// Test Setup method
func TestSetup_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/setup", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId":     "SUB_ORDER123",
			"state":       "PENDING",
			"redirectUrl": "https://subscription.phonepe.com/xyz",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	setupRequest := &request.PgPaymentRequest{
		MerchantOrderID: "SUB_ORDER123",
		Amount:          10000,
	}

	response, err := client.Setup(context.Background(), setupRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "SUB_ORDER123", response.OrderID)
}

func TestSetup_Error(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"code":    "BAD_REQUEST",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	setupRequest := &request.PgPaymentRequest{
		MerchantOrderID: "SUB_ORDER123",
		Amount:          10000,
	}

	response, err := client.Setup(context.Background(), setupRequest)

	assert.Error(t, err)
	assert.Nil(t, response)
}

// Test Notify method
func TestNotify_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/notify", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId": "NOTIFY_ORDER123",
			"state":   "COMPLETED",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	notifyRequest := &request.PgPaymentRequest{
		MerchantOrderID: "NOTIFY_ORDER123",
		Amount:          5000,
	}

	response, err := client.Notify(context.Background(), notifyRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "NOTIFY_ORDER123", response.OrderID)
}

// Test Redeem method
func TestRedeem_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/redeem", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchantOrderId": "REDEEM123",
			"state":           "SUCCESS",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	response, err := client.Redeem(context.Background(), "REDEEM123")

	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test GetSubscriptionStatus method
func TestGetSubscriptionStatus_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/SUB123/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchantSubscriptionId": "SUB123",
			"state":                  "ACTIVE",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	response, err := client.GetSubscriptionStatus(context.Background(), "SUB123")

	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test GetOrderStatus method
func TestGetOrderStatus_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/order/ORDER123/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchantOrderId": "ORDER123",
			"state":           "COMPLETED",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	response, err := client.GetOrderStatus(context.Background(), "ORDER123")

	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test CancelSubscription method
func TestCancelSubscription_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/SUB123/cancel", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	err = client.CancelSubscription(context.Background(), "SUB123")

	require.NoError(t, err)
}

// Test GetTransactionStatus method
func TestGetTransactionStatus_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subscriptions/v2/transaction/TXN123/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"transactionId": "TXN123",
			"state":         "SUCCESS",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	response, err := client.GetTransactionStatus(context.Background(), "TXN123")

	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test Refund method
func TestRefund_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/payments/v2/refund", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchantRefundId": "REFUND123",
			"state":            "PENDING",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	refundRequest := &request.RefundRequest{
		MerchantRefundID:        "REFUND123",
		OriginalMerchantOrderID: "ORDER123",
		Amount:                  5000,
	}

	response, err := client.Refund(context.Background(), refundRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test GetRefundStatus method
func TestGetRefundStatus_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/payments/v2/refund/REFUND123/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"merchantRefundId": "REFUND123",
			"state":            "SUCCESS",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	response, err := client.GetRefundStatus(context.Background(), "REFUND123")

	require.NoError(t, err)
	assert.NotNil(t, response)
}

// Test ValidateCallback method
func TestValidateCallback_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	// Create valid callback JSON
	callbackData := map[string]interface{}{
		"merchantSubscriptionId": "SUB123",
		"state":                  "ACTIVE",
	}
	callbackJSON, _ := json.Marshal(callbackData)

	// Calculate proper SHA256 hash for username:password
	// SHA256 of "test-user:test-pass"
	authHash := "75f05a3e2a10a7488c40eac6ef38da95df902e8fc3886531ce97ee8599481584"

	// Validate with correct credentials
	response, err := client.ValidateCallback("test-user", "test-pass", authHash, string(callbackJSON))

	require.NoError(t, err)
	assert.NotNil(t, response)
}

func TestValidateCallback_InvalidAuth(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	callbackJSON := `{"state": "ACTIVE"}`

	// Invalid auth hash
	response, err := client.ValidateCallback("test-user", "test-pass", "invalid-hash", callbackJSON)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid callback")
}

func TestValidateCallback_InvalidJSON(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	authHash := "75f05a3e2a10a7488c40eac6ef38da95df902e8fc3886531ce97ee8599481584"

	// Invalid JSON
	response, err := client.ValidateCallback("test-user", "test-pass", authHash, "invalid-json{")

	assert.Error(t, err)
	assert.Nil(t, response)
}

// Test retry behavior
func TestSetup_WithRetry_SuccessAfterFailure(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	callCount := 0
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call fails with 503
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"code":    "SERVICE_UNAVAILABLE",
			})
			return
		}
		// Second call succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId": "SUB_ORDER123",
			"state":   "PENDING",
		})
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	retryConfig := httplib.DefaultRetryConfig()
	client, err := GetInstanceWithRetry("test-client-id", "test-secret", 1, env, false, retryConfig)
	require.NoError(t, err)

	setupRequest := &request.PgPaymentRequest{
		MerchantOrderID: "SUB_ORDER123",
		Amount:          10000,
	}

	response, err := client.Setup(context.Background(), setupRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, callCount, "Should have retried once")
}

// Test different environment configurations
func TestGetInstance_DifferentEnvironments(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	tests := []struct {
		name string
		env  types.Env
	}{
		{
			name: "Sandbox",
			env:  types.Sandbox,
		},
		{
			name: "Production",
			env:  types.Production,
		},
		{
			name: "Test",
			env:  types.Test,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override OAuth URL for testing
			testEnv := tt.env
			testEnv.OAuthHostURL = oauthServer.URL

			client, err := GetInstance("test-client-id", "test-secret", 1, testEnv, false)

			require.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, testEnv.PgHostURL, client.BaseClient.Env.PgHostURL)
		})
	}
}
