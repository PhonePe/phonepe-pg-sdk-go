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
	"github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/request"
)

type CallbackData struct {
	OrderID                 string                   `json:"orderId"`
	MerchantID              string                   `json:"merchantId"`
	MerchantRefundID        string                   `json:"merchantRefundId"`
	OriginalMerchantOrderID string                   `json:"originalMerchantOrderId"`
	RefundID                string                   `json:"refundId"`
	MerchantOrderID         string                   `json:"merchantOrderId"`
	State                   string                   `json:"state"`
	Amount                  int64                    `json:"amount"`
	ExpireAt                int64                    `json:"expireAt"`
	ErrorCode               string                   `json:"errorCode"`
	DetailedErrorCode       string                   `json:"detailedErrorCode"`
	MetaInfo                models.MetaInfo          `json:"metaInfo"`
	MerchantSubscriptionID  string                   `json:"merchantSubscriptionId"`
	SubscriptionID          string                   `json:"subscriptionId"`
	AuthWorkflowType        request.AuthWorkflowType `json:"authWorkflowType"`
	AmountType              request.AmountType       `json:"amountType"`
	MaxAmount               int64                    `json:"maxAmount"`
	Frequency               request.Frequency        `json:"frequency"`
	PauseStartDate          int64                    `json:"pauseStartDate"`
	PauseEndDate            int64                    `json:"pauseEndDate"`
	PaymentFlow             PaymentFlowResponse      `json:"paymentFlow"`
	PaymentDetails          []PaymentDetail          `json:"paymentDetails"`
}
