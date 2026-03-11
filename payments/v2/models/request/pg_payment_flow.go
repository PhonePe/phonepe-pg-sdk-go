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
	p_instruments "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request/instruments"
)

type PgPaymentFlow struct {
	common_request.PaymentFlow
	PaymentMode  p_instruments.PaymentV2Instrument `json:"paymentMode,omitempty"`
	MerchantUrls *MerchantUrls                     `json:"merchantUrls,omitempty"`
}

func NewPgPaymentFlow(paymentMode p_instruments.PaymentV2Instrument, merchantUrls *MerchantUrls) *PgPaymentFlow {
	return &PgPaymentFlow{
		PaymentFlow:  common_request.PaymentFlow{Type: models.PG},
		PaymentMode:  paymentMode,
		MerchantUrls: merchantUrls,
	}
}
