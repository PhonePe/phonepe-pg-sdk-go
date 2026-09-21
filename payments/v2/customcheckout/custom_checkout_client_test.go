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
	"strings"
	"testing"

	"github.com/PhonePe/phonepe-pg-sdk-go/common"
	httplib "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	request "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	commonResponse "github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/paymentinstruments"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/rails"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	v2_request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
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
			"issued_at":    1609459200000, // Numeric timestamp
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
	assert.Equal(t, "V2", headerMap["x-source-version"])
	assert.Equal(t, "BACKEND_GO_SDK", headerMap["x-Source-Platform"])
	assert.Equal(t, common.GetSDKVersion(), headerMap["x-Source-Platform-Version"])
}

// Test Pay method
func TestPay_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/payments/v2/pay", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId":     "ORDER123",
			"state":       "PENDING",
			"redirectUrl": "https://payment.phonepe.com/xyz",
			"expireAt":    1234567890,
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

	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER123",
		Amount:          10000,
	}

	response, err := client.Pay(context.Background(), payRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.OrderID)
}

func TestPay_Error(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"code":    "BAD_REQUEST",
			"message": "Invalid request",
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

	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER123",
		Amount:          10000,
	}

	response, err := client.Pay(context.Background(), payRequest)

	assert.Error(t, err)
	assert.Nil(t, response)
}

// Test Pay method - Invalid MetaInfo should be rejected before the HTTP call is made
func TestPay_InvalidMetaInfo_ReturnsValidationError(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
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

	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER123",
		Amount:          10000,
		MetaInfo:        models.MetaInfo{Udf12: "invalid#value"},
	}

	response, err := client.Pay(context.Background(), payRequest)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "udf12")
	assert.False(t, apiCalled, "API should not be called when MetaInfo validation fails")
}

// Test Pay method - Valid MetaInfo should pass validation and succeed
func TestPay_ValidMetaInfo_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId": "ORDER123",
			"state":   "PENDING",
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

	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER123",
		Amount:          10000,
		MetaInfo:        models.MetaInfo{Udf1: "some-value", Udf12: "valid_value-1"},
	}

	response, err := client.Pay(context.Background(), payRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.OrderID)
}

// Test CreateSdkOrder - Invalid MetaInfo should be rejected before the HTTP call is made
func TestCreateSdkOrder_InvalidMetaInfo_ReturnsValidationError(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
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

	orderRequest := &v2_request.CreateSdkOrderRequest{
		MerchantOrderID: "ORDER123",
		Amount:          10000,
		MetaInfo:        models.MetaInfo{Udf1: strings.Repeat("a", 257)}, // exceeds 256 char max
	}

	response, err := client.CreateSdkOrder(context.Background(), orderRequest)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "udf1")
	assert.False(t, apiCalled, "API should not be called when MetaInfo validation fails")
}

// Test GetOrderStatus method
func TestGetOrderStatus_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/payments/v2/order/ORDER123/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("details"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"code":    "SUCCESS",
			"message": "Order status retrieved",
			"data": map[string]interface{}{
				"merchantOrderId": "ORDER123",
				"transactionId":   "TXN123",
				"state":           "COMPLETED",
			},
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

	response, err := client.GetOrderStatus(context.Background(), "ORDER123", true)

	require.NoError(t, err)
	assert.NotNil(t, response)
}

