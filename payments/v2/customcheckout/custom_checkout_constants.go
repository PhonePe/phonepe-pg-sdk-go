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

package customcheckout

const (
	PayApi               = "/payments/v2/pay"
	CreateOrderApi       = "/payments/v2/sdk/order"
	OrderStatusApi       = "/payments/v2/order/%s/status"
	OrderDetails         = "details"
	RefundApi            = "/payments/v2/refund"
	RefundStatusApi      = "/payments/v2/refund/%s/status"
	TransactionStatusApi = "/payments/v2/transaction/%s/status"
)
