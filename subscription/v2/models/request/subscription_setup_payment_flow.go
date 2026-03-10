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

package request

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/instruments"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
)

type SubscriptionSetupPaymentFlow struct {
	request.PaymentFlow
	MerchantSubscriptionId string                          `json:"merchantSubscriptionId"`
	AuthWorkflowType       AuthWorkflowType                `json:"authWorkflowType"`
	AmountType             AmountType                      `json:"amountType"`
	MaxAmount              int64                           `json:"maxAmount"`
	Frequency              Frequency                       `json:"frequency"`
	ExpireAt               *int64                          `json:"expireAt,omitempty"`
	PaymentMode            instruments.PaymentV2Instrument `json:"paymentMode"`
}

func NewSubscriptionSetupPaymentFlow(merchantSubscriptionId string, authWorkflowType AuthWorkflowType,
	amountType AmountType, maxAmount int64, frequency Frequency, expireAt *int64,
	paymentMode instruments.PaymentV2Instrument) *SubscriptionSetupPaymentFlow {
	return &SubscriptionSetupPaymentFlow{
		PaymentFlow:            request.PaymentFlow{Type: models.SUBSCRIPTION_SETUP},
		MerchantSubscriptionId: merchantSubscriptionId,
		AuthWorkflowType:       authWorkflowType,
		AmountType:             amountType,
		MaxAmount:              maxAmount,
		Frequency:              frequency,
		ExpireAt:               expireAt,
		PaymentMode:            paymentMode,
	}
}
