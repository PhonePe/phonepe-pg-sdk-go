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

package customcheckout

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	v2_request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	v2_response "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test CreateSdkOrder - Custom Checkout - Success
func TestCreateSdkOrder_CustomCheckout_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/payments/v2/sdk/order", r.URL.Path)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "TOKEN123",
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

	orderRequest := v2_request.NewCustomCheckoutCreateSdkOrderRequest(
		"ORDER123",
		10000,
		models.MetaInfo{},
		nil,
		3600,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Equal(t, "PENDING", resp.State)
	assert.Equal(t, int64(1234567890), resp.ExpireAt)
	assert.Equal(t, "TOKEN123", resp.Token)
}

// Test CreateSdkOrder - Custom Checkout - DisablePaymentRetry = true
func TestCreateSdkOrder_CustomCheckout_WithDisablePaymentRetryTrue(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "TOKEN123",
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
	orderRequest := v2_request.NewCustomCheckoutCreateSdkOrderRequest(
		"ORDER123",
		10000,
		models.MetaInfo{},
		nil,
		3600,
		&disableRetry,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	require.NotNil(t, capturedRequest.DisablePaymentRetry)
	assert.True(t, *capturedRequest.DisablePaymentRetry)
}

// Test CreateSdkOrder - Custom Checkout - DisablePaymentRetry = false
func TestCreateSdkOrder_CustomCheckout_WithDisablePaymentRetryFalse(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "TOKEN123",
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
	orderRequest := v2_request.NewCustomCheckoutCreateSdkOrderRequest(
		"ORDER123",
		10000,
		models.MetaInfo{},
		nil,
		3600,
		&disableRetry,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	require.NotNil(t, capturedRequest.DisablePaymentRetry)
	assert.False(t, *capturedRequest.DisablePaymentRetry)
}

// Test CreateSdkOrder - Custom Checkout - DisablePaymentRetry = nil
func TestCreateSdkOrder_CustomCheckout_WithDisablePaymentRetryNil(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedRequest *v2_request.CreateSdkOrderRequest

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedRequest)

		orderResponse := v2_response.CreateSdkOrderResponse{
			OrderID:  "ORDER123",
			State:    "PENDING",
			ExpireAt: 1234567890,
			Token:    "TOKEN123",
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

	orderRequest := v2_request.NewCustomCheckoutCreateSdkOrderRequest(
		"ORDER123",
		10000,
		models.MetaInfo{},
		nil,
		3600,
		nil,
	)

	resp, err := client.CreateSdkOrder(context.Background(), orderRequest)

	require.NoError(t, err)
	assert.Equal(t, "ORDER123", resp.OrderID)
	assert.Nil(t, capturedRequest.DisablePaymentRetry)
}
