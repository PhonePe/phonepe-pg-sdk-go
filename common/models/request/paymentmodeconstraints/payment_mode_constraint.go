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

package paymentmodeconstraints

import (
	"encoding/json"
	"fmt"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
)

type PaymentModeConstraint interface {
	GetType() models.PgV2InstrumentType
}

type BasePaymentModeConstraint struct {
	Type models.PgV2InstrumentType `json:"type"`
}

func (b *BasePaymentModeConstraint) GetType() models.PgV2InstrumentType {
	return b.Type
}

// Custom unmarshaler for PaymentModeConstraint to handle polymorphism
func UnmarshalPaymentModeConstraint(data []byte) (PaymentModeConstraint, error) {
	var baseConstraint BasePaymentModeConstraint
	if err := json.Unmarshal(data, &baseConstraint); err != nil {
		return nil, err
	}

	switch baseConstraint.Type {
	case models.CARD:
		var cardConstraint CardPaymentModeConstraint
		if err := json.Unmarshal(data, &cardConstraint); err != nil {
			return nil, err
		}
		return &cardConstraint, nil
	case models.NET_BANKING:
		var netBankingConstraint NetBankingPaymentModeConstraint
		if err := json.Unmarshal(data, &netBankingConstraint); err != nil {
			return nil, err
		}
		return &netBankingConstraint, nil
	case models.UPI_INTENT:
		var upiIntentConstraint UpiIntentPaymentModeConstraint
		if err := json.Unmarshal(data, &upiIntentConstraint); err != nil {
			return nil, err
		}
		return &upiIntentConstraint, nil
	case models.UPI_QR:
		var upiQrConstraint UpiQrPaymentModeConstraint
		if err := json.Unmarshal(data, &upiQrConstraint); err != nil {
			return nil, err
		}
		return &upiQrConstraint, nil
	case models.UPI_COLLECT:
		var upiCollectConstraint UpiCollectPaymentModeConstraint
		if err := json.Unmarshal(data, &upiCollectConstraint); err != nil {
			return nil, err
		}
		return &upiCollectConstraint, nil
	default:
		return nil, fmt.Errorf("unknown PaymentModeConstraint type: %s", baseConstraint.Type)
	}
}
