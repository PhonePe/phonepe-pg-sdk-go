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

package queue

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

type BoundedConcurrentLinkedQueue struct {
	queue   chan *models.BaseEvent
	maxSize int
	Logger  logger.Logger
}

func NewBoundedConcurrentLinkedQueue(maxSize int, log logger.Logger) *BoundedConcurrentLinkedQueue {
	return &BoundedConcurrentLinkedQueue{
		queue:   make(chan *models.BaseEvent, maxSize),
		maxSize: maxSize,
		Logger:  log,
	}
}

func (b *BoundedConcurrentLinkedQueue) Add(data *models.BaseEvent) {
	if data == nil {
		return
	}
	select {
	case b.queue <- data:
	default:
		b.Logger.Warn("reached queue max size, skipping event", "event_name", data.EventName, "max_size", b.maxSize)
	}
}

func (b *BoundedConcurrentLinkedQueue) IsEmpty() bool {
	return len(b.queue) == 0
}

func (b *BoundedConcurrentLinkedQueue) Size() int {
	return len(b.queue)
}

func (b *BoundedConcurrentLinkedQueue) Poll() *models.BaseEvent {
	select {
	case data := <-b.queue:
		return data
	default:
		return nil
	}
}
