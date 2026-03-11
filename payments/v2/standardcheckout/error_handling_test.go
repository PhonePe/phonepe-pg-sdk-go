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
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test R999 - Service Unavailable Error
func TestException_R999_ServiceUnavailable(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/checkout/v2/pay", r.URL.Path)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("R999"))
	}))
	defer apiServer.Close()

	env := types.Env{
		PgHostURL:     apiServer.URL,
		OAuthHostURL:  oauthServer.URL,
		EventsHostURL: "https://events.test.com",
	}

	client, err := GetInstance("client-id", "client-secret", 1, env, false)
	require.NoError(t, err)

	redirectURL := "https://merchant.com/callback"
	metaInfo := models.MetaInfo{}
	message := "Payment for order"
	payRequest := request.NewStandardCheckoutPayRequest(
		"ORDER123",
		10000,
		&redirectURL,
		&metaInfo,
		&message,
		nil,
		nil,
		nil,
		nil,
	)

	_, err = client.Pay(context.Background(), payRequest)

	require.Error(t, err)

	phonePeErr, ok := err.(*exception.PhonePeException)
	require.True(t, ok, "Error should be of type ServerError")

	require.NotNil(t, phonePeErr.HttpStatusCode)
	assert.Equal(t, 503, *phonePeErr.HttpStatusCode)
	assert.Contains(t, phonePeErr.Message, "Service Unavailable")
}

// Test Bad Request with Error Code
func TestException_BadRequest_WithErrorCode(t *testing.T) {
	oauthServer := createMockOAuthServer()
	defer oauthServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/checkout/v2/pay", r.URL.Path)

		errorResponse := map[string]interface{}{
			"errorCode": "OIM000",
			"code":      "INVALID_CLIENT",
			"message":   "Bad Request: Invalid Client, trackingId: 2123d",
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

	redirectURL2 := "https://merchant.com/callback"
	metaInfo2 := models.MetaInfo{}
	message2 := "Payment for order"
	payRequest := request.NewStandardCheckoutPayRequest(
		"ORDER123",
		10000,
		&redirectURL2,
		&metaInfo2,
		&message2,
		nil,
		nil,
		nil,
		nil,
	)

	_, err = client.Pay(context.Background(), payRequest)

	require.Error(t, err)

	badRequestErr, ok := err.(*exception.BadRequest)
	require.True(t, ok, "Error should be of type BadRequest")

	require.NotNil(t, badRequestErr.HttpStatusCode)
	assert.Equal(t, 400, *badRequestErr.HttpStatusCode)
	assert.Contains(t, badRequestErr.Message, "Bad Request")
	assert.Equal(t, "OIM000", badRequestErr.Code)
}
