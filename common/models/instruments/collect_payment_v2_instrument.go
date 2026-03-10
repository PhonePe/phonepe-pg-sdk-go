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

package instruments

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	requestInstruments "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request/instruments"
)

type CollectPaymentV2Instrument struct {
	BasePaymentV2Instrument
	Details requestInstruments.CollectPaymentDetails `json:"details,omitempty"`
	Message string                                   `json:"message,omitempty"`
}

func NewCollectPaymentV2Instrument(details requestInstruments.CollectPaymentDetails, message string) *CollectPaymentV2Instrument {
	return &CollectPaymentV2Instrument{
		BasePaymentV2Instrument: NewBasePaymentV2Instrument(models.UPI_COLLECT),
		Details:                 details,
		Message:                 message,
	}
}
