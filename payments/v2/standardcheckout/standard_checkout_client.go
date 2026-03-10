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
	"fmt"
	"strconv"

	"github.com/PhonePe/phonepe-pg-sdk-go/common"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models/enums"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	commonRequest "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	commonResponse "github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	response "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/response"
)

type StandardCheckoutClient struct {
	*common.BaseClient
	headers []*http.HttpHeaderPair
}

func (s *StandardCheckoutClient) prepareHeaders() {
	s.headers = []*http.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Source", Value: "INTEGRATION"},
		{Key: "x-source-version", Value: "V2"},
		{Key: "x-Source-Platform", Value: "BACKEND_GO_SDK"},
		{Key: "x-Source-Platform-Version", Value: common.GetSDKVersion()},
	}
}

// GetInstance creates a new StandardCheckoutClient instance without retry
func GetInstance(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool) (*StandardCheckoutClient, error) {
	return GetInstanceWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishEvents, nil)
}

// GetInstanceWithRetry creates a new StandardCheckoutClient instance with custom retry configuration
func GetInstanceWithRetry(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool, retryConfig *http.RetryConfig) (*StandardCheckoutClient, error) {
	// Only publish events in production when explicitly requested
	shouldPublishInProd := shouldPublishEvents && env.PgHostURL == types.Production.PgHostURL

	baseClient, err := common.NewBaseClientWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishInProd, retryConfig)
	if err != nil {
		return nil, err
	}

	instance := &StandardCheckoutClient{
		BaseClient: baseClient,
	}
	instance.prepareHeaders()

	// Send client initialization event
	instance.EventPublisher.Send(models.BuildInitClientEvent(
		enums.PG_CHECKOUT,
		enums.STANDARD_CHECKOUT_CLIENT_INITIALIZED,
	))

	return instance, nil
}

// Pay initiates a standard checkout payment request
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (s *StandardCheckoutClient) Pay(ctx context.Context, payRequest *request.StandardCheckoutPayRequest) (*response.StandardCheckoutPayResponse, error) {
	url := PayApi
	var payResponse response.StandardCheckoutPayResponse

	err := s.RequestViaAuthRefresh(ctx, http.POST, payRequest, url, nil, &payResponse, s.headers)
	if err != nil {
		s.EventPublisher.Send(models.BuildStandardCheckoutPayEventWithError(
			enums.FAILED,
			payRequest,
			url,
			enums.PAY_FAILED,
			err,
		))
		return nil, err
	}
	s.EventPublisher.Send(models.BuildStandardCheckoutPayEvent(
		enums.SUCCESS,
		payRequest,
		url,
		enums.PAY_SUCCESS,
	))
	return &payResponse, nil
}

