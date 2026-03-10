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
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models/request/instruments"
)

type PgPaymentRequest struct {
	MerchantOrderID string                 `json:"merchantOrderId"`
	Amount          int64                  `json:"amount"`
	MetaInfo        models.MetaInfo        `json:"metaInfo"`
	PaymentFlow     PaymentFlowInterface   `json:"paymentFlow"`
	Constraints     []InstrumentConstraint `json:"constraints,omitempty"`
	DeviceContext   *DeviceContext         `json:"deviceContext,omitempty"`
	ExpireAfter     int64                  `json:"expireAfter,omitempty"`
	ExpireAt        int64                  `json:"expireAt,omitempty"`
	DeviceOS        string                 `json:"-"`
}

func NewUpiIntentPayRequest(merchantOrderID string, amount int64, metaInfo models.MetaInfo, constraints []InstrumentConstraint, deviceOS, merchantCallBackScheme, targetApp string, expireAfter int64) *PgPaymentRequest {
	var deviceContext *DeviceContext
	// Only create DeviceContext if at least one field is provided
	if deviceOS != "" || merchantCallBackScheme != "" {
		deviceContext = NewDeviceContext(deviceOS, merchantCallBackScheme)
	}
	paymentMode := instruments.NewIntentPaymentV2Instrument(targetApp)
	paymentFlow := NewPgPaymentFlow(paymentMode, nil)
	return &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		DeviceContext:   deviceContext,
		PaymentFlow:     paymentFlow,
	}
}

// ... other builder functions to be added here

func NewUpiCollectPayViaVpaRequest(amount int64, merchantOrderID string, metaInfo models.MetaInfo, constraints []InstrumentConstraint, vpa, message string, expireAfter int64, deviceOS ...string) *PgPaymentRequest {
	details := instruments.NewVpaCollectPaymentDetails(vpa)
	paymentMode := instruments.NewCollectPaymentV2Instrument(details, message)
	paymentFlow := NewPgPaymentFlow(paymentMode, nil)
	req := &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
	if len(deviceOS) > 0 {
		req.DeviceOS = deviceOS[0]
	}
	return req
}

func NewUpiCollectPayViaPhoneNumberRequest(amount int64, metaInfo models.MetaInfo, merchantOrderID, phoneNumber string, constraints []InstrumentConstraint, message string, expireAfter int64, deviceOS ...string) *PgPaymentRequest {
	details := instruments.NewPhoneNumberCollectPaymentDetails(phoneNumber)
	paymentMode := instruments.NewCollectPaymentV2Instrument(details, message)
	paymentFlow := NewPgPaymentFlow(paymentMode, nil)
	req := &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
	if len(deviceOS) > 0 {
		req.DeviceOS = deviceOS[0]
	}
	return req
}

func NewUpiQrRequest(amount int64, metaInfo models.MetaInfo, merchantOrderID string, constraints []InstrumentConstraint, expireAfter int64) *PgPaymentRequest {
	paymentMode := instruments.NewUpiQrPaymentV2Instrument()
	paymentFlow := NewPgPaymentFlow(paymentMode, nil)
	return &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
}

func NewNetBankingPayRequest(amount int64, metaInfo models.MetaInfo, constraints []InstrumentConstraint, merchantOrderID, bankID, merchantUserID, redirectURL string, expireAfter int64) *PgPaymentRequest {
	paymentMode := instruments.NewNetBankingPaymentV2Instrument(bankID, merchantUserID)
	merchantUrls := NewMerchantUrls(redirectURL)
	paymentFlow := NewPgPaymentFlow(paymentMode, merchantUrls)
	return &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
}

func NewTokenPayRequest(merchantOrderID string, amount, encryptionKeyID int64, authMode, encryptedToken, encryptedCvv, cryptogram, panSuffix, expiryMonth, expiryYear, redirectURL, cardHolderName, merchantUserID string, metaInfo models.MetaInfo, constraints []InstrumentConstraint, expireAfter int64) *PgPaymentRequest {
	expiry := instruments.Expiry{Month: expiryMonth, Year: expiryYear}
	tokenDetails := instruments.TokenDetails{
		EncryptedToken:  encryptedToken,
		EncryptedCvv:    encryptedCvv,
		EncryptionKeyID: encryptionKeyID,
		Expiry:          expiry,
		Cryptogram:      cryptogram,
		PanSuffix:       panSuffix,
		CardHolderName:  cardHolderName,
	}
	paymentMode := instruments.NewTokenPaymentV2Instrument(authMode, tokenDetails, merchantUserID)
	merchantUrls := NewMerchantUrls(redirectURL)
	paymentFlow := NewPgPaymentFlow(paymentMode, merchantUrls)
	return &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
}

func NewCardPayRequest(merchantOrderID string, amount, encryptionKeyID int64, authMode, encryptedCardNumber, encryptedCvv, expiryMonth, expiryYear, cardHolderName, merchantUserID string, metaInfo models.MetaInfo, constraints []InstrumentConstraint, redirectURL string, expireAfter int64) *PgPaymentRequest {
	expiry := instruments.Expiry{Month: expiryMonth, Year: expiryYear}
	cardDetails := instruments.NewCardDetails{
		EncryptedCardNumber: encryptedCardNumber,
		EncryptionKeyID:     encryptionKeyID,
		CardHolderName:      cardHolderName,
		Expiry:              &expiry,
		EncryptedCvv:        encryptedCvv,
	}
	paymentMode := instruments.NewCardPaymentV2Instrument(authMode, merchantUserID, cardDetails, false)
	merchantUrls := NewMerchantUrls(redirectURL)
	paymentFlow := NewPgPaymentFlow(paymentMode, merchantUrls)
	return &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
}

// NewPgPaymentRequest is a generic constructor that accepts a pre-built payment flow
func NewPgPaymentRequest(merchantOrderID string, amount int64, metaInfo models.MetaInfo, constraints []InstrumentConstraint, paymentFlow PaymentFlowInterface, redirectURL string, expireAfter int64) *PgPaymentRequest {
	if redirectURL != "" {
		if pgFlow, ok := paymentFlow.(*PgPaymentFlow); ok {
			pgFlow.MerchantUrls = NewMerchantUrls(redirectURL)
		}
	}
	return &PgPaymentRequest{
		MerchantOrderID: merchantOrderID,
		Amount:          amount,
		MetaInfo:        metaInfo,
		Constraints:     constraints,
		ExpireAfter:     expireAfter,
		PaymentFlow:     paymentFlow,
	}
}
