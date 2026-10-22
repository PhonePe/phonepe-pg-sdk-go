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

	"github.com/PhonePe/phonepe-pg-sdk-go/common/exception"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/paymentinstruments"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/rails"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Order Status with Multiple Payment Attempts
func TestOrderStatus_MultiplePaymentAttempts(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/checkout/v2/order/")
		assert.Contains(t, r.URL.Path, "/status")

		// Simulate order with 2 payment attempts (one failed, one succeeded)
		orderResponse := map[string]interface{}{
			"orderId":  "OMO123",
			"state":    "COMPLETED",
			"amount":   10000,
			"expireAt": 1709108893,
			"paymentDetails": []map[string]interface{}{
				{
					"transactionId": "TXN1",
					"paymentMode":   "UPI_INTENT",
					"timestamp":     1709108475,
					"amount":        10000,
					"state":         "FAILED",
					"errorCode":     "PAYMENT_ERROR",
				},
				{
					"transactionId": "TXN2",
					"paymentMode":   "UPI_COLLECT",
					"timestamp":     1709108490,
					"amount":        10000,
					"state":         "COMPLETED",
				},
			},
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

	status, err := client.GetOrderStatus(context.Background(), "ORDER123", true)

	require.NoError(t, err)
	assert.Equal(t, "OMO123", status.OrderID)
	assert.Equal(t, "COMPLETED", status.State)
	assert.Len(t, status.PaymentDetails, 2, "Should have 2 payment attempts")

	// First attempt failed
	assert.Equal(t, "FAILED", status.PaymentDetails[0].State)
	assert.Equal(t, "PAYMENT_ERROR", status.PaymentDetails[0].ErrorCode)

	// Second attempt succeeded
	assert.Equal(t, "COMPLETED", status.PaymentDetails[1].State)
}

// Test Order Status - 404 Not Found
func TestOrderStatus_404NotFound(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := map[string]interface{}{
			"code":    "ORDER_NOT_FOUND",
			"message": "Not Found",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
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

	_, err = client.GetOrderStatus(context.Background(), "INVALID_ORDER", false)

	require.Error(t, err)

	notFoundErr, ok := err.(*exception.ResourceNotFound)
	require.True(t, ok, "Error should be ResourceNotFound")

	require.NotNil(t, notFoundErr.HttpStatusCode)
	assert.Equal(t, 404, *notFoundErr.HttpStatusCode)
	assert.Contains(t, notFoundErr.Message, "Not Found")
}

// Test Order Status - 502 Bad Gateway
func TestOrderStatus_502BadGateway(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse := map[string]interface{}{
			"code":    "BAD_GATEWAY",
			"message": "Bad Gateway",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
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

	_, err = client.GetOrderStatus(context.Background(), "ORDER123", false)

	require.Error(t, err)

	// 502 maps to PhonePeException (not in exception mapper switch)
	phonePeErr, ok := err.(*exception.PhonePeException)
	require.True(t, ok, "Error should be PhonePeException")

	require.NotNil(t, phonePeErr.HttpStatusCode)
	assert.Equal(t, 502, *phonePeErr.HttpStatusCode)
	assert.Contains(t, phonePeErr.Message, "Bad Gateway")
}

// Test Order Status - Token Payment Method
func TestOrderStatus_TokenPaymentMethod(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderResponse := map[string]interface{}{
			"orderId":  "OMO123",
			"state":    "COMPLETED",
			"amount":   10000,
			"expireAt": 1709108893,
			"paymentDetails": []map[string]interface{}{
				{
					"transactionId": "TXN123",
					"paymentMode":   "TOKEN",
					"timestamp":     1709108475,
					"amount":        10000,
					"state":         "COMPLETED",
				},
			},
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

	status, err := client.GetOrderStatus(context.Background(), "ORDER123", true)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", status.State)
	assert.Len(t, status.PaymentDetails, 1)
	assert.Equal(t, models.TOKEN, status.PaymentDetails[0].PaymentMode)
}

// Test Order Status - Card Payment Method
func TestOrderStatus_CardPaymentMethod(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderResponse := map[string]interface{}{
			"orderId":  "OMO123",
			"state":    "COMPLETED",
			"amount":   10000,
			"expireAt": 1709108893,
			"paymentDetails": []map[string]interface{}{
				{
					"transactionId": "TXN123",
					"paymentMode":   "CARD",
					"timestamp":     1709108475,
					"amount":        10000,
					"state":         "COMPLETED",
				},
			},
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

	status, err := client.GetOrderStatus(context.Background(), "ORDER123", true)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", status.State)
	assert.Len(t, status.PaymentDetails, 1)
	assert.Equal(t, models.CARD, status.PaymentDetails[0].PaymentMode)
}

// Test Order Status - NetBanking Payment Method
func TestOrderStatus_NetBankingPaymentMethod(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderResponse := map[string]interface{}{
			"orderId":  "OMO123",
			"state":    "COMPLETED",
			"amount":   10000,
			"expireAt": 1709108893,
			"paymentDetails": []map[string]interface{}{
				{
					"transactionId": "TXN123",
					"paymentMode":   "NET_BANKING",
					"timestamp":     1709108475,
					"amount":        10000,
					"state":         "COMPLETED",
				},
			},
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

	status, err := client.GetOrderStatus(context.Background(), "ORDER123", true)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", status.State)
	assert.Len(t, status.PaymentDetails, 1)
	assert.Equal(t, models.NET_BANKING, status.PaymentDetails[0].PaymentMode)
}

// Test Order Status - Credit Line instrument within splitInstruments
func TestOrderStatus_CreditLineInstrument(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderResponse := map[string]interface{}{
			"merchantId":      "PRODTEST",
			"merchantOrderId": "28D43E2BAD6411EB85B692E8A924135",
			"orderId":         "OMO2607151547579376222138V",
			"state":           "COMPLETED",
			"amount":          100,
			"expireAt":        1784111877938,
			"paymentDetails": []map[string]interface{}{
				{
					"transactionId": "OM2607151547580748627290V",
					"paymentMode":   "UPI_INTENT",
					"timestamp":     1784110678105,
					"amount":        100,
					"state":         "COMPLETED",
					"instrument": map[string]interface{}{
						"type":                "CREDIT_LINE",
						"ifsc":                "KARB00CLUPI",
						"accountHolderName":   "ANIL SASEENDRAN",
						"bankId":              "KBCL",
						"maskedAccountNumber": "XXXXXXXXXXXX2001",
						"providerAccountType": "CREDITLINE",
					},
					"rail": map[string]interface{}{
						"type":             "UPI",
						"utr":              "287823703443",
						"upiTransactionId": "YBL85ca838b05c1452a87742f14f987f864",
						"vpa":              "96XXXXXXXX-7@ybl",
					},
					"splitInstruments": []map[string]interface{}{
						{
							"instrument": map[string]interface{}{
								"type":                "CREDIT_LINE",
								"ifsc":                "KARB00CLUPI",
								"accountHolderName":   "ANIL SASEENDRAN",
								"bankId":              "KBCL",
								"maskedAccountNumber": "XXXXXXXXXXXX2001",
								"providerAccountType": "CREDITLINE",
							},
							"rail": map[string]interface{}{
								"type":             "UPI",
								"utr":              "287823703443",
								"upiTransactionId": "YBL85ca838b05c1452a87742f14f987f864",
								"vpa":              "96XXXXXXXX-7@ybl",
							},
							"amount": 100,
						},
					},
				},
			},
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

	status, err := client.GetOrderStatus(context.Background(), "ORDER123", true)

	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", status.State)
	require.Len(t, status.PaymentDetails, 1)

	splitInstruments := status.PaymentDetails[0].SplitInstruments
	require.Len(t, splitInstruments, 1)

	instrument, ok := splitInstruments[0].Instrument.(*paymentinstruments.CreditLinePaymentInstrumentV2)
	require.True(t, ok, "expected instrument to be *CreditLinePaymentInstrumentV2")
	assert.Equal(t, paymentinstruments.CREDIT_LINE, instrument.Type)
	assert.Equal(t, "KARB00CLUPI", instrument.Ifsc)
	assert.Equal(t, "ANIL SASEENDRAN", instrument.AccountHolderName)
	assert.Equal(t, "KBCL", instrument.BankID)
	assert.Equal(t, "XXXXXXXXXXXX2001", instrument.MaskedAccountNumber)
	assert.Equal(t, "CREDITLINE", instrument.ProviderAccountType)

	rail, ok := splitInstruments[0].Rail.(*rails.UpiPaymentRail)
	require.True(t, ok, "expected rail to be *UpiPaymentRail")
	assert.Equal(t, "287823703443", rail.Utr)
}

// Test Refund - 404 Not Found
func TestRefund_404NotFound(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/payments/v2/refund", r.URL.Path)

		errorResponse := map[string]interface{}{
			"code":    "ORDER_NOT_FOUND",
			"message": "Order not found for refund",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
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

	refundRequest := request.NewRefundRequest("REFUND123", "INVALID_ORDER", 5000)

	_, err = client.Refund(context.Background(), refundRequest)

	require.Error(t, err)

	notFoundErr, ok := err.(*exception.ResourceNotFound)
	require.True(t, ok, "Error should be ResourceNotFound")

	require.NotNil(t, notFoundErr.HttpStatusCode)
	assert.Equal(t, 404, *notFoundErr.HttpStatusCode)
	assert.Contains(t, notFoundErr.Message, "not found")
}

// Test Refund Status - 404 Not Found
func TestRefundStatus_404NotFound(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/payments/v2/refund/")
		assert.Contains(t, r.URL.Path, "/status")

		errorResponse := map[string]interface{}{
			"code":    "REFUND_NOT_FOUND",
			"message": "Refund not found",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
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

	_, err = client.GetRefundStatus(context.Background(), "INVALID_REFUND_ID")

	require.Error(t, err)

	notFoundErr, ok := err.(*exception.ResourceNotFound)
	require.True(t, ok, "Error should be ResourceNotFound")

	require.NotNil(t, notFoundErr.HttpStatusCode)
	assert.Equal(t, 404, *notFoundErr.HttpStatusCode)
	assert.Contains(t, notFoundErr.Message, "not found")
}
