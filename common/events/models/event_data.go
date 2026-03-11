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
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/events/models/enums"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
)

type EventData struct {
	FlowType   enums.FlowType `json:"flowType,omitempty"`
	SdkType    string         `json:"sdkType,omitempty"`
	SdkVersion string         `json:"sdkVersion,omitempty"`

	ApiPath                 string                    `json:"apiPath,omitempty"`
	Amount                  int64                     `json:"amount,omitempty"`
	TargetApp               string                    `json:"targetApp,omitempty"`
	DeviceContext           *request.DeviceContext    `json:"deviceContext,omitempty"`
	ExpireAfter             int64                     `json:"expireAfter,omitempty"`
	MerchantRefundId        string                    `json:"merchantRefundId,omitempty"`
	OriginalMerchantOrderId string                    `json:"originalMerchantOrderId,omitempty"`
	TransactionId           string                    `json:"transactionId,omitempty"`
	EventState              enums.EventState          `json:"eventState,omitempty"`
	PaymentInstrument       models.PgV2InstrumentType `json:"paymentInstrument,omitempty"`

	CachedTokenIssuedAt        int64 `json:"cachedTokenIssuedAt,omitempty"`
	CachedTokenExpiresAt       int64 `json:"cachedTokenExpiresAt,omitempty"`
	TokenFetchAttemptTimestamp int64 `json:"tokenFetchAttemptTimestamp,omitempty"`

	SubscriptionEventData *SubscriptionEventData `json:"subscriptionEventData,omitempty"`

	ExceptionClass          string                 `json:"exceptionClass,omitempty"`
	ExceptionMessage        string                 `json:"exceptionMessage,omitempty"`
	ExceptionCode           string                 `json:"exceptionCode,omitempty"`
	ExceptionHttpStatusCode int                    `json:"exceptionHttpStatusCode,omitempty"`
	ExceptionData           map[string]interface{} `json:"exceptionData,omitempty"`
}

func NewEventData() *EventData {
	return &EventData{
		SdkType:    events.SDK_TYPE,
		SdkVersion: events.SDK_VERSION,
	}
}
