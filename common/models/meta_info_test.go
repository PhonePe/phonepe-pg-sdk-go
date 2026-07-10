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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// callNewMetaInfo forwards a fixed-size array to NewMetaInfo to keep table tests concise.
func callNewMetaInfo(args [15]string) (MetaInfo, error) {
	return NewMetaInfo(
		args[0], args[1], args[2], args[3], args[4],
		args[5], args[6], args[7], args[8], args[9],
		args[10], args[11], args[12], args[13], args[14],
	)
}

func TestMetaInfoHasAllUdfFields(t *testing.T) {
	m := MetaInfo{
		Udf1:  "v1",
		Udf2:  "v2",
		Udf3:  "v3",
		Udf4:  "v4",
		Udf5:  "v5",
		Udf6:  "v6",
		Udf7:  "v7",
		Udf8:  "v8",
		Udf9:  "v9",
		Udf10: "v10",
		Udf11: "v11",
		Udf12: "v12",
		Udf13: "v13",
		Udf14: "v14",
		Udf15: "v15",
	}

	assert.Equal(t, "v1", m.Udf1)
	assert.Equal(t, "v5", m.Udf5)
	assert.Equal(t, "v6", m.Udf6)
	assert.Equal(t, "v10", m.Udf10)
	assert.Equal(t, "v15", m.Udf15)
}

func TestMetaInfoJSONFieldNames(t *testing.T) {
	m := MetaInfo{
		Udf1:  "a",
		Udf6:  "b",
		Udf10: "c",
		Udf15: "d",
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var raw map[string]string
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	assert.Equal(t, "a", raw["udf1"])
	assert.Equal(t, "b", raw["udf6"])
	assert.Equal(t, "c", raw["udf10"])
	assert.Equal(t, "d", raw["udf15"])
}

func TestMetaInfoOmitempty(t *testing.T) {
	// Only set udf1 and udf15; others should be omitted from JSON.
	m := MetaInfo{
		Udf1:  "hello",
		Udf15: "world",
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"udf1"`)
	assert.Contains(t, jsonStr, `"udf15"`)

	// Fields not set must not appear in output.
	for _, field := range []string{
		`"udf2"`, `"udf3"`, `"udf4"`, `"udf5"`,
		`"udf6"`, `"udf7"`, `"udf8"`, `"udf9"`,
		`"udf10"`, `"udf11"`, `"udf12"`, `"udf13"`, `"udf14"`,
	} {
		assert.NotContains(t, jsonStr, field)
	}
}

func TestMetaInfoAllFieldsSerializedCorrectly(t *testing.T) {
	m := MetaInfo{
		Udf1:  "1", Udf2: "2", Udf3: "3", Udf4: "4", Udf5: "5",
		Udf6:  "6", Udf7: "7", Udf8: "8", Udf9: "9", Udf10: "10",
		Udf11: "11", Udf12: "12", Udf13: "13", Udf14: "14", Udf15: "15",
	}

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var raw map[string]string
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)

	expectedKeys := []string{
		"udf1", "udf2", "udf3", "udf4", "udf5",
		"udf6", "udf7", "udf8", "udf9", "udf10",
		"udf11", "udf12", "udf13", "udf14", "udf15",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, raw, key, "expected key %q in JSON", key)
	}
}

// ---------------------------------------------------------------------------
// NewMetaInfo – happy paths
// ---------------------------------------------------------------------------

func TestNewMetaInfo_ValidFields(t *testing.T) {
	m, err := NewMetaInfo(
		"v1", "v2", "v3", "v4", "v5",
		"v6", "v7", "v8", "v9", "v10",
		"v11", "v12", "v13", "v14", "v15",
	)
	require.NoError(t, err)
	assert.Equal(t, "v1", m.Udf1)
	assert.Equal(t, "v10", m.Udf10)
	assert.Equal(t, "v11", m.Udf11)
	assert.Equal(t, "v15", m.Udf15)
}

func TestNewMetaInfo_AllEmpty(t *testing.T) {
	m, err := NewMetaInfo("", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	require.NoError(t, err)
	assert.Equal(t, MetaInfo{}, m)
}

func TestNewMetaInfo_Udf1To10_AtMaxLength(t *testing.T) {
	val := strings.Repeat("a", 256)
	var args [15]string
	for i := 0; i < 10; i++ {
		args[i] = val
	}
	m, err := callNewMetaInfo(args)
	require.NoError(t, err)
	assert.Equal(t, val, m.Udf1)
	assert.Equal(t, val, m.Udf10)
}

func TestNewMetaInfo_Udf11To15_AtMaxLength(t *testing.T) {
	val := strings.Repeat("a", 50)
	var args [15]string
	for i := 10; i < 15; i++ {
		args[i] = val
	}
	m, err := callNewMetaInfo(args)
	require.NoError(t, err)
	assert.Equal(t, val, m.Udf11)
	assert.Equal(t, val, m.Udf15)
}

func TestNewMetaInfo_Udf11To15_ValidPatternChars(t *testing.T) {
	validValues := []string{
		"abc123",
		"hello world",
		"user_name",
		"user-name",
		"user@domain.com",
		"value+extra",
		"Mix_ed-val @. +",
	}
	for _, val := range validValues {
		var args [15]string
		for i := 10; i < 15; i++ {
			args[i] = val
		}
		_, err := callNewMetaInfo(args)
		assert.NoError(t, err, "expected valid pattern for %q", val)
	}
}

// ---------------------------------------------------------------------------
// NewMetaInfo – udf1–10 length errors
// ---------------------------------------------------------------------------

func TestNewMetaInfo_Udf1To10_ExceedsMaxLength(t *testing.T) {
	over := strings.Repeat("a", 257)
	tests := []struct {
		field string
		index int
	}{
		{"udf1", 0}, {"udf2", 1}, {"udf3", 2}, {"udf4", 3}, {"udf5", 4},
		{"udf6", 5}, {"udf7", 6}, {"udf8", 7}, {"udf9", 8}, {"udf10", 9},
	}
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			var args [15]string
			args[tt.index] = over
			_, err := callNewMetaInfo(args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.field)
			assert.Contains(t, err.Error(), "256")
		})
	}
}

// ---------------------------------------------------------------------------
// NewMetaInfo – udf11–15 length errors
// ---------------------------------------------------------------------------

func TestNewMetaInfo_Udf11To15_ExceedsMaxLength(t *testing.T) {
	over := strings.Repeat("a", 51)
	tests := []struct {
		field string
		index int
	}{
		{"udf11", 10}, {"udf12", 11}, {"udf13", 12}, {"udf14", 13}, {"udf15", 14},
	}
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			var args [15]string
			args[tt.index] = over
			_, err := callNewMetaInfo(args)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.field)
			assert.Contains(t, err.Error(), "50")
		})
	}
}

// ---------------------------------------------------------------------------
// NewMetaInfo – udf11–15 pattern errors
// ---------------------------------------------------------------------------

func TestNewMetaInfo_Udf11To15_InvalidPatternChars(t *testing.T) {
	invalidValues := []string{"value!", "val#ue", "val$", "val%", "val^", "val&", "val*", "val()", "val<>", "val~"}
	tests := []struct {
		field string
		index int
	}{
		{"udf11", 10}, {"udf12", 11}, {"udf13", 12}, {"udf14", 13}, {"udf15", 14},
	}
	for _, tt := range tests {
		for _, invalid := range invalidValues {
			t.Run(tt.field+"/"+invalid, func(t *testing.T) {
				var args [15]string
				args[tt.index] = invalid
				_, err := callNewMetaInfo(args)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.field)
			})
		}
	}
}
