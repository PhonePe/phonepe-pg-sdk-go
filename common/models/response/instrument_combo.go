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
	"encoding/json"
	"fmt"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/paymentinstruments"
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/response/rails"
)

type InstrumentCombo struct {
	Instrument paymentinstruments.PaymentInstrumentV2 `json:"instrument"`
	Rail       rails.PaymentRail                      `json:"rail"`
	Amount     int64                                  `json:"amount"`
}

// UnmarshalJSON implements custom JSON unmarshaling for polymorphic instrument and rail types
func (ic *InstrumentCombo) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary structure to read raw JSON
	var temp struct {
		Instrument json.RawMessage `json:"instrument"`
		Rail       json.RawMessage `json:"rail"`
		Amount     int64           `json:"amount"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Unmarshal instrument based on type
	var instrument paymentinstruments.PaymentInstrumentV2
	if len(temp.Instrument) > 0 && string(temp.Instrument) != "null" {
		var instrumentTypeHolder struct {
			Type paymentinstruments.PaymentInstrumentType `json:"type"`
		}

		if err := json.Unmarshal(temp.Instrument, &instrumentTypeHolder); err != nil {
			return err
		}

		switch instrumentTypeHolder.Type {
		case paymentinstruments.WALLET:
			var w paymentinstruments.WalletPaymentInstrumentV2
			if err := json.Unmarshal(temp.Instrument, &w); err != nil {
				return err
			}
			instrument = &w
		case paymentinstruments.EGV:
			var e paymentinstruments.EGVPaymentInstrumentV2
			if err := json.Unmarshal(temp.Instrument, &e); err != nil {
				return err
			}
			instrument = &e
		case paymentinstruments.NET_BANKING:
			var n paymentinstruments.NetBankingPaymentInstrumentV2
			if err := json.Unmarshal(temp.Instrument, &n); err != nil {
				return err
			}
			instrument = &n
		case paymentinstruments.ACCOUNT:
			var a paymentinstruments.AccountPaymentInstrumentV2
			if err := json.Unmarshal(temp.Instrument, &a); err != nil {
				return err
			}
			instrument = &a
		case paymentinstruments.CREDIT_CARD:
			var c paymentinstruments.CreditCardPaymentInstrumentV2
			if err := json.Unmarshal(temp.Instrument, &c); err != nil {
				return err
			}
			instrument = &c
		case paymentinstruments.DEBIT_CARD:
			var d paymentinstruments.DebitCardPaymentInstrumentV2
			if err := json.Unmarshal(temp.Instrument, &d); err != nil {
				return err
			}
			instrument = &d
		default:
			return fmt.Errorf("unknown payment instrument type: %s", instrumentTypeHolder.Type)
		}
	}

	// Unmarshal rail based on type
	var rail rails.PaymentRail
	if len(temp.Rail) > 0 && string(temp.Rail) != "null" {
		var railTypeHolder struct {
			Type rails.PaymentRailType `json:"type"`
		}

		if err := json.Unmarshal(temp.Rail, &railTypeHolder); err != nil {
			return err
		}

		switch railTypeHolder.Type {
		case rails.UPI:
			var u rails.UpiPaymentRail
			if err := json.Unmarshal(temp.Rail, &u); err != nil {
				return err
			}
			rail = &u
		case rails.PG:
			var p rails.PgPaymentRail
			if err := json.Unmarshal(temp.Rail, &p); err != nil {
				return err
			}
			rail = &p
		case rails.PPI_WALLET:
			var pw rails.PpiWalletPaymentRail
			if err := json.Unmarshal(temp.Rail, &pw); err != nil {
				return err
			}
			rail = &pw
		case rails.PPI_EGV:
			var pe rails.PpiEgvPaymentRail
			if err := json.Unmarshal(temp.Rail, &pe); err != nil {
				return err
			}
			rail = &pe
		default:
			return fmt.Errorf("unknown payment rail type: %s", railTypeHolder.Type)
		}
	}

	ic.Instrument = instrument
	ic.Rail = rail
	ic.Amount = temp.Amount

	return nil
}
