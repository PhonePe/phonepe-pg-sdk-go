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

import (
	"fmt"
)

type PhonePeException struct {
	HttpStatusCode *int
	Message        string
	Data           map[string]interface{}
	Code           string
}

func (e *PhonePeException) Error() string {
	var statusCode interface{} = "<nil>"
	if e.HttpStatusCode != nil {
		statusCode = *e.HttpStatusCode
	}
	return fmt.Sprintf(
		"%T\nhttpStatusCode: %v\nmessage: %s\ndata: %v\ncode : %s\n",
		e, statusCode, e.Message, e.Data, e.Code,
	)
}

func (e *PhonePeException) GetHttpStatusCode() *int {
	return e.HttpStatusCode
}

func (e *PhonePeException) GetCode() string {
	return e.Code
}

func (e *PhonePeException) GetData() map[string]interface{} {
	return e.Data
}

func (e *PhonePeException) GetMessage() string {
	return e.Message
}

func NewPhonePeException(message string) *PhonePeException {
	return &PhonePeException{Message: message}
}

func NewPhonePeExceptionWithHttpStatus(httpStatusCode int, message string) *PhonePeException {
	return &PhonePeException{HttpStatusCode: &httpStatusCode, Message: message}
}

func NewPhonePeExceptionWithPhonePeResponse(
	httpStatusCode int,
	message string,
	phonePeResponse *PhonePeResponse,
) *PhonePeException {
	exception := &PhonePeException{HttpStatusCode: &httpStatusCode, Message: message}
	if phonePeResponse != nil {
		if phonePeResponse.Message != "" {
			exception.Message = phonePeResponse.Message
		}
		exception.Data = phonePeResponse.Data
		if phonePeResponse.ErrorCode != "" {
			exception.Code = phonePeResponse.ErrorCode
		} else if phonePeResponse.Code != "" {
			exception.Code = phonePeResponse.Code
		}
	}
	return exception
}
