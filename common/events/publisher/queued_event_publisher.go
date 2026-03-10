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
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/queue"
	commonHttp "github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

type QueuedEventPublisher struct {
	HttpClient        *http.Client
	EventQueue        queue.EventQueue
	HostURL           string
	AuthTokenSupplier func(context.Context) (string, error)
	Logger            logger.Logger
	scheduler         *time.Ticker
	quit              chan struct{}
	mu                sync.Mutex
	stopped           bool
}

func NewQueuedEventPublisher(
	httpClient *http.Client,
	eventQueue queue.EventQueue,
	hostURL string,
	log logger.Logger,
) *QueuedEventPublisher {
	return &QueuedEventPublisher{
		HttpClient: httpClient,
		EventQueue: eventQueue,
		HostURL:    hostURL,
		Logger:     log,
	}
}

func (qep *QueuedEventPublisher) SetAuthTokenSupplier(authTokenSupplier func(context.Context) (string, error)) {
	qep.AuthTokenSupplier = authTokenSupplier
}

func (qep *QueuedEventPublisher) Send(baseEvent *models.BaseEvent) {
	qep.EventQueue.Add(baseEvent)
}

func (qep *QueuedEventPublisher) StartPublishingEvents(authTokenSupplier func(context.Context) (string, error)) {
	qep.SetAuthTokenSupplier(authTokenSupplier)
	if qep.scheduler == nil {
		qep.scheduler = time.NewTicker(events.DELAY * time.Second)
		qep.quit = make(chan struct{})
		go func() {
			for {
				select {
				case <-qep.scheduler.C:
					qep.Run()
				case <-qep.quit:
					return
				}
			}
		}()
	}
}

// Stop gracefully stops the event publisher and its background goroutine
func (qep *QueuedEventPublisher) Stop() {
	qep.mu.Lock()
	defer qep.mu.Unlock()

	if qep.stopped {
		return // Already stopped
	}

	if qep.scheduler != nil {
		qep.scheduler.Stop()
	}
	if qep.quit != nil {
		close(qep.quit)
	}
	qep.stopped = true
}

func (qep *QueuedEventPublisher) Run() {
	qep.sendBatchData()
}

func (qep *QueuedEventPublisher) sendBatchData() {
	defer func() {
		if r := recover(); r != nil {
			qep.Logger.Error("recovered from panic in sendBatchData", "panic", r)
		}
	}()

	if qep.EventQueue.IsEmpty() {
		return
	}
	qep.Logger.Debug("processing event queue", "queue_size", qep.EventQueue.Size())

	bulkEventBatch := qep.createEventBatches()

	for _, sdkEventList := range bulkEventBatch {
		err := qep.sendSingleBatchData(sdkEventList)
		if err != nil {
			qep.Logger.Warn("error occurred sending events batch to backend", "error", err)
		}
	}
}

func (qep *QueuedEventPublisher) createEventBatches() [][]*models.BaseEvent {
	curQueueSize := qep.EventQueue.Size()
	bulkEventBatch := make([][]*models.BaseEvent, 0)
	currentBatch := make([]*models.BaseEvent, 0, events.MAX_EVENTS_IN_BATCH)

	for numEventsProcessed := 0; numEventsProcessed < curQueueSize; numEventsProcessed++ {
		event := qep.EventQueue.Poll()
		if event == nil {
			break
		}
		currentBatch = append(currentBatch, event)
		if len(currentBatch) == events.MAX_EVENTS_IN_BATCH {
			bulkEventBatch = append(bulkEventBatch, currentBatch)
			currentBatch = make([]*models.BaseEvent, 0, events.MAX_EVENTS_IN_BATCH)
		}
	}
	if len(currentBatch) > 0 {
		bulkEventBatch = append(bulkEventBatch, currentBatch)
	}

	return bulkEventBatch
}

func (qep *QueuedEventPublisher) getHeaders() []*commonHttp.HttpHeaderPair {
	headers := make([]*commonHttp.HttpHeaderPair, 0)
	headers = append(headers, &commonHttp.HttpHeaderPair{Key: constants.Accept, Value: commonHttp.APPLICATION_JSON})
	if qep.AuthTokenSupplier != nil {
		// Use background context for event publishing - it's async and shouldn't be cancelled
		ctx := context.Background()
		token, err := qep.AuthTokenSupplier(ctx)
		if err == nil {
			headers = append(headers, &commonHttp.HttpHeaderPair{Key: events.AUTHORIZATION, Value: token})
		}
	}
	headers = append(headers, &commonHttp.HttpHeaderPair{Key: constants.ContentType, Value: commonHttp.APPLICATION_JSON})
	return headers
}

func (qep *QueuedEventPublisher) sendSingleBatchData(sdkEventList []*models.BaseEvent) error {
	bulkEvent := models.NewBulkEvent(sdkEventList)
	headers := qep.getHeaders()
	httpCommand := qep.buildHttpCommand(headers, bulkEvent)

	// Use background context for event publishing - it runs async in background
	ctx := context.Background()
	var response interface{}
	return httpCommand.Execute(ctx, &response)
}

func (qep *QueuedEventPublisher) buildHttpCommand(
	headers []*commonHttp.HttpHeaderPair, bulkEvent *models.BulkEvent) *commonHttp.HttpCommand {
	return commonHttp.NewHttpCommand(
		qep.HttpClient,
		qep.HostURL,
		events.EVENTS_ENDPOINT,
		headers,
		bulkEvent,
		commonHttp.APPLICATION_JSON,
		commonHttp.POST,
		nil, // No query params for this endpoint
		nil, // Response type will be handled generically
		qep.Logger,
	)
}
