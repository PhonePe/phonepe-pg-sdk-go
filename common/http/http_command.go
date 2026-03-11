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

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/exception"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/logger"
)

const (
	APPLICATION_JSON            = "application/json"
	APPLICATION_FORM_URLENCODED = "application/x-www-form-urlencoded"
)

type HttpCommand struct {
	Client       *http.Client
	HostURL      string
	URL          string
	Headers      []*HttpHeaderPair
	RequestData  interface{}
	EncodingType string
	MethodName   HttpMethodType
	QueryParams  map[string]string
	ResponseType reflect.Type
	Logger       logger.Logger
}

func (hc *HttpCommand) prepareRequestBody() (io.Reader, error) {
	if hc.RequestData == nil {
		return nil, nil
	}

	switch hc.EncodingType {
	case APPLICATION_JSON:
		body, err := json.Marshal(hc.RequestData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data to JSON: %w", err)
		}
		return bytes.NewBuffer(body), nil
	case APPLICATION_FORM_URLENCODED:
		// Assuming requestData is already url.Values or similar for form encoding
		if formData, ok := hc.RequestData.(url.Values); ok {
			return bytes.NewBufferString(formData.Encode()), nil
		}
		return nil, fmt.Errorf("request data for form-urlencoded must be of type url.Values")
	default:
		return nil, fmt.Errorf("unsupported encoding type: %s", hc.EncodingType)
	}
}

func (hc *HttpCommand) prepareHttpURL() (*url.URL, error) {
	baseURL, err := url.Parse(hc.HostURL + hc.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	queryParams := baseURL.Query()
	for key, value := range hc.QueryParams {
		queryParams.Add(key, value)
	}
	baseURL.RawQuery = queryParams.Encode()

	return baseURL, nil
}

func (hc *HttpCommand) prepareRequest(ctx context.Context, httpURL *url.URL) (*http.Request, error) {
	var req *http.Request
	var err error

	bodyReader, err := hc.prepareRequestBody()
	if err != nil {
		return nil, err
	}

	switch hc.MethodName {
	case GET:
		req, err = http.NewRequestWithContext(ctx, string(GET), httpURL.String(), nil)
	case POST:
		req, err = http.NewRequestWithContext(ctx, string(POST), httpURL.String(), bodyReader)
	case PUT:
		req, err = http.NewRequestWithContext(ctx, string(PUT), httpURL.String(), bodyReader)
	case DELETE:
		req, err = http.NewRequestWithContext(ctx, string(DELETE), httpURL.String(), bodyReader)
	default:
		return nil, exception.NewPhonePeException(fmt.Sprintf("method not supported: %s", hc.MethodName))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	for _, header := range hc.Headers {
		req.Header.Add(header.Key, header.Value)
	}

	if hc.EncodingType != "" {
		req.Header.Set("Content-Type", hc.EncodingType)
	}

	return req, nil
}

func (hc *HttpCommand) handleResponse(response *http.Response, responseObj interface{}) error {
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		if hc.ResponseType == nil || len(responseBody) == 0 {
			return nil
		}
		return json.Unmarshal(responseBody, responseObj)
	} else {
		var phonePeResponse exception.PhonePeResponse
		err := json.Unmarshal(responseBody, &phonePeResponse)
		if err != nil {
			// If we can't unmarshal to PhonePeResponse, return a generic PhonePeException
			return exception.NewPhonePeExceptionWithHttpStatus(response.StatusCode, fmt.Sprintf("HTTP Error %d: %s", response.StatusCode, response.Status))
		}
		// Use the ExceptionMapper to return a specific error type
		return exception.MapError(response.StatusCode, phonePeResponse.Message, &phonePeResponse)
	}
}

func (hc *HttpCommand) Execute(ctx context.Context, responseObj interface{}) error {
	hc.Logger.Debug("calling API", "method", hc.MethodName, "url", hc.HostURL+hc.URL)
	httpURL, err := hc.prepareHttpURL()
	if err != nil {
		return err
	}

	req, err := hc.prepareRequest(ctx, httpURL)
	if err != nil {
		return err
	}

	resp, err := hc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	return hc.handleResponse(resp, responseObj)
}

func NewHttpCommand(
	client *http.Client,
	hostURL string,
	url string,
	headers []*HttpHeaderPair,
	requestData interface{},
	encodingType string,
	methodName HttpMethodType,
	queryParams map[string]string,
	responseType reflect.Type,
	log logger.Logger,
) *HttpCommand {
	return &HttpCommand{
		Client:       client,
		HostURL:      hostURL,
		URL:          url,
		Headers:      headers,
		RequestData:  requestData,
		EncodingType: encodingType,
		MethodName:   methodName,
		QueryParams:  queryParams,
		ResponseType: responseType,
		Logger:       log,
	}
}

// ExecuteWithRetry executes the HTTP command with retry logic based on the provided configuration
// It respects context cancellation and will stop retrying if context is cancelled
func (hc *HttpCommand) ExecuteWithRetry(ctx context.Context, responseObj interface{}, config *RetryConfig) error {
	if config == nil || !config.Enabled {
		return hc.Execute(ctx, responseObj)
	}

	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check if context is already cancelled before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := hc.Execute(ctx, responseObj)

		// Success! No need to retry
		if err == nil {
			if attempt > 0 {
				hc.Logger.Info("request succeeded after retries", "retries", attempt)
			}
			return nil
		}

		// Check if the error is retryable
		if !isRetryableError(err, config.RetryableStatusCodes) {
			// Non-retryable error, fail immediately
			return err
		}

		lastErr = err

		// Don't sleep after the last attempt
		if attempt < config.MaxAttempts-1 {
			delay := calculateBackoff(attempt, config)
			hc.Logger.Debug("request failed, retrying",
				"attempt", attempt+1,
				"max_attempts", config.MaxAttempts,
				"retry_after", delay,
				"error", err)

			// Sleep with context awareness - stop sleeping if context is cancelled
			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// All retries exhausted
	return fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// calculateBackoff calculates the delay before the next retry using exponential backoff
func calculateBackoff(attempt int, config *RetryConfig) time.Duration {
	// Calculate exponential backoff: initialDelay * multiplier^attempt
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	return time.Duration(delay)
}

// isRetryableError checks if an error should trigger a retry
func isRetryableError(err error, retryableStatusCodes []int) bool {
	if err == nil {
		return false
	}

	// Check if error has GetHttpStatusCode method (all PhonePe errors do)
	type httpStatusCodeGetter interface {
		GetHttpStatusCode() *int
	}

	if statusGetter, ok := err.(httpStatusCodeGetter); ok {
		statusCodePtr := statusGetter.GetHttpStatusCode()
		if statusCodePtr != nil {
			statusCode := *statusCodePtr
			for _, code := range retryableStatusCodes {
				if statusCode == code {
					return true
				}
			}
		}
		return false
	}

	// For network errors or other generic errors without status codes, retry
	// (these don't have status codes but might be transient)
	return true
}
