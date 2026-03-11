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

package v2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PhonePe/phonepe-pg-sdk-go/common"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models/enums"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/http"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	subscriptionRequest "github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/request"
	subscriptionResponse "github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/response"
)

type SubscriptionClient struct {
	*common.BaseClient
	headers []*http.HttpHeaderPair
}

func (c *SubscriptionClient) prepareHeaders() {
	c.headers = []*http.HttpHeaderPair{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Source", Value: "INTEGRATION"},
		{Key: "x-source-version", Value: constants.SubscriptionAPIVersion},
		{Key: "x-Source-Platform", Value: "BACKEND_GO_SDK"},
		{Key: "x-Source-Platform-Version", Value: common.GetSDKVersion()},
	}
}

// GetInstance creates a new SubscriptionClient instance without retry
func GetInstance(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool) (*SubscriptionClient, error) {
	return GetInstanceWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishEvents, nil)
}

// GetInstanceWithRetry creates a new SubscriptionClient instance with custom retry configuration
func GetInstanceWithRetry(clientId string, clientSecret string, clientVersion int, env types.Env, shouldPublishEvents bool, retryConfig *http.RetryConfig) (*SubscriptionClient, error) {
	shouldPublishInProd := shouldPublishEvents && env.PgHostURL == types.Production.PgHostURL

	baseClient, err := common.NewBaseClientWithRetry(clientId, clientSecret, clientVersion, env, shouldPublishInProd, retryConfig)
	if err != nil {
		return nil, err
	}

	instance := &SubscriptionClient{
		BaseClient: baseClient,
	}
	instance.prepareHeaders()

	// Send client initialization event
	instance.EventPublisher.Send(models.BuildInitClientEvent(
		enums.SUBSCRIPTION,
		enums.SUBSCRIPTION_CLIENT_INITIALIZED,
	))

	return instance, nil
}

// Setup initiates a subscription setup with the provided payment request.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) Setup(ctx context.Context, paymentRequest *request.PgPaymentRequest) (*response.PgPaymentResponse, error) {
	url := SetupApi
	var setupResponse response.PgPaymentResponse

	err := c.RequestViaAuthRefresh(ctx, http.POST, paymentRequest, url, nil, &setupResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildSubscriptionSetupEventWithError(
			enums.FAILED,
			paymentRequest,
			url,
			enums.SETUP_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildSubscriptionSetupEvent(
		enums.SUCCESS,
		paymentRequest,
		url,
		enums.SETUP_SUCCESS,
	))
	return &setupResponse, nil
}

// Notify sends a subscription notification with the provided payment request.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) Notify(ctx context.Context, paymentRequest *request.PgPaymentRequest) (*response.PgPaymentResponse, error) {
	url := NotifyApi
	var notifyResponse response.PgPaymentResponse

	err := c.RequestViaAuthRefresh(ctx, http.POST, paymentRequest, url, nil, &notifyResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildSubscriptionNotifyEventWithError(
			enums.FAILED,
			paymentRequest,
			url,
			enums.NOTIFY_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildSubscriptionNotifyEvent(
		enums.SUCCESS,
		paymentRequest,
		url,
		enums.NOTIFY_SUCCESS,
	))
	return &notifyResponse, nil
}

// Redeem redeems a subscription using the merchant order ID.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) Redeem(ctx context.Context, merchantOrderId string) (*subscriptionResponse.SubscriptionRedeemResponseV2, error) {
	url := RedeemApi
	var redeemResponse subscriptionResponse.SubscriptionRedeemResponseV2
	redeemRequest := subscriptionRequest.NewSubscriptionRedeemRequestV2(merchantOrderId)

	err := c.RequestViaAuthRefresh(ctx, http.POST, redeemRequest, url, nil, &redeemResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildSubscriptionRedeemEventWithError(
			enums.FAILED,
			merchantOrderId,
			url,
			enums.REDEEM_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildSubscriptionRedeemEvent(
		enums.SUCCESS,
		merchantOrderId,
		url,
		enums.REDEEM_SUCCESS,
	))
	return &redeemResponse, nil
}

// GetSubscriptionStatus retrieves the status of a subscription by merchant subscription ID.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) GetSubscriptionStatus(ctx context.Context, merchantSubscriptionId string) (*subscriptionResponse.SubscriptionStatusResponseV2, error) {
	url := fmt.Sprintf(SubscriptionStatusApi, merchantSubscriptionId)
	var statusResponse subscriptionResponse.SubscriptionStatusResponseV2

	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &statusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildSubscriptionStatusEventWithError(
			enums.FAILED,
			merchantSubscriptionId,
			url,
			enums.SUBSCRIPTION_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildSubscriptionStatusEvent(
		enums.SUCCESS,
		merchantSubscriptionId,
		url,
		enums.SUBSCRIPTION_STATUS_SUCCESS,
	))
	return &statusResponse, nil
}

