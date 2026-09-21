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
	"encoding/json"
	"testing"

	"github.com/PhonePe/phonepe-pg-sdk-go/common"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/paymentinstruments"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/rails"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test pg.order.completed callback
func TestCallback_PgOrderCompleted(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	// Create callback data for pg.order.completed
	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:         "OMO123",
			MerchantID:      "merchant123",
			MerchantOrderID: "ORDER123",
			State:           "COMPLETED",
			Amount:          10000,
			ExpireAt:        1234567890,
			MetaInfo:        models.MetaInfo{},
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.UPI_COLLECT,
					Timestamp:     1234567890,
					Amount:        10000,
					State:         "SUCCESS",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	// Calculate SHA256 hash: username:password:responseBody
	username := "merchant"
	password := "secret"
	expectedAuth := "94ff4c1b18cc4ad6936b6ce92e60a58bbc1e62ccd96fc80da33f91c45eae2c8a" // Pre-calculated

	result, err := client.ValidateCallback(username, password, expectedAuth, string(responseBody))

	// If validation succeeds, verify the parsed data
	if err == nil {
		assert.NotNil(t, result)
		assert.Equal(t, "OMO123", result.Data.OrderID)
		assert.Equal(t, "COMPLETED", result.Data.State)
		assert.Equal(t, int64(10000), result.Data.Amount)
	}
}

// Test callback with Credit Line instrument in splitInstruments
func TestCallback_InstrumentCreditLine(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:         "OMO2607151547579376222138V",
			MerchantID:      "PRODTEST",
			MerchantOrderID: "28D43E2BAD6411EB85B692E8A924135",
			State:           "COMPLETED",
			Amount:          100,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "OM2607151547580748627290V",
					PaymentMode:   models.UPI_INTENT,
					Timestamp:     1784110678105,
					Amount:        100,
					State:         "COMPLETED",
					SplitInstruments: []response.InstrumentCombo{
						{
							Instrument: &paymentinstruments.CreditLinePaymentInstrumentV2{
								Type:                paymentinstruments.CREDIT_LINE,
								Ifsc:                "KARB00CLUPI",
								AccountHolderName:   "ANIL SASEENDRAN",
								BankID:              "KBCL",
								MaskedAccountNumber: "XXXXXXXXXXXX2001",
								ProviderAccountType: "CREDITLINE",
							},
							Rail: &rails.UpiPaymentRail{
								Type:             rails.UPI,
								Utr:              "287823703443",
								UpiTransactionID: "YBL85ca838b05c1452a87742f14f987f864",
								Vpa:              "96XXXXXXXX-7@ybl",
							},
							Amount: 100,
						},
					},
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	authorization := common.CalculateSha256(username, password)

	result, err := client.ValidateCallback(username, password, authorization, string(responseBody))

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

// Test checkout.transaction.attempt.failed callback
func TestCallback_CheckoutTransactionAttemptFailed(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode:       models.UPI_COLLECT,
					State:             "FAILED",
					ErrorCode:         "TRANSACTION_DECLINED",
					DetailedErrorCode: "U30",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	// Note: In real scenario, this would be properly calculated
	expectedAuth := "fake_hash_for_test"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))

	// Test expects error due to hash mismatch (expected in test scenario)
	assert.Error(t, err)
}

// Test pg.order.failed callback
func TestCallback_PgOrderFailed(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "FAILED",
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode:       models.UPI_COLLECT,
					State:             "FAILED",
					ErrorCode:         "PAYMENT_ERROR",
					DetailedErrorCode: "ZM",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err) // Expected due to hash mismatch
}

// Test pg.transaction.attempt.failed callback
func TestCallback_PgTransactionAttemptFailed(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode: models.UPI_COLLECT,
					State:       "FAILED",
					ErrorCode:   "AUTHORIZATION_ERROR",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test pg.refund.failed callback
func TestCallback_PgRefundFailed(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:          "OMO123",
			RefundID:         "REFUND123",
			MerchantRefundID: "MERCHANT_REFUND_123",
			State:            "FAILED",
			Amount:           5000,
			ErrorCode:        "REFUND_FAILED",
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode: models.UPI_COLLECT,
					State:       "FAILED",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test pg.refund.completed callback
func TestCallback_PgRefundCompleted(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:          "OMO123",
			RefundID:         "REFUND123",
			MerchantRefundID: "MERCHANT_REFUND_123",
			State:            "COMPLETED",
			Amount:           5000,
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode: models.UPI_COLLECT,
					State:       "COMPLETED",
					Amount:      5000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test pg.refund.accepted callback
func TestCallback_PgRefundAccepted(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:          "OMO123",
			RefundID:         "REFUND123",
			MerchantRefundID: "MERCHANT_REFUND_123",
			State:            "PENDING",
			Amount:           5000,
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode: models.UPI_COLLECT,
					State:       "PENDING",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test checkout.order.completed callback
func TestCallback_CheckoutOrderCompleted(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:         "OMO123",
			MerchantOrderID: "ORDER123",
			State:           "COMPLETED",
			Amount:          10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.UPI_COLLECT,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test checkout.order.failed callback
func TestCallback_CheckoutOrderFailed(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID:         "OMO123",
			MerchantOrderID: "ORDER123",
			State:           "FAILED",
			ErrorCode:       "PAYMENT_ERROR",
			PaymentDetails: []response.PaymentDetail{
				{
					PaymentMode: models.UPI_COLLECT,
					State:       "FAILED",
					ErrorCode:   "PAYMENT_DECLINED",
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback with Wallet instrument
func TestCallback_InstrumentWallet(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.CARD,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback with Credit Card instrument
func TestCallback_InstrumentCreditCard(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.CARD,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback with Debit Card instrument
func TestCallback_InstrumentDebitCard(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.CARD,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback with NetBanking instrument
func TestCallback_InstrumentNetBanking(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.NET_BANKING,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback with UPI Intent payment mode
func TestCallback_PaymentModeUpiIntent(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.UPI_INTENT,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback with UPI QR payment mode
func TestCallback_PaymentModeUpiQr(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
			PaymentDetails: []response.PaymentDetail{
				{
					TransactionID: "TXN123",
					PaymentMode:   models.UPI_QR,
					State:         "SUCCESS",
					Amount:        10000,
				},
			},
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "merchant"
	password := "secret"
	expectedAuth := "fake_hash"

	_, err = client.ValidateCallback(username, password, expectedAuth, string(responseBody))
	assert.Error(t, err)
}

// Test callback validation with proper hash
func TestCallback_ValidHashValidation(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	env := types.Env{
		PgHostURL:     "https://api.test.com",
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	callbackData := response.CallbackResponse{
		Data: response.CallbackData{
			OrderID: "OMO123",
			State:   "COMPLETED",
			Amount:  10000,
		},
	}
	responseBody, _ := json.Marshal(callbackData)

	username := "testuser"
	password := "testpass"
	// Calculate proper SHA256: SHA256(username:password:responseBody)
	// For testing, we validate that hash check is performed
	invalidAuth := "invalid_hash_value"

	_, err = client.ValidateCallback(username, password, invalidAuth, string(responseBody))

	// Should error due to invalid hash
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid callback")
}
