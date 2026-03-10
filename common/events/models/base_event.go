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

package models

import (
	"reflect"
	"time"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/events"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models/enums"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/exception"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request/instruments"
	paymentsV2Request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	subscriptionV2Request "github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/request"
)

type BaseEvent struct {
	MerchantOrderId string          `json:"merchantOrderId,omitempty"`
	EventName       enums.EventType `json:"eventName"`
	EventTime       int64           `json:"eventTime"`
	Data            *EventData      `json:"data"`
}

func populateExceptionFields(baseEvent *BaseEvent, err error) *BaseEvent {
	if baseEvent == nil || baseEvent.Data == nil || err == nil {
		return baseEvent
	}

	baseEvent.Data.ExceptionMessage = err.Error()
	baseEvent.Data.ExceptionClass = reflect.TypeOf(err).Elem().Name()

	if phonePeException, ok := err.(*exception.PhonePeException); ok {
		if httpStatusCode := phonePeException.GetHttpStatusCode(); httpStatusCode != nil {
			baseEvent.Data.ExceptionHttpStatusCode = *httpStatusCode
		}
		baseEvent.Data.ExceptionCode = phonePeException.GetCode()
		baseEvent.Data.ExceptionData = phonePeException.GetData()
	}

	return baseEvent
}

// Client initialization Event
func BuildInitClientEvent(flowType enums.FlowType, eventName enums.EventType) *BaseEvent {
	baseEvent := BuildInitClientEventWithEventType(eventName)
	baseEvent.Data.FlowType = flowType
	return baseEvent
}

