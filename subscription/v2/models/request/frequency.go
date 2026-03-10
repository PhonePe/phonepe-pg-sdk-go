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

type Frequency string

const (
	DAILY       Frequency = "DAILY"
	WEEKLY      Frequency = "WEEKLY"
	MONTHLY     Frequency = "MONTHLY"
	YEARLY      Frequency = "YEARLY"
	FORTNIGHTLY Frequency = "FORTNIGHTLY"
	BIMONTHLY   Frequency = "BIMONTHLY"
	ON_DEMAND   Frequency = "ON_DEMAND"
	QUARTERLY   Frequency = "QUARTERLY"
	HALFYEARLY  Frequency = "HALFYEARLY"
)
