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

type CreditLinePaymentInstrumentV2 struct {
	Type                PaymentInstrumentType `json:"type"`
	Ifsc                string                `json:"ifsc"`
	AccountHolderName   string                `json:"accountHolderName"`
	BankID              string                `json:"bankId"`
	MaskedAccountNumber string                `json:"maskedAccountNumber"`
	ProviderAccountType string                `json:"providerAccountType"`
}

func (c *CreditLinePaymentInstrumentV2) GetType() PaymentInstrumentType {
	return c.Type
}

func NewCreditLinePaymentInstrumentV2(ifsc, accountHolderName, bankID, maskedAccountNumber, providerAccountType string) *CreditLinePaymentInstrumentV2 {
	return &CreditLinePaymentInstrumentV2{
		Type:                CREDIT_LINE,
		Ifsc:                ifsc,
		AccountHolderName:   accountHolderName,
		BankID:              bankID,
		MaskedAccountNumber: maskedAccountNumber,
		ProviderAccountType: providerAccountType,
	}
}
