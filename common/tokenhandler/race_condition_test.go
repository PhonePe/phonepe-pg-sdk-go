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
	"sync"
	"testing"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/configs"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
)

// mockEventPublisherForRace for testing
type mockEventPublisherForRace struct{}

func (m *mockEventPublisherForRace) SetAuthTokenSupplier(authTokenSupplier func(context.Context) (string, error)) {
}
func (m *mockEventPublisherForRace) StartPublishingEvents(authTokenSupplier func(context.Context) (string, error)) {
}
func (m *mockEventPublisherForRace) Send(baseEvent *models.BaseEvent) {}
func (m *mockEventPublisherForRace) Run()                             {}
func (m *mockEventPublisherForRace) Stop()                            {}

func TestRaceCondition(t *testing.T) {
	ts := NewTokenService(
		&http.Client{},
		&configs.CredentialConfig{
			ClientId:      "test-id",
			ClientSecret:  "test-secret",
			ClientVersion: 1,
		},
		types.Env{
			OAuthHostURL: "http://localhost:9999",
		},
		&mockEventPublisherForRace{},
		logger.NewLogger(logger.DefaultLogConfig()),
	)

	// Simulate having a token already
	ts.oAuthResponse = &OAuthResponse{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour).Unix(),
		IssuedAt:    time.Now().Unix(),
	}

	var wg sync.WaitGroup

	// Start 10 goroutines calling GetAuthToken (reads oAuthResponse)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts.GetAuthToken(context.Background())
		}()
	}

	// Start 10 goroutines calling ForceRefreshToken (writes oAuthResponse)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ts.ForceRefreshToken()
		}()
	}

	wg.Wait()
	t.Log("Race condition test completed successfully - no races detected!")
}
