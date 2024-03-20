package godiff

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test case for function writeReport
func TestWriteReport(t *testing.T) {
	content := []string{"line 1", "line 2", "line 3"}
	reportPath := "./test_report.txt"

	err := writeReport(content, reportPath)

	assert.NoError(t, err)

	// Check if the file was created
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Errorf("Report file was not created: %v", err)
	}

	// Check if the file content matches the expected content
	fileContent, err := ioutil.ReadFile(reportPath)
	if err != nil {
		t.Errorf("Failed to read report file: %v", err)
	}
	expectedContent := "line 1line 2line 3"
	if string(fileContent) != expectedContent {
		t.Errorf("Report file content does not match. Expected: %s, Actual: %s", expectedContent, string(fileContent))
	}

	// Cleanup: remove the test report file
	err = os.Remove(reportPath)
	if err != nil {
		t.Errorf("Failed to remove test report file: %v", err)
	}
}

// Test case for function PrintReport
func TestPrintReport(t *testing.T) {
	tests := []struct {
		name   string
		report []string
	}{
		{
			name:   "Empty report",
			report: []string{},
		},
		{
			name:   "Report with lines starting with +",
			report: []string{"+line1", "line2", "+line3"},
		},
		{
			name:   "Report with lines starting with -",
			report: []string{"-line1", "line2", "-line3"},
		},
		{
			name:   "Report with normal lines",
			report: []string{"line1", "line2", "line3"},
		},
		{
			name:   "Report with mix of lines",
			report: []string{"-line1", "+line2", "line3", "-line4"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := PrintReport(test.report)
			if err != nil {
				t.Errorf("PrintReport() error = %v, want nil", err)
			}
		})
	}
}

// Test case for function CompareJSONFiles
func TestCompareJSONFiles(t *testing.T) {
	originJSON := []byte(`{"key1": "value1"}`)
	destJSON := []byte(`{"key1": "value1"}`)

	expectedReport := []string{}
	expectedErr := error(nil)

	report, err := CompareJSONFiles(originJSON, destJSON)

	if !reflect.DeepEqual(report, expectedReport) {
		t.Errorf("Report does not match expected value. Got: %v, Want: %v", report, expectedReport)
	}

	if err != expectedErr {
		t.Errorf("Error does not match expected value. Got: %v, Want: %v", err, expectedErr)
	}

	// Test unmarshalling error case
	invalidJSON := []byte(`{"key1": "value1"}`)
	_, err = CompareJSONFiles(invalidJSON, destJSON)
	if err == nil {
		t.Errorf("Expected error when unmarshalling invalid JSON but got nil")
	}
}
