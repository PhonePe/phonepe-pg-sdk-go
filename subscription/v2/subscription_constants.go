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

package v2

const (
	SetupApi              = "/subscriptions/v2/setup"
	NotifyApi             = "/subscriptions/v2/notify"
	RedeemApi             = "/subscriptions/v2/redeem"
	SubscriptionStatusApi = "/subscriptions/v2/%s/status"
	OrderStatusApi        = "/subscriptions/v2/order/%s/status"
	CancelApi             = "/subscriptions/v2/%s/cancel"
	TransactionStatusApi  = "/subscriptions/v2/transaction/%s/status"
	RefundApi             = "/payments/v2/refund"
	RefundStatusApi       = "/payments/v2/refund/%s/status"
)