// GetOrderStatus retrieves the status of a subscription order by merchant order ID.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) GetOrderStatus(ctx context.Context, merchantOrderId string) (*response.OrderStatusResponse, error) {
	url := fmt.Sprintf(OrderStatusApi, merchantOrderId)
	var orderStatusResponse response.OrderStatusResponse

	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &orderStatusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildOrderStatusEventWithError(
			enums.FAILED,
			merchantOrderId,
			url,
			enums.SUBSCRIPTION,
			enums.ORDER_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildOrderStatusEvent(
		enums.SUCCESS,
		merchantOrderId,
		url,
		enums.SUBSCRIPTION,
		enums.ORDER_STATUS_SUCCESS,
	))
	return &orderStatusResponse, nil
}

// CancelSubscription cancels an active subscription by merchant subscription ID.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) CancelSubscription(ctx context.Context, merchantSubscriptionId string) error {
	url := fmt.Sprintf(CancelApi, merchantSubscriptionId)

	err := c.RequestViaAuthRefresh(ctx, http.POST, nil, url, nil, nil, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildSubscriptionCancelEventWithError(
			enums.FAILED,
			merchantSubscriptionId,
			url,
			enums.CANCEL_FAILED,
			err,
		))
		return err
	}
	c.EventPublisher.Send(models.BuildSubscriptionCancelEvent(
		enums.SUCCESS,
		merchantSubscriptionId,
		url,
		enums.CANCEL_SUCCESS,
	))
	return nil
}

// GetTransactionStatus retrieves the status of a subscription transaction by transaction ID.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) GetTransactionStatus(ctx context.Context, transactionId string) (*response.OrderStatusResponse, error) {
	url := fmt.Sprintf(TransactionStatusApi, transactionId)
	var transactionStatusResponse response.OrderStatusResponse

	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &transactionStatusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildTransactionStatusEventWithError(
			enums.FAILED,
			transactionId,
			url,
			enums.SUBSCRIPTION,
			enums.TRANSACTION_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildTransactionStatusEvent(
		enums.SUCCESS,
		transactionId,
		url,
		enums.SUBSCRIPTION,
		enums.TRANSACTION_STATUS_SUCCESS,
	))
	return &transactionStatusResponse, nil
}

// Refund initiates a refund for a subscription transaction.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) Refund(ctx context.Context, refundRequest *request.RefundRequest) (*response.RefundResponse, error) {
	url := RefundApi
	var refundResponse response.RefundResponse

	err := c.RequestViaAuthRefresh(ctx, http.POST, refundRequest, url, nil, &refundResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildRefundEventWithError(
			enums.FAILED,
			refundRequest,
			url,
			enums.SUBSCRIPTION,
			enums.REFUND_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildRefundEvent(
		enums.SUCCESS,
		refundRequest,
		url,
		enums.SUBSCRIPTION,
		enums.REFUND_SUCCESS,
	))
	return &refundResponse, nil
}

// GetRefundStatus retrieves the status of a refund by refund ID.
// The context can be used for timeout control and cancellation.
func (c *SubscriptionClient) GetRefundStatus(ctx context.Context, refundId string) (*response.RefundStatusResponse, error) {
	url := fmt.Sprintf(RefundStatusApi, refundId)
	var refundStatusResponse response.RefundStatusResponse

	err := c.RequestViaAuthRefresh(ctx, http.GET, nil, url, nil, &refundStatusResponse, c.headers)
	if err != nil {
		c.EventPublisher.Send(models.BuildRefundStatusEventWithError(
			enums.FAILED,
			refundId,
			url,
			enums.SUBSCRIPTION,
			enums.REFUND_STATUS_FAILED,
			err,
		))
		return nil, err
	}
	c.EventPublisher.Send(models.BuildRefundStatusEvent(
		enums.SUCCESS,
		refundId,
		url,
		enums.SUBSCRIPTION,
		enums.REFUND_STATUS_SUCCESS,
	))
	return &refundStatusResponse, nil
}

func (c *SubscriptionClient) ValidateCallback(username string, password string, authorization string, responseBody string) (interface{}, error) {
	if !common.IsCallbackValid(username, password, authorization) {
		return nil, fmt.Errorf("invalid callback")
	}
	var callbackResponse interface{}
	err := json.Unmarshal([]byte(responseBody), &callbackResponse)
	if err != nil {
		c.EventPublisher.Send(models.BuildCallbackSerializationFailedEvent(
			enums.FAILED,
			enums.SUBSCRIPTION,
			enums.CALLBACK_SERIALIZATION_FAILED,
			err,
		))
		return nil, err
	}
	return callbackResponse, nil
}
