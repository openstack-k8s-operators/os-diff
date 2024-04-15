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
package common_test

import (
	"testing"

	"github.com/openstack-k8s-operators/os-diff/pkg/common"
)

// Test case for function stringInSlice
func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		name      string
		inputStr  string
		inputList []string
		expected  bool
	}{
		{"String in slice - positive case", "apple", []string{"apple", "banana", "cherry"}, true},
		{"String not in slice - negative case", "pear", []string{"apple", "banana", "cherry"}, false},
		{"Empty slice - edge case", "apple", []string{}, false},
		{"Empty string - edge case", "", []string{"apple", "banana", "cherry"}, false},
		{"String at the beginning of the slice", "apple", []string{"apple", "banana", "cherry"}, true},
		{"String at the end of the slice", "cherry", []string{"apple", "banana", "cherry"}, true},
		{"String in slice with duplicates", "apple", []string{"apple", "apple", "apple"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := common.StringInSlice(tc.inputStr, tc.inputList)
			if result != tc.expected {
				t.Errorf("Expected %v but got %v for input string %s in slice %v", tc.expected, result, tc.inputStr, tc.inputList)
			}
		})
	}
}

// Test case for function sliceIndex
func TestSliceIndex(t *testing.T) {
	tests := []struct {
		name     string
		element  string
		data     []string
		expected int
	}{
		{
			name:     "Element exists in data",
			element:  "apple",
			data:     []string{"orange", "banana", "apple", "grape"},
			expected: 2,
		},
		{
			name:     "Element does not exist in data",
			element:  "pear",
			data:     []string{"orange", "banana", "apple", "grape"},
			expected: -1,
		},
		{
			name:     "Element is empty",
			element:  "",
			data:     []string{"orange", "banana", "apple", "grape"},
			expected: -1,
		},
		{
			name:     "Data is empty",
			element:  "apple",
			data:     []string{},
			expected: -1,
		},
		{
			name:     "Duplicate elements in data",
			element:  "apple",
			data:     []string{"orange", "banana", "apple", "apple", "grape"},
			expected: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := common.SliceIndex(test.element, test.data)
			if result != test.expected {
				t.Errorf("Expected index %v, but got %v", test.expected, result)
			}
		})
	}
}

// Test case for function isIni
func TestIsIni(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "Single character ini",
			data:     []byte{'['},
			expected: true,
		},
		{
			name:     "First character is not a bracket",
			data:     []byte{'a', 'b', 'c'},
			expected: false,
		},
		{
			name:     "First character is a bracket in a larger data set",
			data:     []byte{'[', 'a', 'b', 'c'},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := common.IsIni(test.data)
			if result != test.expected {
				t.Errorf("For data %v, expected %t, but got %t", test.data, test.expected, result)
			}
		})
	}
}

// Test case for function isYaml
func TestIsYaml(t *testing.T) {
	// Test when valid YAML data is provided
	yamlData := []byte("key: value\n")
	if !common.IsYaml(yamlData) {
		t.Errorf("Expected isYaml to return true for valid YAML data, but it returned false")
	}

	// Test when invalid YAML data is provided
	invalidYamlData := []byte("key: value:")
	if common.IsYaml(invalidYamlData) {
		t.Errorf("Expected isYaml to return false for invalid YAML data, but it returned true")
	}

	// Test when empty data is provided
	emptyData := []byte("{")
	if common.IsYaml(emptyData) {
		t.Errorf("Expected isYaml to return false for empty data, but it returned true")
	}
}

// Test case for function isJson
func TestIsJson_ValidJson(t *testing.T) {
	data := []byte(`{"name": "John", "age": 30}`)
	result := common.IsJson(data)
	if !result {
		t.Errorf("Expected isJson to return true for valid JSON, but got false")
	}
}

func TestIsJson_InvalidJson(t *testing.T) {
	data := []byte(`{invalid_json}`)
	result := common.IsJson(data)
	if result {
		t.Errorf("Expected isJson to return false for invalid JSON, but got true")
	}
}

func TestIsJson_EmptyJson(t *testing.T) {
	data := []byte(`{}`)
	result := common.IsJson(data)
	if !result {
		t.Errorf("Expected isJson to return true for empty JSON, but got false")
	}
}

func TestIsJson_EmptyData(t *testing.T) {
	data := []byte(``)
	result := common.IsJson(data)
	if result {
		t.Errorf("Expected isJson to return false for empty data, but got true")
	}
}

