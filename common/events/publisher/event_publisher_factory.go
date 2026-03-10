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

package publisher

import (
	"context"
	"net/http"
	"sync"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/events"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/queue"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

type EventPublisherFactory struct {
	HttpClient *http.Client
	HostURL    string
	Logger     logger.Logger
}

var ( // Package-level variable for cached EventPublisher
	cachedEventPublisher EventPublisher
	once                 sync.Once
)

func NewEventPublisherFactory(
	httpClient *http.Client,
	hostURL string,
	log logger.Logger,
) *EventPublisherFactory {
	return &EventPublisherFactory{
		HttpClient: httpClient,
		HostURL:    hostURL,
		Logger:     log,
	}
}

func (epf *EventPublisherFactory) GetEventPublisher(shouldPublishEvents bool) EventPublisher {
	if shouldPublishEvents {
		once.Do(func() {
			cachedEventPublisher = NewQueuedEventPublisher(
				epf.HttpClient,
				queue.NewBoundedConcurrentLinkedQueue(events.QUEUE_MAX_SIZE, epf.Logger),
				epf.HostURL,
				epf.Logger,
			)
		})
		return cachedEventPublisher
	}
	return &NoOpEventPublisher{}
}

type NoOpEventPublisher struct{}

func (nop *NoOpEventPublisher) SetAuthTokenSupplier(authTokenSupplier func(context.Context) (string, error)) {
}
func (nop *NoOpEventPublisher) StartPublishingEvents(authTokenSupplier func(context.Context) (string, error)) {
}
func (nop *NoOpEventPublisher) Send(baseEvent *models.BaseEvent) {}
func (nop *NoOpEventPublisher) Run()                             {}
func (nop *NoOpEventPublisher) Stop()                            {} // No-op: nothing to stop
