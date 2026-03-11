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
	"github.com/PhonePe/phonepe-pg-sdk-go/common/constants"
)

// Version can be set at build time using ldflags:
// go build -ldflags "-X github.com/PhonePe/phonepe-pg-sdk-go/common/constants.Version=v1.0.0"
// This is an alias to constants.Version for backwards compatibility
var Version = &constants.Version

// GetSDKVersion returns the current SDK version from constants package
func GetSDKVersion() string {
	return constants.SDKVersion
}
