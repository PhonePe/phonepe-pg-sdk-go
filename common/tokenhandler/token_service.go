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
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/configs"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models/enums"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/publisher"
	commonHttp "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
)

type TokenService struct {
	HttpClient       *http.Client
	CredentialConfig *configs.CredentialConfig
	Env              types.Env
	oAuthResponse    *OAuthResponse
	EventPublisher   publisher.EventPublisher
	Logger           logger.Logger
	mutex            sync.Mutex
}

func NewTokenService(httpClient *http.Client, credentialConfig *configs.CredentialConfig, env types.Env, eventPublisher publisher.EventPublisher, log logger.Logger) *TokenService {
	ts := &TokenService{
		HttpClient:       httpClient,
		CredentialConfig: credentialConfig,
		Env:              env,
		EventPublisher:   eventPublisher,
		Logger:           log,
	}
	ts.EventPublisher.Send(models.BuildInitClientEventWithEventType(enums.TOKEN_SERVICE_INITIALIZED))
	return ts
}

func (t *TokenService) prepareRequestHeaders() []commonHttp.HttpHeaderPair {
	return []commonHttp.HttpHeaderPair{
		{Key: constants.ContentType, Value: commonHttp.APPLICATION_FORM_URLENCODED},
		{Key: constants.Accept, Value: commonHttp.APPLICATION_JSON},
	}
}

func (t *TokenService) formatCachedToken() string {
	return t.oAuthResponse.TokenType + " " + t.oAuthResponse.AccessToken
}

func (t *TokenService) getCurrentTime() int64 {
	return time.Now().Unix()
}

func (t *TokenService) GetAuthToken(ctx context.Context) (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isCachedTokenValid() {
		t.Logger.Debug("returning cached token")
		return t.formatCachedToken(), nil
	}

	oAuthResponse, err := t.fetchTokenFromPhonePe(ctx)
	if err != nil {
		if t.oAuthResponse == nil {
			t.Logger.Error("no cached token available, error occurred while fetching new token", "error", err)
			return "", err
		}
		t.Logger.Warn("returning cached token, error occurred while fetching new token", "error", err)
		// Publish event for OAuth failure when falling back to cached token
		t.EventPublisher.Send(models.BuildOAuthEvent(
			t.getCurrentTime(),
			"/v1/oauth/token",
			enums.OAUTH_FETCH_FAILED_USED_CACHED_TOKEN,
			err,
			t.oAuthResponse.IssuedAt,
			t.oAuthResponse.ExpiresAt,
		))
		return t.formatCachedToken(), nil
	}

	t.oAuthResponse = oAuthResponse
	return t.formatCachedToken(), nil
}

func (t *TokenService) isCachedTokenValid() bool {
	if t.oAuthResponse == nil {
		return false
	}
	issuedAt := t.oAuthResponse.IssuedAt
	expireAt := t.oAuthResponse.ExpiresAt
	currentTime := t.getCurrentTime()
	reloadTime := issuedAt + (expireAt-issuedAt)/2
	return currentTime < reloadTime
}

func (t *TokenService) ForceRefreshToken() {
	t.Logger.Info("force refreshing token")

	// Lock the entire operation to prevent multiple concurrent refresh requests
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Recover from panics while holding the lock
	defer func() {
		if r := recover(); r != nil {
			t.Logger.Error("recovered from panic during force refresh", "panic", r)
		}
	}()

	// Fetch token while holding the lock using background context
	// Force refresh is typically triggered by 401 errors and shouldn't be cancelled
	// by the original request context
	ctx := context.Background()

	// This ensures only ONE goroutine makes the network call
	oAuthResponse, err := t.fetchTokenFromPhonePe(ctx)
	if err != nil {
		t.Logger.Error("error refreshing token", "error", err)
		return
	}

	// Update the cached token (still holding lock)
	t.oAuthResponse = oAuthResponse
}

func (t *TokenService) fetchTokenFromPhonePe(ctx context.Context) (*OAuthResponse, error) {
	formBody := t.prepareFormBody(t.CredentialConfig)

	httpCommand := commonHttp.NewHttpCommand(
		t.HttpClient,
		t.Env.OAuthHostURL,
		OauthGetToken,
		[]*commonHttp.HttpHeaderPair{
			{Key: constants.ContentType, Value: commonHttp.APPLICATION_FORM_URLENCODED},
			{Key: constants.Accept, Value: commonHttp.APPLICATION_JSON},
		},
		formBody,
		commonHttp.APPLICATION_FORM_URLENCODED,
		commonHttp.POST,
		nil,
		reflect.TypeOf(OAuthResponse{}),
	)

	var oAuthResponse OAuthResponse
	err := httpCommand.Execute(ctx, &oAuthResponse)
	if err != nil {
		return nil, err
	}

	return &oAuthResponse, nil
}

func (t *TokenService) prepareFormBody(credentialConfig *configs.CredentialConfig) url.Values {
	data := url.Values{}
	data.Set("client_id", credentialConfig.ClientId)
	data.Set("client_secret", credentialConfig.ClientSecret)
	data.Set("grant_type", OauthGrantType)
	data.Set("client_version", strconv.Itoa(credentialConfig.ClientVersion))
	return data
}
