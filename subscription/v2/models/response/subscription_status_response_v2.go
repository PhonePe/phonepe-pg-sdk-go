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

import "github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/request"

type SubscriptionStatusResponseV2 struct {
	MerchantSubscriptionId string                   `json:"merchantSubscriptionId"`
	SubscriptionId         string                   `json:"subscriptionId"`
	State                  string                   `json:"state"`
	AuthWorkflowType       request.AuthWorkflowType `json:"authWorkflowType"`
	AmountType             request.AmountType       `json:"amountType"`
	MaxAmount              int64                    `json:"maxAmount"`
	Frequency              request.Frequency        `json:"frequency"`
	ExpireAt               int64                    `json:"expireAt"`
	PauseStartDate         int64                    `json:"pauseStartDate"`
	PauseEndDate           int64                    `json:"pauseEndDate"`
}

func NewSubscriptionStatusResponseV2(merchantSubscriptionId string, subscriptionId string, state string,
	authWorkflowType request.AuthWorkflowType, amountType request.AmountType, maxAmount int64,
	frequency request.Frequency, expireAt int64, pauseStartDate int64,
	pauseEndDate int64) *SubscriptionStatusResponseV2 {
	return &SubscriptionStatusResponseV2{
		MerchantSubscriptionId: merchantSubscriptionId,
		SubscriptionId:         subscriptionId,
		State:                  state,
		AuthWorkflowType:       authWorkflowType,
		AmountType:             amountType,
		MaxAmount:              maxAmount,
		Frequency:              frequency,
		ExpireAt:               expireAt,
		PauseStartDate:         pauseStartDate,
		PauseEndDate:           pauseEndDate,
	}
}