// Client initialization Event
func BuildInitClientEventWithEventType(eventName enums.EventType) *BaseEvent {
	return &BaseEvent{
		EventName: eventName,
		EventTime: time.Now().Unix(),
		Data: &EventData{
			EventState: enums.INITIATED,
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

// Pay Event Builder For Standard Checkout
func BuildStandardCheckoutPayEvent(
	eventState enums.EventState,
	standardCheckoutPayRequest *paymentsV2Request.StandardCheckoutPayRequest,
	apiPath string,
	eventName enums.EventType,
) *BaseEvent {
	// Safely dereference ExpireAfter pointer
	var expireAfter int64
	if standardCheckoutPayRequest.ExpireAfter != nil {
		expireAfter = *standardCheckoutPayRequest.ExpireAfter
	}

	return &BaseEvent{
		EventName:       eventName,
		EventTime:       time.Now().Unix(),
		MerchantOrderId: standardCheckoutPayRequest.MerchantOrderID,
		Data: &EventData{
			EventState:  eventState,
			Amount:      standardCheckoutPayRequest.Amount,
			ApiPath:     apiPath,
			ExpireAfter: expireAfter,
			FlowType:    enums.PG_CHECKOUT,
			SdkType:     events.SDK_TYPE,
			SdkVersion:  events.SDK_VERSION,
		},
	}
}

// Pay Failure Event for Standard Checkout --> Exception is PhonePeException
func BuildStandardCheckoutPayEventWithError(
	eventState enums.EventState,
	standardCheckoutPayRequest *paymentsV2Request.StandardCheckoutPayRequest,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildStandardCheckoutPayEvent(eventState, standardCheckoutPayRequest, apiPath, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// Pay Event Builder For Custom Checkout
func BuildCustomCheckoutPayEvent(
	eventState enums.EventState,
	pgPaymentRequest *request.PgPaymentRequest,
	apiPath string,
	eventName enums.EventType,
) *BaseEvent {
	var targetApp string
	var paymentInstrumentType models.PgV2InstrumentType

	if pgPaymentRequest.PaymentFlow != nil {
		if pgPaymentFlow, ok := pgPaymentRequest.PaymentFlow.(*request.PgPaymentFlow); ok {
			if pgPaymentFlow.PaymentMode != nil {
				paymentInstrumentType = pgPaymentFlow.PaymentMode.GetType()
				if paymentInstrumentType == models.UPI_INTENT {
					if intentInstrument, ok := pgPaymentFlow.PaymentMode.(*instruments.IntentPaymentV2Instrument); ok {
						targetApp = intentInstrument.TargetApp
					}
				}
			}
		}
	}

	return &BaseEvent{
		EventName:       eventName,
		EventTime:       time.Now().Unix(),
		MerchantOrderId: pgPaymentRequest.MerchantOrderID,
		Data: &EventData{
			EventState:        eventState,
			Amount:            pgPaymentRequest.Amount,
			ApiPath:           apiPath,
			DeviceContext:     pgPaymentRequest.DeviceContext,
			ExpireAfter:       pgPaymentRequest.ExpireAfter,
			TargetApp:         targetApp,
			PaymentInstrument: paymentInstrumentType,
			FlowType:          enums.PG,
			SdkType:           events.SDK_TYPE,
			SdkVersion:        events.SDK_VERSION,
		},
	}
}

// Pay Failure Event for Custom Checkout --> Exception is PhonePeException
func BuildCustomCheckoutPayEventWithError(
	eventState enums.EventState,
	pgPaymentRequest *request.PgPaymentRequest,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildCustomCheckoutPayEvent(eventState, pgPaymentRequest, apiPath, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// Order Status Event Builder
func BuildOrderStatusEvent(
	eventState enums.EventState,
	merchantOrderId string,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		EventName:       eventName,
		EventTime:       time.Now().Unix(),
		MerchantOrderId: merchantOrderId,
		Data: &EventData{
			EventState: eventState,
			FlowType:   flowType,
			ApiPath:    apiPath,
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

// Order Status Failure Event
func BuildOrderStatusEventWithError(
	eventState enums.EventState,
	merchantOrderId string,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildOrderStatusEvent(eventState, merchantOrderId, apiPath, flowType, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// Refund Event Builder
func BuildRefundEvent(
	eventState enums.EventState,
	refundRequest *request.RefundRequest,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		EventName:       eventName,
		EventTime:       time.Now().Unix(),
		MerchantOrderId: refundRequest.OriginalMerchantOrderID,
		Data: &EventData{
			MerchantRefundId:        refundRequest.MerchantRefundID,
			EventState:              eventState,
			ApiPath:                 apiPath,
			Amount:                  refundRequest.Amount,
			OriginalMerchantOrderId: refundRequest.OriginalMerchantOrderID,
			FlowType:                flowType,
			SdkType:                 events.SDK_TYPE,
			SdkVersion:              events.SDK_VERSION,
		},
	}
}

// Failure Event for Refund --> Exception is PhonePeException
func BuildRefundEventWithError(
	eventState enums.EventState,
	refundRequest *request.RefundRequest,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildRefundEvent(eventState, refundRequest, apiPath, flowType, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// Refund Status Event Builder
func BuildRefundStatusEvent(
	eventState enums.EventState,
	refundId string,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		EventName: eventName,
		EventTime: time.Now().Unix(),
		Data: &EventData{
			MerchantRefundId: refundId,
			EventState:       eventState,
			ApiPath:          apiPath,
			FlowType:         flowType,
			SdkType:          events.SDK_TYPE,
			SdkVersion:       events.SDK_VERSION,
		},
	}
}

// Failure Event for Refund Status --> Exception is PhonePeException
func BuildRefundStatusEventWithError(
	eventState enums.EventState,
	refundId string,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildRefundStatusEvent(eventState, refundId, apiPath, flowType, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// CreateSdkOrder Event Builder
func BuildCreateSdkOrderEvent(
	eventState enums.EventState,
	createSdkOrderRequest *paymentsV2Request.CreateSdkOrderRequest,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		MerchantOrderId: createSdkOrderRequest.MerchantOrderID,
		EventName:       eventName,
		EventTime:       time.Now().Unix(),
		Data: &EventData{
			Amount:      createSdkOrderRequest.Amount,
			EventState:  eventState,
			ApiPath:     apiPath,
			ExpireAfter: createSdkOrderRequest.ExpireAfter,
			FlowType:    flowType,
			SdkType:     events.SDK_TYPE,
			SdkVersion:  events.SDK_VERSION,
		},
	}
}

// Failure Event for CreateSdkOrder --> Exception is PhonePeException
func BuildCreateSdkOrderEventWithError(
	eventState enums.EventState,
	createSdkOrderRequest *paymentsV2Request.CreateSdkOrderRequest,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildCreateSdkOrderEvent(eventState, createSdkOrderRequest, apiPath, flowType, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// Transaction Status Event Builder
func BuildTransactionStatusEvent(
	eventState enums.EventState,
	transactionId string,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		EventName: eventName,
		EventTime: time.Now().Unix(),
		Data: &EventData{
			EventState:    eventState,
			TransactionId: transactionId,
			ApiPath:       apiPath,
			FlowType:      flowType,
			SdkType:       events.SDK_TYPE,
			SdkVersion:    events.SDK_VERSION,
		},
	}
}

// Failure Event for Transaction Status --> Exception is PhonePeException
func BuildTransactionStatusEventWithError(
	eventState enums.EventState,
	transactionId string,
	apiPath string,
	flowType enums.FlowType,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildTransactionStatusEvent(eventState, transactionId, apiPath, flowType, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// Subscription Setup Event Builder
func BuildSubscriptionSetupEvent(
	eventState enums.EventState,
	setupRequest *request.PgPaymentRequest,
	apiPath string,
	eventName enums.EventType,
) *BaseEvent {
	var targetApp string
	var paymentInstrumentType models.PgV2InstrumentType
	var subscriptionExpireAt int64
	var merchantSubscriptionId string

	if setupRequest.PaymentFlow != nil {
		if setupPaymentFlow, ok := setupRequest.PaymentFlow.(*subscriptionV2Request.SubscriptionSetupPaymentFlow); ok {
			if setupPaymentFlow.PaymentMode != nil {
				paymentInstrumentType = setupPaymentFlow.PaymentMode.GetType()
				if paymentInstrumentType == models.UPI_INTENT {
					if intentInstrument, ok := setupPaymentFlow.PaymentMode.(*instruments.IntentPaymentV2Instrument); ok {
						targetApp = intentInstrument.TargetApp
					}
				}
			}
			if setupPaymentFlow.ExpireAt != nil {
				subscriptionExpireAt = *setupPaymentFlow.ExpireAt
			}
			merchantSubscriptionId = setupPaymentFlow.MerchantSubscriptionId
		}
	}

	return &BaseEvent{
		MerchantOrderId: setupRequest.MerchantOrderID,
		EventTime:       time.Now().Unix(),
		EventName:       eventName,
		Data: &EventData{
			EventState:        eventState,
			ApiPath:           apiPath,
			FlowType:          enums.SUBSCRIPTION,
			PaymentInstrument: paymentInstrumentType,
			TargetApp:         targetApp,
			DeviceContext:     setupRequest.DeviceContext,
			Amount:            setupRequest.Amount,
			SubscriptionEventData: &SubscriptionEventData{
				OrderExpireAt:          setupRequest.ExpireAt,
				SubscriptionExpireAt:   subscriptionExpireAt,
				MerchantSubscriptionId: merchantSubscriptionId,
			},
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

// Failure Event for Subscription Setup
func BuildSubscriptionSetupEventWithError(
	eventState enums.EventState,
	setupRequest *request.PgPaymentRequest,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	event := BuildSubscriptionSetupEvent(eventState, setupRequest, apiPath, eventName)
	return populateExceptionFields(event, exception)
}

// Subscription Notify Event Builder
func BuildSubscriptionNotifyEvent(
	eventState enums.EventState,
	notifyRequest *request.PgPaymentRequest,
	apiPath string,
	eventName enums.EventType,
) *BaseEvent {
	var merchantSubscriptionId string
	if notifyRequest.PaymentFlow != nil {
		if paymentFlow, ok := notifyRequest.PaymentFlow.(*subscriptionV2Request.SubscriptionRedemptionPaymentFlow); ok {
			merchantSubscriptionId = paymentFlow.MerchantSubscriptionId
		}
	}

	return &BaseEvent{
		EventName:       eventName,
		EventTime:       time.Now().Unix(),
		MerchantOrderId: notifyRequest.MerchantOrderID,
		Data: &EventData{
			EventState: eventState,
			ApiPath:    apiPath,
			Amount:     notifyRequest.Amount,
			FlowType:   enums.SUBSCRIPTION,
			SubscriptionEventData: &SubscriptionEventData{
				OrderExpireAt:          notifyRequest.ExpireAt,
				MerchantSubscriptionId: merchantSubscriptionId,
			},
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

// Failure Event for Subscription Notify
func BuildSubscriptionNotifyEventWithError(
	eventState enums.EventState,
	notifyRequest *request.PgPaymentRequest,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	event := BuildSubscriptionNotifyEvent(eventState, notifyRequest, apiPath, eventName)
	return populateExceptionFields(event, exception)
}

func BuildSubscriptionRedeemEvent(
	eventState enums.EventState, merchantOrderId string, apiPath string, eventName enums.EventType) *BaseEvent {
	return &BaseEvent{
		EventTime:       time.Now().Unix(),
		EventName:       eventName,
		MerchantOrderId: merchantOrderId,
		Data: &EventData{
			FlowType:   enums.SUBSCRIPTION,
			EventState: eventState,
			ApiPath:    apiPath,
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

func BuildSubscriptionRedeemEventWithError(
	eventState enums.EventState,
	merchantOrderId string,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	event := BuildSubscriptionRedeemEvent(eventState, merchantOrderId, apiPath, eventName)
	return populateExceptionFields(event, exception)
}

// Order Status Event Builder
func BuildSubscriptionStatusEvent(
	eventState enums.EventState,
	merchantSubscriptionId string,
	apiPath string,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		EventName: eventName,
		EventTime: time.Now().Unix(),
		Data: &EventData{
			EventState: eventState,
			ApiPath:    apiPath,
			FlowType:   enums.SUBSCRIPTION,
			SubscriptionEventData: &SubscriptionEventData{
				MerchantSubscriptionId: merchantSubscriptionId,
			},
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

func BuildSubscriptionStatusEventWithError(
	eventState enums.EventState,
	merchantSubscriptionId string,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildSubscriptionStatusEvent(eventState, merchantSubscriptionId, apiPath, eventName)
	return populateExceptionFields(baseEvent, exception)
}

func BuildSubscriptionCancelEvent(
	eventState enums.EventState,
	merchantSubscriptionId string,
	apiPath string,
	eventName enums.EventType,
) *BaseEvent {
	return &BaseEvent{
		EventTime: time.Now().Unix(),
		EventName: eventName,
		Data: &EventData{
			EventState: eventState,
			ApiPath:    apiPath,
			FlowType:   enums.SUBSCRIPTION,
			SubscriptionEventData: &SubscriptionEventData{
				MerchantSubscriptionId: merchantSubscriptionId,
			},
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
}

func BuildSubscriptionCancelEventWithError(
	eventState enums.EventState,
	merchantSubscriptionId string,
	apiPath string,
	eventName enums.EventType,
	exception error,
) *BaseEvent {
	baseEvent := BuildSubscriptionCancelEvent(eventState, merchantSubscriptionId, apiPath, eventName)
	return populateExceptionFields(baseEvent, exception)
}

// OAuth Event Builder for Cached Token Present
func BuildOAuthEvent(
	fetchAttemptTime int64,
	apiPath string,
	eventName enums.EventType,
	exception error,
	cachedTokenIssuedAt int64,
	cachedTokenExpiresAt int64,
) *BaseEvent {
	baseEvent := &BaseEvent{
		EventName: eventName,
		EventTime: time.Now().Unix(),
		Data: &EventData{
			TokenFetchAttemptTimestamp: fetchAttemptTime,
			ApiPath:                    apiPath,
			EventState:                 enums.FAILED,
			CachedTokenExpiresAt:       cachedTokenExpiresAt,
			CachedTokenIssuedAt:        cachedTokenIssuedAt,
			SdkType:                    events.SDK_TYPE,
			SdkVersion:                 events.SDK_VERSION,
		},
	}
	return populateExceptionFields(baseEvent, exception)
}

func BuildCallbackSerializationFailedEvent(
	eventState enums.EventState, flowType enums.FlowType, eventName enums.EventType, exception error) *BaseEvent {
	baseEvent := &BaseEvent{
		EventName: eventName,
		EventTime: time.Now().Unix(),
		Data: &EventData{
			EventState: eventState,
			FlowType:   flowType,
			SdkType:    events.SDK_TYPE,
			SdkVersion: events.SDK_VERSION,
		},
	}
	return populateExceptionFields(baseEvent, exception)
}
