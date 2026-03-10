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

package exception

import ()

func MapError(responseCode int, message string, phonePeResponse *PhonePeResponse) error {
	switch responseCode {
	case 400:
		return NewBadRequest(responseCode, message, phonePeResponse)
	case 401:
		return NewUnauthorizedAccess(responseCode, message, phonePeResponse)
	case 403:
		return NewForbiddenAccess(responseCode, message, phonePeResponse)
	case 404:
		return NewResourceNotFound(responseCode, message, phonePeResponse)
	case 409:
		return NewResourceConflict(responseCode, message, phonePeResponse)
	case 410:
		return NewResourceGone(responseCode, message, phonePeResponse)
	case 417:
		return NewExpectationFailed(responseCode, message, phonePeResponse)
	case 422:
		return NewResourceInvalid(responseCode, message, phonePeResponse)
	case 429:
		return NewTooManyRequest(responseCode, message, phonePeResponse)
	case 500:
		return NewServerError(responseCode, message, phonePeResponse)
	default:
		return NewPhonePeExceptionWithPhonePeResponse(responseCode, message, phonePeResponse)
	}
}
