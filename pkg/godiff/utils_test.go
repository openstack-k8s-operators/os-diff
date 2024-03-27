/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 *
 */
package godiff_test

import (
	"reflect"
	"testing"

	"github.com/openstack-k8s-operators/os-diff/pkg/godiff"
	"github.com/stretchr/testify/assert"
)

// Test case for function CompareYAML
func TestCompareYAML(t *testing.T) {
	origin := []byte("key1: value1\nkey2: value2")
	dest := []byte("key1: value1\nkey3: value3")
	expected := []string{"+key2: value2\n", "-key3: value3\n"}
	report, err := godiff.CompareYAML(origin, dest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(report, expected) {
		t.Errorf("Report mismatch, got: %v, want: %v", report, expected)
	}
}

func TestCompareYAMLEqualMaps(t *testing.T) {
	origin := []byte("key1: value1\nkey2: value2")
	dest := []byte("key1: value1\nkey2: value2")
	report, err := godiff.CompareYAML(origin, dest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(report) > 0 {
		t.Errorf("Report should be empty for equal maps, got: %v", report)
	}
}

// Test case for function CompareJSON
func TestCompareJSON(t *testing.T) {
	// Test case for comparing two identical JSON objects
	orgData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": []string{"a", "b", "c"},
	}
	destData := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": []string{"a", "b", "c"},
	}
	diff, err := godiff.CompareJSON(orgData, destData, "")

	// Assertion for no differences and no error
	assert.NoError(t, err)
	assert.Empty(t, diff)

	// Test case for comparing JSON objects with differences
	orgData2 := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": []string{"a", "b", "c"},
	}
	destData2 := map[string]interface{}{
		"key1": "value1",
		"key2": 456,
		"key4": "value4",
	}
	diff2, err2 := godiff.CompareJSON(orgData2, destData2, "")
	// Assertion for differences and no error
	assert.NoError(t, err2)
	assert.NotEmpty(t, diff2)
	assert.Len(t, diff2, 3)
	assert.Contains(t, diff2, "+key4")
	assert.Contains(t, diff2, "-key3")

	// Test case for comparing JSON objects with type mismatch
	orgData3 := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	destData3 := []interface{}{"value1", 123}
	diff3, err3 := godiff.CompareJSON(orgData3, destData3, "")

	// Assertion for type mismatch error
	assert.Error(t, err3)
	assert.Empty(t, diff3)
}

// Test case for function CompareRawData
func TestCompareRawData(t *testing.T) {
	rawdata1 := []byte("line1\nline2\nline3\n#comment\nline5")
	rawdata2 := []byte("line1\nline3\nline4\n#comment\nline5")
	origin := "file1.txt"
	dest := "file2.txt"

	expectedReport := []string{
		"Source file path: file1.txt, difference with: file2.txt\n",
		"@ line: 2\n", "+line2\n",
		"@ line: 3\n", "-line4\n",
	}

	report, err := godiff.CompareRawData(rawdata1, rawdata2, origin, dest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(report, expectedReport) {
		t.Errorf("Report mismatch, expected: %v, got: %v", expectedReport, report)
	}
}
