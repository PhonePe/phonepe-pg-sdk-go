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

package models

type PgV2InstrumentType string

const (
	UPI_COLLECT  PgV2InstrumentType = "UPI_COLLECT"
	UPI_INTENT   PgV2InstrumentType = "UPI_INTENT"
	PPE_INTENT   PgV2InstrumentType = "PPE_INTENT"
	UPI_QR       PgV2InstrumentType = "UPI_QR"
	CARD         PgV2InstrumentType = "CARD"
	TOKEN        PgV2InstrumentType = "TOKEN"
	NET_BANKING  PgV2InstrumentType = "NET_BANKING"
	UPI_AUTO_PAY PgV2InstrumentType = "UPI_AUTO_PAY"
	ACCOUNT      PgV2InstrumentType = "ACCOUNT"
)
