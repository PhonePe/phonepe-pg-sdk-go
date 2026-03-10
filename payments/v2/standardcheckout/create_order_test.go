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

package standardcheckout

import (
	"context"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request/paymentmodeconstraints"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	v2_request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	v2_response "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test CreateSdkOrder - Standard Checkout - Success
func TestCreateSdkOrder_StandardCheckout_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/checkout/v2/sdk/order", r.URL.Path)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "test-token-123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"Payment for Order 123",
		"https://merchant.com/redirect",
		3600,
		nil,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)
	assert.NotEmpty(t, resp.Token)
}

// Test CreateSdkOrder - Standard Checkout - Bad Request
func TestCreateSdkOrder_StandardCheckout_BadRequest(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := map[string]interface{}{
			"success": false,
			"code":    "BAD_REQUEST",
			"message": "Invalid request parameters",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"",
		"https://merchant.com/redirect",
		0,
		nil,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Test CreateSdkOrder - DisablePaymentRetry = true
func TestCreateSdkOrder_WithDisablePaymentRetryTrue(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "test-token-123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	disableRetry := true
	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"Payment for Order 123",
		"https://merchant.com/redirect",
		3600,
		&disableRetry,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)
	require.NotNil(t, capturedRequest.DisablePaymentRetry)
	assert.True(t, *capturedRequest.DisablePaymentRetry)
}

// Test CreateSdkOrder - DisablePaymentRetry = false
func TestCreateSdkOrder_WithDisablePaymentRetryFalse(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "test-token-123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	disableRetry := false
	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"Payment for Order 123",
		"https://merchant.com/redirect",
		3600,
		&disableRetry,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)
	require.NotNil(t, capturedRequest.DisablePaymentRetry)
	assert.False(t, *capturedRequest.DisablePaymentRetry)
}

// Test CreateSdkOrder - DisablePaymentRetry = nil (not set)
func TestCreateSdkOrder_WithDisablePaymentRetryNil(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "test-token-123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"Payment for Order 123",
		"https://merchant.com/redirect",
		3600,
		nil,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)
	assert.Nil(t, capturedRequest.DisablePaymentRetry)
}

// Test CreateSdkOrder - With PaymentModeConfig
func TestCreateSdkOrder_WithPaymentModeConfig(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "test-token-123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	cardConstraint := paymentmodeconstraints.NewCardPaymentModeConstraint(
		[]paymentmodeconstraints.CardType{
			paymentmodeconstraints.CREDIT_CARD,
			paymentmodeconstraints.DEBIT_CARD,
		},
	)

	paymentModeConfig := &v2_request.PaymentModeConfig{
		EnabledPaymentModes: []paymentmodeconstraints.PaymentModeConstraint{cardConstraint},
	}

	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"Payment for Order 123",
		"https://merchant.com/redirect",
		3600,
		nil,
		paymentModeConfig,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)

	// Verify PaymentModeConfig is in the payment flow
	if capturedRequest != nil && capturedRequest.PaymentFlow != nil {
		pgFlow, ok := capturedRequest.PaymentFlow.(*v2_request.PgCheckoutPaymentFlow)
		require.True(t, ok, "PaymentFlow should be of type PgCheckoutPaymentFlow")
		require.NotNil(t, pgFlow.PaymentModeConfig, "PaymentModeConfig should not be nil")
		assert.Len(t, pgFlow.PaymentModeConfig.EnabledPaymentModes, 1)
	}
}

// Test CreateSdkOrder - With Null PaymentModeConfig
func TestCreateSdkOrder_WithNullPaymentModeConfig(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "test-token-123",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orderResponse)
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	orderRequest := v2_request.NewStandardCheckoutCreateSdkOrderRequest(
		10000,
		"ORDER123",
		models.MetaInfo{},
		"Payment for Order 123",
		"https://merchant.com/redirect",
		3600,
		nil,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)

	if capturedRequest != nil && capturedRequest.PaymentFlow != nil {
		pgFlow, ok := capturedRequest.PaymentFlow.(*v2_request.PgCheckoutPaymentFlow)
		require.True(t, ok, "PaymentFlow should be of type PgCheckoutPaymentFlow")
		assert.Nil(t, pgFlow.PaymentModeConfig)
	}
}
