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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
