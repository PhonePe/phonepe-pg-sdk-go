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

import (
	"fmt"
	"regexp"
)

var metaInfoRestrictedPattern = regexp.MustCompile(`^[a-zA-Z0-9_\- @.+]*$`)

type MetaInfo struct {
	Udf1  string `json:"udf1,omitempty"`
	Udf2  string `json:"udf2,omitempty"`
	Udf3  string `json:"udf3,omitempty"`
	Udf4  string `json:"udf4,omitempty"`
	Udf5  string `json:"udf5,omitempty"`
	Udf6  string `json:"udf6,omitempty"`
	Udf7  string `json:"udf7,omitempty"`
	Udf8  string `json:"udf8,omitempty"`
	Udf9  string `json:"udf9,omitempty"`
	Udf10 string `json:"udf10,omitempty"`
	Udf11 string `json:"udf11,omitempty"`
	Udf12 string `json:"udf12,omitempty"`
	Udf13 string `json:"udf13,omitempty"`
	Udf14 string `json:"udf14,omitempty"`
	Udf15 string `json:"udf15,omitempty"`
}

// NewMetaInfo creates a MetaInfo and validates all udf fields automatically.
// udf1–udf10: max 256 characters.
// udf11–udf15: max 50 characters, restricted to alphanumeric and [_ - @ . +].
func NewMetaInfo(udf1, udf2, udf3, udf4, udf5, udf6, udf7, udf8, udf9, udf10, udf11, udf12, udf13, udf14, udf15 string) (MetaInfo, error) {
	m := MetaInfo{
		Udf1: udf1, Udf2: udf2, Udf3: udf3, Udf4: udf4, Udf5: udf5,
		Udf6: udf6, Udf7: udf7, Udf8: udf8, Udf9: udf9, Udf10: udf10,
		Udf11: udf11, Udf12: udf12, Udf13: udf13, Udf14: udf14, Udf15: udf15,
	}
	return m, m.Validate()
}

// Validate checks udf field length and pattern constraints.
// udf1–udf10: max 256 characters.
// udf11–udf15: max 50 characters, restricted to alphanumeric and [_ - @ . +].
func (m MetaInfo) Validate() error {
	const freeMax = 256
	const restrictedMax = 50

	freeFields := []struct{ name, value string }{
		{"udf1", m.Udf1}, {"udf2", m.Udf2}, {"udf3", m.Udf3}, {"udf4", m.Udf4}, {"udf5", m.Udf5},
		{"udf6", m.Udf6}, {"udf7", m.Udf7}, {"udf8", m.Udf8}, {"udf9", m.Udf9}, {"udf10", m.Udf10},
	}
	for _, f := range freeFields {
		if len(f.value) > freeMax {
			return fmt.Errorf("%s exceeds maximum allowed size of %d characters", f.name, freeMax)
		}
	}

	restrictedFields := []struct{ name, value string }{
		{"udf11", m.Udf11}, {"udf12", m.Udf12}, {"udf13", m.Udf13}, {"udf14", m.Udf14}, {"udf15", m.Udf15},
	}
	for _, f := range restrictedFields {
		if len(f.value) > restrictedMax {
			return fmt.Errorf("%s exceeds maximum allowed size of %d characters", f.name, restrictedMax)
		}
		if f.value != "" && !metaInfoRestrictedPattern.MatchString(f.value) {
			return fmt.Errorf("%s should only contain alphanumeric characters, underscores, hyphens, spaces, @, ., and +", f.name)
		}
	}
	return nil
}
