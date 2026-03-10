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
	"testing"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/queue"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
	"github.com/stretchr/testify/assert"
)

// TestStopMethod verifies that Stop() properly cleans up resources
func TestStopMethod(t *testing.T) {
	publisher := NewQueuedEventPublisher(
		&http.Client{},
		queue.NewBoundedConcurrentLinkedQueue(100, logger.NewLogger(logger.DisabledLogConfig())),
		"http://localhost:8080",
		logger.NewLogger(logger.DisabledLogConfig()),
	)

	authTokenSupplier := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}

	// Start publishing events
	publisher.StartPublishingEvents(authTokenSupplier)

	// Verify scheduler and quit channel are initialized
	assert.NotNil(t, publisher.scheduler, "Scheduler should be initialized")
	assert.NotNil(t, publisher.quit, "Quit channel should be initialized")

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the publisher
	publisher.Stop()

	// Verify cleanup (channels don't block after Stop)
	// If Stop() didn't work properly, this test would hang or leak goroutines
	time.Sleep(100 * time.Millisecond)

	t.Log("Stop() method successfully cleaned up resources")
}

// TestMultipleStopCalls verifies Stop() can be called multiple times safely
func TestMultipleStopCalls(t *testing.T) {
	publisher := NewQueuedEventPublisher(
		&http.Client{},
		queue.NewBoundedConcurrentLinkedQueue(100, logger.NewLogger(logger.DisabledLogConfig())),
		"http://localhost:8080",
		logger.NewLogger(logger.DisabledLogConfig()),
	)

	authTokenSupplier := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}

	publisher.StartPublishingEvents(authTokenSupplier)
	time.Sleep(100 * time.Millisecond)

	// Call Stop multiple times - should not panic
	publisher.Stop()
	publisher.Stop()
	publisher.Stop()

	t.Log("Multiple Stop() calls handled safely")
}

// TestStopWithoutStart verifies Stop() can be called even if Start was never called
func TestStopWithoutStart(t *testing.T) {
	publisher := NewQueuedEventPublisher(
		&http.Client{},
		queue.NewBoundedConcurrentLinkedQueue(100, logger.NewLogger(logger.DisabledLogConfig())),
		"http://localhost:8080",
		logger.NewLogger(logger.DisabledLogConfig()),
	)

	// Call Stop without starting - should not panic
	publisher.Stop()

	t.Log("Stop() without Start() handled safely")
}

// TestStoppedFlagPreventsDoubleClose verifies the stopped flag prevents double channel close
func TestStoppedFlagPreventsDoubleClose(t *testing.T) {
	publisher := NewQueuedEventPublisher(
		&http.Client{},
		queue.NewBoundedConcurrentLinkedQueue(100, logger.NewLogger(logger.DisabledLogConfig())),
		"http://localhost:8080",
		logger.NewLogger(logger.DisabledLogConfig()),
	)

	authTokenSupplier := func(ctx context.Context) (string, error) {
		return "test-token", nil
	}

	publisher.StartPublishingEvents(authTokenSupplier)
	time.Sleep(100 * time.Millisecond)

	// First stop
	publisher.Stop()
	assert.True(t, publisher.stopped, "stopped flag should be true after first Stop()")

	// Second stop - should return early due to stopped flag
	// This would panic if we tried to close the channel twice
	publisher.Stop()
	assert.True(t, publisher.stopped, "stopped flag should remain true after second Stop()")

	t.Log("Stopped flag correctly prevents double channel close")
}
