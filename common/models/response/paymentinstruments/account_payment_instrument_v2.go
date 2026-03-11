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

package paymentinstruments

type AccountPaymentInstrumentV2 struct {
	Type                PaymentInstrumentType `json:"type"`
	MaskedAccountNumber string                `json:"maskedAccountNumber"`
	Ifsc                string                `json:"ifsc"`
	AccountHolderName   string                `json:"accountHolderName"`
	AccountType         string                `json:"accountType"`
}

func (a *AccountPaymentInstrumentV2) GetType() PaymentInstrumentType {
	return a.Type
}

func NewAccountPaymentInstrumentV2(maskedAccountNumber, ifsc, accountHolderName, accountType string) *AccountPaymentInstrumentV2 {
	return &AccountPaymentInstrumentV2{
		Type:                ACCOUNT,
		MaskedAccountNumber: maskedAccountNumber,
		Ifsc:                ifsc,
		AccountHolderName:   accountHolderName,
		AccountType:         accountType,
	}
}
