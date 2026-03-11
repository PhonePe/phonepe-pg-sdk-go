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
	common_request "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
)

type PgCheckoutPaymentFlow struct {
	common_request.PaymentFlow
	Message           string             `json:"message"`
	MerchantUrls      *MerchantUrls      `json:"merchantUrls"`
	PaymentModeConfig *PaymentModeConfig `json:"paymentModeConfig"`
}

func NewPgCheckoutPaymentFlow(message string, merchantUrls *MerchantUrls, paymentModeConfig *PaymentModeConfig) *PgCheckoutPaymentFlow {
	return &PgCheckoutPaymentFlow{
		PaymentFlow:       common_request.PaymentFlow{Type: models.PG_CHECKOUT},
		Message:           message,
		MerchantUrls:      merchantUrls,
		PaymentModeConfig: paymentModeConfig,
	}
}
