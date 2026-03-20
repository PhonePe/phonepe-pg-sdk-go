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
	"errors"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/configs"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/publisher"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/exception"
	commonHttp "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/tokenhandler"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
)

type BaseClient struct {
	HttpClient            *http.Client
	Env                   types.Env
	TokenService          *tokenhandler.TokenService
	CredentialConfig      *configs.CredentialConfig
	EventPublisherFactory *publisher.EventPublisherFactory
	EventPublisher        publisher.EventPublisher
	ShouldPublishEvents   bool
	RetryConfig           *commonHttp.RetryConfig
	Logger                logger.Logger
}

func NewBaseClient(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool) (*BaseClient, error) {
	return NewBaseClientWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishEvents, nil)
}

// NewBaseClientWithRetry creates a new BaseClient with custom retry configuration
func NewBaseClientWithRetry(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool, retryConfig *commonHttp.RetryConfig) (*BaseClient, error) {
	return NewBaseClientWithOptions(clientId, clientSecret, clientVersion, env, shouldPublishEvents, retryConfig, nil)
}

// NewBaseClientWithOptions creates a new BaseClient with custom retry configuration and logger
func NewBaseClientWithOptions(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool, retryConfig *commonHttp.RetryConfig, logConfig *logger.LogConfig) (*BaseClient, error) {
	credentialConfig := configs.NewCredentialConfig(clientId, clientSecret, clientVersion)
	httpClient := createHTTPClientWithTimeouts()

	// Initialize logger - use provided config or default
	var sdkLogger logger.Logger
	if logConfig != nil {
		sdkLogger = logger.NewLogger(logConfig)
	} else {
		sdkLogger = logger.NewLogger(logger.DefaultLogConfig())
	}

	eventPublisherFactory := publisher.NewEventPublisherFactory(httpClient, env.EventsHostURL, sdkLogger)
	eventPublisher := eventPublisherFactory.GetEventPublisher(shouldPublishEvents)

	tokenService := tokenhandler.NewTokenService(httpClient, credentialConfig, env, eventPublisher, sdkLogger)
	eventPublisher.StartPublishingEvents(tokenService.GetAuthToken)

	return &BaseClient{
		HttpClient:            httpClient,
		Env:                   env,
		TokenService:          tokenService,
		CredentialConfig:      credentialConfig,
		EventPublisherFactory: eventPublisherFactory,
		EventPublisher:        eventPublisher,
		ShouldPublishEvents:   shouldPublishEvents,
		RetryConfig:           retryConfig,
		Logger:                sdkLogger,
	}, nil
}

// createHTTPClientWithTimeouts creates an HTTP client with appropriate timeout configurations
func createHTTPClientWithTimeouts() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second, // Overall request timeout
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,  // Connection timeout
				KeepAlive: 30 * time.Second, // Keep-alive probe interval
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second, // TLS handshake timeout
			ResponseHeaderTimeout: 10 * time.Second, // Time to receive response headers
			ExpectContinueTimeout: 1 * time.Second,  // 100-continue timeout
			MaxIdleConns:          100,              // Maximum idle connections
			IdleConnTimeout:       90 * time.Second, // Idle connection timeout
			MaxIdleConnsPerHost:   10,               // Max idle connections per host
		},
	}
}

// RequestViaAuthRefresh makes an HTTP request with automatic token refresh on unauthorized access
// and optional retry logic based on the client's retry configuration
// responseObj should be a pointer to the expected response type
// Context can be used for cancellation, timeouts, and request-scoped values
func (bc *BaseClient) RequestViaAuthRefresh(ctx context.Context, methodName commonHttp.HttpMethodType, requestData interface{}, url string, queryParams map[string]string, responseObj interface{}, headers []*commonHttp.HttpHeaderPair) error {
	return bc.RequestViaAuthRefreshWithHost(ctx, methodName, requestData, url, queryParams, responseObj, headers, bc.Env.PgHostURL)
}

// RequestViaAuthRefreshWithHost is like RequestViaAuthRefresh but uses an explicit hostURL instead of the default PgHostURL.
// Use this when a specific endpoint (e.g. PCI) requires a different base host.
func (bc *BaseClient) RequestViaAuthRefreshWithHost(ctx context.Context, methodName commonHttp.HttpMethodType, requestData interface{}, url string, queryParams map[string]string, responseObj interface{}, headers []*commonHttp.HttpHeaderPair, hostURL string) error {
	httpHeaders := make([]*commonHttp.HttpHeaderPair, len(headers))
	copy(httpHeaders, headers)

	// Add auth header - fail fast if token fetch fails
	httpHeaders, err := bc.AddAuthHeader(ctx, httpHeaders)
	if err != nil {
		// Token fetch failed - don't proceed with API call
		// This is different from an expired token (401 from API)
		return err
	}

	// Get response type from responseObj if it's not nil
	var responseType reflect.Type
	if responseObj != nil {
		responseType = reflect.TypeOf(responseObj).Elem() // Get element type (deref pointer)
	}

	httpCommand := commonHttp.NewHttpCommand(
		bc.HttpClient,
		hostURL,
		url,
		httpHeaders,
		requestData,
		commonHttp.APPLICATION_JSON,
		methodName,
		queryParams,
		responseType,
		bc.Logger,
	)

	// Use retry logic if configured, otherwise execute directly
	if bc.RetryConfig != nil && bc.RetryConfig.Enabled {
		err = httpCommand.ExecuteWithRetry(ctx, responseObj, bc.RetryConfig)
	} else {
		err = httpCommand.Execute(ctx, responseObj)
	}

	if err != nil {
		// Use idiomatic Go error handling with errors.As
		var authErr *exception.UnauthorizedAccess
		if errors.As(err, &authErr) {
			// API returned 401 - token likely expired, force refresh for next call
			bc.TokenService.ForceRefreshToken()
		}
		return err
	}

	return nil
}

// Close gracefully shuts down the BaseClient and releases all resources
func (bc *BaseClient) Close() {
	if bc.EventPublisher != nil {
		bc.EventPublisher.Stop()
	}
}

func (bc *BaseClient) AddAuthHeader(ctx context.Context, headers []*commonHttp.HttpHeaderPair) ([]*commonHttp.HttpHeaderPair, error) {
	authToken, err := bc.TokenService.GetAuthToken(ctx)
	if err != nil {
		// Token fetch failed - return error immediately
		// Don't proceed with API call as it will fail with 401
		// This is different from an expired token scenario
		return headers, err
	}

	headers = append(headers, &commonHttp.HttpHeaderPair{
		Key:   constants.OAuthAuthorization,
		Value: authToken,
	})
	return headers, nil
}

// GetObjectMapper returns a basic JSON encoder/decoder (Go's encoding/json package)
func (bc *BaseClient) GetObjectMapper() *json.Encoder {
	// In Go, we don't have a single ObjectMapper like Jackson
	// Instead, we use json.Marshal/Unmarshal directly
	// This method is kept for interface compatibility but may not be needed
	return nil
}
