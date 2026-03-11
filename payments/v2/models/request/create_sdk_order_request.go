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
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
)

type CreateSdkOrderRequest struct {
	MerchantOrderID     string                         `json:"merchantOrderId"`
	Amount              int64                          `json:"amount"`
	MetaInfo            models.MetaInfo                `json:"metaInfo"`
	PaymentFlow         request.PaymentFlowInterface   `json:"paymentFlow"`
	ExpireAfter         int64                          `json:"expireAfter,omitempty"`
	Constraints         []request.InstrumentConstraint `json:"constraints,omitempty"`
	DisablePaymentRetry *bool                          `json:"disablePaymentRetry,omitempty"`
}

func NewStandardCheckoutCreateSdkOrderRequest(amount int64, merchantOrderID string, metaInfo models.MetaInfo, message string, redirectURL string, expireAfter int64, disablePaymentRetry *bool, paymentModeConfig *PaymentModeConfig) *CreateSdkOrderRequest {
	merchantUrls := NewMerchantUrls(redirectURL)
	paymentFlow := NewPgCheckoutPaymentFlow(message, merchantUrls, paymentModeConfig)
	return &CreateSdkOrderRequest{
		MerchantOrderID:     merchantOrderID,
		Amount:              amount,
		MetaInfo:            metaInfo,
		ExpireAfter:         expireAfter,
		PaymentFlow:         paymentFlow,
		DisablePaymentRetry: disablePaymentRetry,
	}
}

func NewCustomCheckoutCreateSdkOrderRequest(merchantOrderID string, amount int64, metaInfo models.MetaInfo, constraints []request.InstrumentConstraint, expireAfter int64, disablePaymentRetry *bool) *CreateSdkOrderRequest {
	// For custom checkout, we don't specify a payment instrument upfront
	paymentFlow := &request.PaymentFlow{Type: models.PG}
	return &CreateSdkOrderRequest{
		MerchantOrderID:     merchantOrderID,
		Amount:              amount,
		MetaInfo:            metaInfo,
		ExpireAfter:         expireAfter,
		Constraints:         constraints,
		PaymentFlow:         paymentFlow,
		DisablePaymentRetry: disablePaymentRetry,
	}
}