func TestGetOrderStatus_WithoutDetails(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "false", r.URL.Query().Get("details"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"code":    "SUCCESS",
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

// Test GetOrderStatus - Credit Line instrument within splitInstruments
func TestGetOrderStatus_CreditLineInstrument(t *testing.T) {
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
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
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

// Test GetTransactionStatus method
func TestGetTransactionStatus_Success(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/payments/v2/transaction/TXN123/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"code":    "SUCCESS",
			"data": map[string]interface{}{
				"transactionId": "TXN123",
				"state":         "COMPLETED",
			},
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
			"success": true,
			"code":    "SUCCESS",
			"message": "Refund initiated",
			"data": map[string]interface{}{
				"merchantRefundId": "REFUND123",
				"transactionId":    "TXN123",
			},
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
			"success": true,
			"code":    "SUCCESS",
			"data": map[string]interface{}{
				"merchantRefundId": "REFUND123",
				"state":            "COMPLETED",
			},
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
		"success": true,
		"code":    "PAYMENT_SUCCESS",
		"data": map[string]interface{}{
			"merchantOrderId": "ORDER123",
			"transactionId":   "TXN123",
			"state":           "COMPLETED",
		},
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

// Test ValidateCallback with Credit Line instrument in splitInstruments
func TestValidateCallback_CreditLineInstrument(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.phonepe.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.phonepe.com",
	}

	client, err := GetInstance("test-client-id", "test-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := map[string]interface{}{
		"data": map[string]interface{}{
			"merchantId":      "PRODTEST",
			"merchantOrderId": "28D43E2BAD6411EB85B692E8A924135",
			"orderId":         "OMO2607151547579376222138V",
			"state":           "COMPLETED",
			"amount":          100,
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
		},
	}
	callbackJSON, _ := json.Marshal(callbackData)

	// SHA256 of "test-user:test-pass"
	authHash := "75f05a3e2a10a7488c40eac6ef38da95df902e8fc3886531ce97ee8599481584"

	result, err := client.ValidateCallback("test-user", "test-pass", authHash, string(callbackJSON))

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Data.PaymentDetails, 1)

	splitInstruments := result.Data.PaymentDetails[0].SplitInstruments
	require.Len(t, splitInstruments, 1)

	instrument, ok := splitInstruments[0].Instrument.(*paymentinstruments.CreditLinePaymentInstrumentV2)
	require.True(t, ok, "expected instrument to be *CreditLinePaymentInstrumentV2")
	assert.Equal(t, "KBCL", instrument.BankID)
	assert.Equal(t, "XXXXXXXXXXXX2001", instrument.MaskedAccountNumber)
	assert.Equal(t, "CREDITLINE", instrument.ProviderAccountType)
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

	callbackJSON := `{"success": true}`

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

	// Invalid JSON
	response, err := client.ValidateCallback("test-user", "test-pass", "Basic dGVzdC11c2VyOnRlc3QtcGFzcw==", "invalid-json{")

	assert.Error(t, err)
	assert.Nil(t, response)
}

// Test retry behavior with different configurations
func TestPay_WithRetry_SuccessAfterFailure(t *testing.T) {
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
				"message": "Service unavailable",
			})
			return
		}
		// Second call succeeds
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"code":    "SUCCESS",
			"data": map[string]interface{}{
				"merchantOrderId": "ORDER123",
			},
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

	payRequest := &request.PgPaymentRequest{
		MerchantOrderID: "ORDER123",
		Amount:          10000,
	}

	response, err := client.Pay(context.Background(), payRequest)

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, 2, callCount, "Should have retried once")
}

// Test x-device-os header behavior in Pay
func TestPayUpiCollectWithXDeviceOsHeader(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedDeviceOs string
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedDeviceOs = r.Header.Get("x-device-os")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId": "ORDER123",
			"state":   "PENDING",
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

	payRequest := request.NewUpiCollectPayViaVpaRequest(
		10000, "ORDER123", models.MetaInfo{}, nil, "test@upi", "Pay now", 300, "ANDROID",
	)

	_, err = client.Pay(context.Background(), payRequest)
	require.NoError(t, err)
	assert.Equal(t, "ANDROID", capturedDeviceOs)
}

func TestPayUpiCollectWithoutXDeviceOsHeader(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	var capturedDeviceOs string
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedDeviceOs = r.Header.Get("x-device-os")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orderId": "ORDER123",
			"state":   "PENDING",
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

	payRequest := request.NewUpiCollectPayViaVpaRequest(
		10000, "ORDER123", models.MetaInfo{}, nil, "test@upi", "Pay now", 300,
	)

	_, err = client.Pay(context.Background(), payRequest)
	require.NoError(t, err)
	assert.Equal(t, "", capturedDeviceOs)
}

func TestPayUpiCollectDeviceOSNotInRequestBody(t *testing.T) {
	payRequest := request.NewUpiCollectPayViaVpaRequest(
		10000, "ORDER123", models.MetaInfo{}, nil, "test@upi", "Pay now", 300, "ANDROID",
	)

	body, err := json.Marshal(payRequest)
	require.NoError(t, err)
	assert.NotContains(t, string(body), "deviceOS")
	assert.NotContains(t, string(body), "deviceOS")
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
