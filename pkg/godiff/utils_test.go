package godiff

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
			result := stringInSlice(tc.inputStr, tc.inputList)
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
			result := sliceIndex(test.element, test.data)
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
			result := isIni(test.data)
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
	if !isYaml(yamlData) {
		t.Errorf("Expected isYaml to return true for valid YAML data, but it returned false")
	}

	// Test when invalid YAML data is provided
	invalidYamlData := []byte("key: value:")
	if isYaml(invalidYamlData) {
		t.Errorf("Expected isYaml to return false for invalid YAML data, but it returned true")
	}

	// Test when empty data is provided
	emptyData := []byte("")
	if isYaml(emptyData) {
		t.Errorf("Expected isYaml to return false for empty data, but it returned true")
	}
}

// Test case for function isJson
func TestIsJson_ValidJson(t *testing.T) {
	data := []byte(`{"name": "John", "age": 30}`)
	result := isJson(data)
	if !result {
		t.Errorf("Expected isJson to return true for valid JSON, but got false")
	}
}

func TestIsJson_InvalidJson(t *testing.T) {
	data := []byte(`{invalid_json}`)
	result := isJson(data)
	if result {
		t.Errorf("Expected isJson to return false for invalid JSON, but got true")
	}
}

func TestIsJson_EmptyJson(t *testing.T) {
	data := []byte(`{}`)
	result := isJson(data)
	if !result {
		t.Errorf("Expected isJson to return true for empty JSON, but got false")
	}
}

func TestIsJson_EmptyData(t *testing.T) {
	data := []byte(``)
	result := isJson(data)
	if result {
		t.Errorf("Expected isJson to return false for empty data, but got true")
	}
}

// Test case for function isJson
func TestIsJson(t *testing.T) {
	// Test valid JSON data
	jsonData := []byte(`{"key": "value"}`)
	if !isJson(jsonData) {
		t.Error("Expected true for valid JSON data")
	}

	// Test invalid JSON data
	invalidJsonData := []byte(`{"key": "value"`)
	if isJson(invalidJsonData) {
		t.Error("Expected false for invalid JSON data")
	}

	// Test empty data
	emptyData := []byte("")
	if isJson(emptyData) {
		t.Error("Expected false for empty data")
	}
}

// Test case for function CompareYAML
func TestCompareYAML(t *testing.T) {
	origin := []byte(`
		key1: value1
		key2: value2
	`)
	dest := []byte(`
		key1: value1
		key3: value3
	`)
	expected := []string{
		"+key2: value2",
		"-key3: value3",
	}
	report, err := CompareYAML(origin, dest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(report, expected) {
		t.Errorf("Report mismatch, got: %v, want: %v", report, expected)
	}
}

func TestCompareYAMLEqualMaps(t *testing.T) {
	origin := []byte(`
		key1: value1
		key2: value2
	`)
	dest := []byte(`
		key1: value1
		key2: value2
	`)
	report, err := CompareYAML(origin, dest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(report) > 0 {
		t.Errorf("Report should be empty for equal maps, got: %v", report)
	}
}

//func TestCompareYAMLInvalidYAML(t *testing.T) {
//	origin := []byte(`
//		key1: value1
//		key2: value2
//	`)
//	dest := []byte(`
//		key1: value1
//		key2: value2
//		invalid: yaml
//	`)
//	expectedError := "Error unmarshalling"
//	report, err := CompareYAML(origin, dest)
//	if err == nil || !reflect.DeepEqual(err.Error(), expectedError) {
//		t.Errorf("Expected error: %v, got: %v", expectedError, err)
//	}
//}

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
	diff, err := CompareJSON(orgData, destData, "")

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
	diff2, err2 := CompareJSON(orgData2, destData2, "")

	// Assertion for differences and no error
	assert.NoError(t, err2)
	assert.NotEmpty(t, diff2)
	assert.Len(t, diff2, 2)
	assert.Contains(t, diff2, "+key4")
	assert.Contains(t, diff2, "-key3")

	// Test case for comparing JSON objects with type mismatch
	orgData3 := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	destData3 := []interface{}{"value1", 123}
	diff3, err3 := CompareJSON(orgData3, destData3, "")

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
		"@ line: 1\n+line2\n",
		"@ line: 2\n-line3\n+line4\n",
		"@@ line: 4\n-line5\n",
	}

	report, err := CompareRawData(rawdata1, rawdata2, origin, dest)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(report, expectedReport) {
		t.Errorf("Report mismatch, expected: %v, got: %v", expectedReport, report)
	}
}

// Test case for function GetConfigFromRemote
func TestGetConfigFromRemote(t *testing.T) {
	remoteCmd := "ssh user@hostname"     // Change this according to the remote command
	configPath := "/path/to/config/file" // Change this according to the config file path

	// Test case for successful execution
	t.Run("Success", func(t *testing.T) {
		_, err := GetConfigFromRemote(remoteCmd, configPath)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test case for error scenario
	t.Run("Error", func(t *testing.T) {
		invalidCmd := "invalidCmd"
		_, err := GetConfigFromRemote(invalidCmd, configPath)
		if err == nil {
			t.Error("Expected an error, got nil")
		}
	})
}
