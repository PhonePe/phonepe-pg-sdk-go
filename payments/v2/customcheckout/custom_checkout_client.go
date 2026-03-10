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

package customcheckout

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/PhonePe/phonepe-pg-sdk-go/common"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models/enums"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	request "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	commonResponse "github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	v2_request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	v2_response "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/response"
)

type CustomCheckoutClient struct {
	*common.BaseClient
	headers []*http.HttpHeaderPair
}

func (c *CustomCheckoutClient) prepareHeaders() {
	c.headers = []*http.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Source", Value: "INTEGRATION"},
		{Key: "x-source-version", Value: "V2"},
		{Key: "x-Source-Platform", Value: "BACKEND_GO_SDK"},
		{Key: "x-Source-Platform-Version", Value: common.GetSDKVersion()},
	}
}

// GetInstance creates a new CustomCheckoutClient instance without retry
func GetInstance(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool) (*CustomCheckoutClient, error) {
	return GetInstanceWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishEvents, nil)
}

// GetInstanceWithRetry creates a new CustomCheckoutClient instance with custom retry configuration
func GetInstanceWithRetry(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool, retryConfig *http.RetryConfig) (*CustomCheckoutClient, error) {
	shouldPublishInProd := shouldPublishEvents && env.PgHostURL == types.Production.PgHostURL

	baseClient, err := common.NewBaseClientWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishInProd, retryConfig)
	if err != nil {
		return nil, err
	}

	instance := &CustomCheckoutClient{
		BaseClient: baseClient,
	}
	instance.prepareHeaders()

	// Send client initialization event
	instance.EventPublisher.Send(models.BuildInitClientEvent(
		enums.PG,
		enums.CUSTOM_CHECKOUT_CLIENT_INITIALIZED,
	))

	return instance, nil
}

// Pay initiates a custom checkout payment request
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (c *CustomCheckoutClient) Pay(ctx context.Context, payRequest *request.PgPaymentRequest) (*commonResponse.PgPaymentResponse, error) {
	url := PayApi
	var payResponse commonResponse.PgPaymentResponse

	requestHeaders := c.headers
	if payRequest.DeviceOS != "" {
		requestHeaders = make([]*http.HttpHeaderPair, len(c.headers))
		copy(requestHeaders, c.headers)
		requestHeaders = append(requestHeaders, &http.HttpHeaderPair{Key: "x-device-os", Value: payRequest.DeviceOS})
	}

	err := c.RequestViaAuthRefresh(ctx, http.POST, payRequest, url, nil, &payResponse, requestHeaders)
	if err != nil {
		c.EventPublisher.Send(models.BuildCustomCheckoutPayEventWithError(
			enums.FAILED,
			payRequest,
			url,
			enums.PAY_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildCustomCheckoutPayEvent(
		enums.SUCCESS,
		payRequest,
		url,
		enums.PAY_SUCCESS,
	))
	return &payResponse, nil
}

// CreateSdkOrder creates an order for mobile SDK integration
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (c *CustomCheckoutClient) CreateSdkOrder(ctx context.Context, orderRequest *v2_request.CreateSdkOrderRequest) (*v2_response.CreateSdkOrderResponse, error) {
	url := CreateOrderApi
	var orderResponse v2_response.CreateSdkOrderResponse

	err := c.RequestViaAuthRefresh(ctx, http.POST, orderRequest, url, nil, &orderResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildCreateSdkOrderEventWithError(
			enums.FAILED,
			orderRequest,
			url,
			enums.PG,
			enums.CREATE_SDK_ORDER_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildCreateSdkOrderEvent(
		enums.SUCCESS,
		orderRequest,
		url,
		enums.PG,
		enums.CREATE_SDK_ORDER_SUCCESS,
	))
	return &orderResponse, nil
}

// GetOrderStatus retrieves the status of an order by merchant order ID
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (c *CustomCheckoutClient) GetOrderStatus(ctx context.Context, merchantOrderId string, details ...bool) (*commonResponse.OrderStatusResponse, error) {
	includeDetails := false
	if len(details) > 0 {
		includeDetails = details[0]
	}

	url := fmt.Sprintf(OrderStatusApi, merchantOrderId)
	queryParams := map[string]string{
		OrderDetails: strconv.FormatBool(includeDetails),
	}

	var statusResponse commonResponse.OrderStatusResponse
	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, queryParams, &statusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildOrderStatusEventWithError(
			enums.FAILED,
			merchantOrderId,
			url,
			enums.PG,
			enums.ORDER_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildOrderStatusEvent(
		enums.SUCCESS,
		merchantOrderId,
		url,
		enums.PG,
		enums.ORDER_STATUS_SUCCESS,
	))
	return &statusResponse, nil
}

// Refund initiates a refund request
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (c *CustomCheckoutClient) Refund(ctx context.Context, refundRequest *request.RefundRequest) (*commonResponse.RefundResponse, error) {
	url := RefundApi
	var refundResponse commonResponse.RefundResponse

	err := c.RequestViaAuthRefresh(ctx, http.POST, refundRequest, url, nil, &refundResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildRefundEventWithError(
			enums.FAILED,
			refundRequest,
			url,
			enums.PG,
			enums.REFUND_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildRefundEvent(
		enums.SUCCESS,
		refundRequest,
		url,
		enums.PG,
		enums.REFUND_SUCCESS,
	))
	return &refundResponse, nil
}

// GetTransactionStatus retrieves the status of a transaction by transaction ID
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (c *CustomCheckoutClient) GetTransactionStatus(ctx context.Context, transactionId string) (*commonResponse.OrderStatusResponse, error) {
	url := fmt.Sprintf(TransactionStatusApi, transactionId)
	var statusResponse commonResponse.OrderStatusResponse

	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &statusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildTransactionStatusEventWithError(
			enums.FAILED,
			transactionId,
			url,
			enums.PG,
			enums.TRANSACTION_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildTransactionStatusEvent(
		enums.SUCCESS,
		transactionId,
		url,
		enums.PG,
		enums.TRANSACTION_STATUS_SUCCESS,
	))
	return &statusResponse, nil
}

// GetRefundStatus retrieves the status of a refund by refund ID
// ctx can be used to cancel the request, set timeouts, or propagate trace IDs
func (c *CustomCheckoutClient) GetRefundStatus(ctx context.Context, refundId string) (*commonResponse.RefundStatusResponse, error) {
	url := fmt.Sprintf(RefundStatusApi, refundId)
	var refundStatusResponse commonResponse.RefundStatusResponse

	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &refundStatusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildRefundStatusEventWithError(
			enums.FAILED,
			refundId,
			url,
			enums.PG,
			enums.REFUND_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildRefundStatusEvent(
		enums.SUCCESS,
		refundId,
		url,
		enums.PG,
		enums.REFUND_STATUS_SUCCESS,
	))
	return &refundStatusResponse, nil
}

func (c *CustomCheckoutClient) ValidateCallback(username string, password string, authorization string, responseBody string) (*commonResponse.CallbackResponse, error) {
	if !common.IsCallbackValid(username, password, authorization) {
		return nil, fmt.Errorf("invalid callback")
	}
	var callbackResponse commonResponse.CallbackResponse
	err := json.Unmarshal([]byte(responseBody), &callbackResponse)
	if err != nil {
		c.EventPublisher.Send(models.BuildCallbackSerializationFailedEvent(
			enums.FAILED,
			enums.PG,
			enums.CALLBACK_SERIALIZATION_FAILED,
			err,
		))
		return nil, err
	}
	return &callbackResponse, nil
}
