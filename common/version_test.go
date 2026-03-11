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
	"testing"
)

func TestGetSDKVersion(t *testing.T) {
	version := GetSDKVersion()

	if version == "" {
		t.Error("SDK version should not be empty")
	}

	// Version should be in format like "1.0.0"
	if len(version) < 5 {
		t.Errorf("SDK version seems invalid: %s", version)
	}

	t.Logf("Current SDK Version: %s", version)
}