// Test case for function isJson
func TestIsJson(t *testing.T) {
	// Test valid JSON data
	jsonData := []byte(`{"key": "value"}`)
	if !common.IsJson(jsonData) {
		t.Error("Expected true for valid JSON data")
	}

	// Test invalid JSON data
	invalidJsonData := []byte(`{"key": "value"`)
	if common.IsJson(invalidJsonData) {
		t.Error("Expected false for invalid JSON data")
	}

	// Test empty data
	emptyData := []byte("")
	if common.IsJson(emptyData) {
		t.Error("Expected false for empty data")
	}
}

func TestBuildFullSshCmdWithoutHost(t *testing.T) {
	sshCmd := "ssh -i key.pem user@"
	host := "example.com"
	expected := "ssh -i key.pem user@example.com"

	fullCmd, directorHost, _ := common.BuildFullSshCmd(sshCmd, host)

	if fullCmd != expected {
		t.Errorf("Unexpected full command, got: %s, want: %s", fullCmd, expected)
	}
	if directorHost != "example.com" {
		t.Errorf("Unexpected director host, got: %s, want: %s", directorHost, "example.com")
	}
}

func TestBuildFullSshCmdWithHost(t *testing.T) {
	sshCmd := "ssh -i key.pem user@example.com"
	host := "example.com"
	expected := "ssh -i key.pem user@example.com"

	fullCmd, directorHost, _ := common.BuildFullSshCmd(sshCmd, host)

	if fullCmd != expected {
		t.Errorf("Unexpected full command, got: %s, want: %s", fullCmd, expected)
	}
	if directorHost != "example.com" {
		t.Errorf("Unexpected director host, got: %s, want: %s", directorHost, "example.com")
	}
}

func TestBuildFullSshCmdWithOutDirectorHost(t *testing.T) {
	sshCmd := "ssh -i key.pem user@example.com"
	host := ""
	expected := "ssh -i key.pem user@example.com"

	fullCmd, directorHost, _ := common.BuildFullSshCmd(sshCmd, host)

	if fullCmd != expected {
		t.Errorf("Unexpected full command, got: %s, want: %s", fullCmd, expected)
	}
	if directorHost != "example.com" {
		t.Errorf("Unexpected director host, got: %s, want: %s", directorHost, "example.com")
	}
}

func TestBuildFullSshCmdWithWhiteSpaces(t *testing.T) {
	sshCmd := "ssh -F   config  "
	host := "example.com"
	expected := "ssh -F config example.com"

	fullCmd, directorHost, _ := common.BuildFullSshCmd(sshCmd, host)

	if fullCmd != expected {
		t.Errorf("Unexpected full command, got: %s, want: %s", fullCmd, expected)
	}
	if directorHost != "example.com" {
		t.Errorf("Unexpected director host, got: %s, want: %s", directorHost, "example.com")
	}
}

func TestBuildFullSshCmdEmptyHost(t *testing.T) {
	sshCmd := "ssh -F config example.com"
	host := ""
	expected := "ssh -F config example.com"

	fullCmd, directorHost, _ := common.BuildFullSshCmd(sshCmd, host)

	if fullCmd != expected {
		t.Errorf("Unexpected full command, got: %s, want: %s", fullCmd, expected)
	}
	if directorHost != "example.com" {
		t.Errorf("Unexpected director host, got: %s, want: %s", directorHost, "example.com")
	}
}

func TestBuildFullSshCmdWrongExtraArg(t *testing.T) {
	sshCmd := "ssh -F config example.com ls"
	host := ""

	_, _, err := common.BuildFullSshCmd(sshCmd, host)

	expectedError := "error: Too many arguments after -F option"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Unexpected error, got: %v, want: %s", err, expectedError)
	}
}

func TestBuildFullSshCmdWrongHost(t *testing.T) {
	sshCmd := "ssh -i key.pem root@foo"
	host := "example.com"

	_, _, err := common.BuildFullSshCmd(sshCmd, host)

	expectedError := "error: The host in the sshCmd foo does not match the directorHost example.com"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Unexpected error, got: %v, want: %s", err, expectedError)
	}
}

func TestBuildFullSshCmdWrongHostWithConfig(t *testing.T) {
	sshCmd := "ssh -F config foo"
	host := "example.com"

	_, _, err := common.BuildFullSshCmd(sshCmd, host)

	expectedError := "error: The host in the ssh_cmd: foo does not match the director_host: example.com"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Unexpected error, got: %v, want: %s", err, expectedError)
	}
}