// GetOrderStatus retrieves the status of an order by merchant order ID
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (s *StandardCheckoutClient) GetOrderStatus(ctx context.Context, merchantOrderId string, details ...bool) (*commonResponse.OrderStatusResponse, error) {
	includeDetails := false
	if len(details) > 0 {
		includeDetails = details[0]
	}

	url := fmt.Sprintf(OrderStatusApi, merchantOrderId)
	queryParams := map[string]string{
		OrderDetails: strconv.FormatBool(includeDetails),
	}

	var statusResponse commonResponse.OrderStatusResponse
	err := s.RequestViaAuthRefresh(ctx, http.GET, nil, url, queryParams, &statusResponse, s.headers)
	if err != nil {
		s.EventPublisher.Send(models.BuildOrderStatusEventWithError(
			enums.FAILED,
			merchantOrderId,
			url,
			enums.PG_CHECKOUT,
			enums.ORDER_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	s.EventPublisher.Send(models.BuildOrderStatusEvent(
		enums.SUCCESS,
		merchantOrderId,
		url,
		enums.PG_CHECKOUT,
		enums.ORDER_STATUS_SUCCESS,
	))
	return &statusResponse, nil
}

// Refund initiates a refund request
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (s *StandardCheckoutClient) Refund(ctx context.Context, refundRequest *commonRequest.RefundRequest) (*commonResponse.RefundResponse, error) {
	url := RefundApi
	var refundResponse commonResponse.RefundResponse

	err := s.RequestViaAuthRefresh(ctx, http.POST, refundRequest, url, nil, &refundResponse, s.headers)
	if err != nil {
		s.EventPublisher.Send(models.BuildRefundEventWithError(
			enums.FAILED,
			refundRequest,
			url,
			enums.PG_CHECKOUT,
			enums.REFUND_FAILED,
			err,
		))
		return nil, err
	}
	s.EventPublisher.Send(models.BuildRefundEvent(
		enums.SUCCESS,
		refundRequest,
		url,
		enums.PG_CHECKOUT,
		enums.REFUND_SUCCESS,
	))
	return &refundResponse, nil
}

// CreateSdkOrder creates an order for mobile SDK integration
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (s *StandardCheckoutClient) CreateSdkOrder(ctx context.Context, orderRequest *request.CreateSdkOrderRequest) (*response.CreateSdkOrderResponse, error) {
	url := CreateOrderApi
	var orderResponse response.CreateSdkOrderResponse

	err := s.RequestViaAuthRefresh(ctx, http.POST, orderRequest, url, nil, &orderResponse, s.headers)
	if err != nil {
		s.EventPublisher.Send(models.BuildCreateSdkOrderEventWithError(
			enums.FAILED,
			orderRequest,
			url,
			enums.PG_CHECKOUT,
			enums.CREATE_SDK_ORDER_FAILED,
			err,
		))
		return nil, err
	}
	s.EventPublisher.Send(models.BuildCreateSdkOrderEvent(
		enums.SUCCESS,
		orderRequest,
		url,
		enums.PG_CHECKOUT,
		enums.CREATE_SDK_ORDER_SUCCESS,
	))
	return &orderResponse, nil
}

// GetTransactionStatus retrieves the status of a transaction by transaction ID
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (s *StandardCheckoutClient) GetTransactionStatus(ctx context.Context, transactionId string) (*commonResponse.OrderStatusResponse, error) {
	url := fmt.Sprintf(TransactionStatusApi, transactionId)
	var statusResponse commonResponse.OrderStatusResponse

	err := s.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &statusResponse, s.headers)
	if err != nil {
		s.EventPublisher.Send(models.BuildTransactionStatusEventWithError(
			enums.FAILED,
			transactionId,
			url,
			enums.PG_CHECKOUT,
			enums.TRANSACTION_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	s.EventPublisher.Send(models.BuildTransactionStatusEvent(
		enums.SUCCESS,
		transactionId,
		url,
		enums.PG_CHECKOUT,
		enums.TRANSACTION_STATUS_SUCCESS,
	))
	return &statusResponse, nil
}

// GetRefundStatus retrieves the status of a refund by refund ID
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (s *StandardCheckoutClient) GetRefundStatus(ctx context.Context, refundId string) (*commonResponse.RefundStatusResponse, error) {
	url := fmt.Sprintf(RefundStatusApi, refundId)
	var refundStatusResponse commonResponse.RefundStatusResponse

	err := s.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &refundStatusResponse, s.headers)
	if err != nil {
		s.EventPublisher.Send(models.BuildRefundStatusEventWithError(
			enums.FAILED,
			refundId,
			url,
			enums.PG_CHECKOUT,
			enums.REFUND_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	s.EventPublisher.Send(models.BuildRefundStatusEvent(
		enums.SUCCESS,
		refundId,
		url,
		enums.PG_CHECKOUT,
		enums.REFUND_STATUS_SUCCESS,
	))
	return &refundStatusResponse, nil
}

func (s *StandardCheckoutClient) ValidateCallback(username string, password string, authorization string, responseBody string) (*commonResponse.CallbackResponse, error) {
	if !common.IsCallbackValid(username, password, authorization) {
		return nil, fmt.Errorf("invalid callback")
	}
	var callbackResponse commonResponse.CallbackResponse
	err := json.Unmarshal([]byte(responseBody), &callbackResponse)
	if err != nil {
		s.EventPublisher.Send(models.BuildCallbackSerializationFailedEvent(
			enums.FAILED,
			enums.PG_CHECKOUT,
			enums.CALLBACK_SERIALIZATION_FAILED,
			err,
		))
		return nil, err
	}
	return &callbackResponse, nil
}
