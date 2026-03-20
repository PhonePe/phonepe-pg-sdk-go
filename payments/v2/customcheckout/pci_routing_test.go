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
	request "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// paySuccessHandler returns a handler that writes a successful Pay response
// and records which server received the request.
func paySuccessHandler(received *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if received != nil {
			*received = true
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId": "ORDER123",
			"state":   "PENDING",
		})
	}
}

// createDualServerEnv creates two mock API servers (default PG and PCI) and an OAuth server,
// returning an Env that points to them.
func createDualServerEnv(pgHandler, pciHandler http.HandlerFunc) (env types.Env, oauthServer, pgServer, pciServer *httptest.Server) {
	oauthServer = createMockOAuthServer()
	pgServer = httptest.NewServer(pgHandler)
	pciServer = httptest.NewServer(pciHandler)
	env = types.Env{
		PgHostURL:     pgServer.URL,
		PciPgHostURL:  pciServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}
	return
}

// TestPay_CardInstrument_UsesPciHost verifies that a CARD payment is routed to PciPgHostURL.
func TestPay_CardInstrument_UsesPciHost(t *testing.T) {
	var pgReceived, pciReceived bool

	env, oauthServer, pgServer, pciServer := createDualServerEnv(
		paySuccessHandler(&pgReceived),
		paySuccessHandler(&pciReceived),
	)
	defer oauthServer.Close()
	defer pgServer.Close()
	defer pciServer.Close()

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	payRequest := request.NewCardPayRequest(
		"ORDER_CARD", 10000, 1,
		"OTP", "encCard", "encCvv", "12", "2026", "John", "user1",
		models.MetaInfo{}, nil, "https://redirect.example.com", 300,
	)

	resp, err := client.Pay(context.Background(), payRequest)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	assert.True(t, pciReceived, "CARD payment should be routed to PCI host")
	assert.False(t, pgReceived, "CARD payment must NOT hit the default PG host")
}

// TestPay_TokenInstrument_UsesPciHost verifies that a TOKEN payment is routed to PciPgHostURL.
func TestPay_TokenInstrument_UsesPciHost(t *testing.T) {
	var pgReceived, pciReceived bool

	env, oauthServer, pgServer, pciServer := createDualServerEnv(
		paySuccessHandler(&pgReceived),
		paySuccessHandler(&pciReceived),
	)
	defer oauthServer.Close()
	defer pgServer.Close()
	defer pciServer.Close()

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	payRequest := request.NewTokenPayRequest(
		"ORDER_TOKEN", 10000, 1,
		"OTP", "encToken", "encCvv", "cryptogram", "1234", "12", "2026",
		"https://redirect.example.com", "Jane", "user2",
		models.MetaInfo{}, nil, 300,
	)

	resp, err := client.Pay(context.Background(), payRequest)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	assert.True(t, pciReceived, "TOKEN payment should be routed to PCI host")
	assert.False(t, pgReceived, "TOKEN payment must NOT hit the default PG host")
}

// TestPay_UpiCollect_UsesDefaultHost verifies that a UPI Collect payment uses the default PgHostURL.
func TestPay_UpiCollect_UsesDefaultHost(t *testing.T) {
	var pgReceived, pciReceived bool

	env, oauthServer, pgServer, pciServer := createDualServerEnv(
		paySuccessHandler(&pgReceived),
		paySuccessHandler(&pciReceived),
	)
	defer oauthServer.Close()
	defer pgServer.Close()
	defer pciServer.Close()

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	payRequest := request.NewUpiCollectPayViaVpaRequest(
		10000, "ORDER_UPI", models.MetaInfo{}, nil, "test@upi", "Pay now", 300,
	)

	resp, err := client.Pay(context.Background(), payRequest)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	assert.True(t, pgReceived, "UPI Collect should be routed to the default PG host")
	assert.False(t, pciReceived, "UPI Collect must NOT hit the PCI host")
}

// TestPay_NilPaymentFlow_UsesDefaultHost verifies that a request with no PaymentFlow uses the default host.
func TestPay_NilPaymentFlow_UsesDefaultHost(t *testing.T) {
	var pgReceived, pciReceived bool

	env, oauthServer, pgServer, pciServer := createDualServerEnv(
		paySuccessHandler(&pgReceived),
		paySuccessHandler(&pciReceived),
	)
	defer oauthServer.Close()
	defer pgServer.Close()
	defer pciServer.Close()

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER_NOFLOW",
		Amount:          5000,
	}

	resp, err := client.Pay(context.Background(), payRequest)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	assert.True(t, pgReceived, "Request with nil PaymentFlow should use the default PG host")
	assert.False(t, pciReceived, "Request with nil PaymentFlow must NOT hit the PCI host")
}

// TestIsPciInstrument_Card confirms isPciInstrument returns true for CARD.
func TestIsPciInstrument_Card(t *testing.T) {
	payRequest := request.NewCardPayRequest(
		"ORDER1", 100, 1, "OTP", "encCard", "encCvv", "12", "2026",
		"John", "user1", models.MetaInfo{}, nil, "https://r.example.com", 0,
	)
	assert.True(t, isPciInstrument(payRequest))
}

// TestIsPciInstrument_Token confirms isPciInstrument returns true for TOKEN.
func TestIsPciInstrument_Token(t *testing.T) {
	payRequest := request.NewTokenPayRequest(
		"ORDER2", 100, 1, "OTP", "encTok", "encCvv", "cryp", "1234",
		"12", "2026", "https://r.example.com", "Jane", "user2",
		models.MetaInfo{}, nil, 0,
	)
	assert.True(t, isPciInstrument(payRequest))
}

// TestIsPciInstrument_UpiCollect confirms isPciInstrument returns false for UPI_COLLECT.
func TestIsPciInstrument_UpiCollect(t *testing.T) {
	payRequest := request.NewUpiCollectPayViaVpaRequest(
		100, "ORDER3", models.MetaInfo{}, nil, "test@upi", "msg", 0,
	)
	assert.False(t, isPciInstrument(payRequest))
}

// TestIsPciInstrument_NilFlow confirms isPciInstrument returns false when PaymentFlow is nil.
func TestIsPciInstrument_NilFlow(t *testing.T) {
	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER4",
		Amount:          100,
	}
	assert.False(t, isPciInstrument(payRequest))
}
