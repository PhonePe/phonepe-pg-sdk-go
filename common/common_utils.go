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

package common

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

func CalculateSha256(args ...interface{}) string {
	var strArgs []string
	for _, arg := range args {
		strArgs = append(strArgs, fmt.Sprintf("%v", arg))
	}
	data := strings.Join(strArgs, ":")
	return shaHex(data)
}

func shaHex(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func IsCallbackValid(username, password, authorization string) bool {
	sha256hash := CalculateSha256(username, password)
	// Use constant time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(sha256hash), []byte(authorization)) == 1
}
