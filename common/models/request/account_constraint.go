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
	common_models "github.com/PhonePe/phonepe-pg-sdk-go/common/models"
)

type AccountConstraint struct {
	InstrumentConstraint
	AccountNumber string `json:"accountNumber,omitempty"`
	Ifsc          string `json:"ifsc,omitempty"`
}

func NewAccountConstraint(accountNumber, ifsc string) *AccountConstraint {
	return &AccountConstraint{
		InstrumentConstraint: NewInstrumentConstraint(string(common_models.ACCOUNT)),
		AccountNumber:        accountNumber,
		Ifsc:                 ifsc,
	}
}
