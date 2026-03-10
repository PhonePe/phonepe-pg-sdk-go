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

package response

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response"
	"github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/request"
)

type SubscriptionSetupPaymentFlowResponse struct {
	response.PaymentFlowResponse
	MerchantSubscriptionId string                   `json:"merchantSubscriptionId"`
	AuthWorkflowType       request.AuthWorkflowType `json:"authWorkflowType"`
	AmountType             request.AmountType       `json:"amountType"`
	MaxAmount              int64                    `json:"maxAmount"`
	Frequency              request.Frequency        `json:"frequency"`
	ExpireAt               int64                    `json:"expireAt"`
	SubscriptionId         string                   `json:"subscriptionId"`
}

func NewSubscriptionSetupPaymentFlowResponse(merchantSubscriptionId string, authWorkflowType request.AuthWorkflowType,
	amountType request.AmountType, maxAmount int64, frequency request.Frequency, expireAt int64,
	subscriptionId string) *SubscriptionSetupPaymentFlowResponse {
	return &SubscriptionSetupPaymentFlowResponse{
		PaymentFlowResponse:    response.PaymentFlowResponse{Type: models.SUBSCRIPTION_SETUP},
		MerchantSubscriptionId: merchantSubscriptionId,
		AuthWorkflowType:       authWorkflowType,
		AmountType:             amountType,
		MaxAmount:              maxAmount,
		Frequency:              frequency,
		ExpireAt:               expireAt,
		SubscriptionId:         subscriptionId,
	}
}
