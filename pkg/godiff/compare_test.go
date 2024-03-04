package godiff

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"fmt"
	"reflect"
	"github.com/stretchr/testify/assert"
	"encoding/json"
	"bytes"
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
```

// Test case for function CompareFiles
func TestCompareFiles(t *testing.T) {
	// Create temporary test files
	originContent := []byte("test content for origin file")
	destContent := []byte("test content for destination file")
	origin := "test_origin.txt"
	dest := "test_dest.txt"

	err := ioutil.WriteFile(origin, originContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(origin)

	err = ioutil.WriteFile(dest, destContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(dest)

	// Test CompareFiles function
	report, err := CompareFiles(origin, dest, false, false)
	if err != nil {
		t.Fatalf("Error while running CompareFiles: %v", err)
	}

	// Check if report contains expected differences or not
	if len(report) != 0 {
		t.Errorf("Expected no differences, found differences in the files")
	}

	// Test with verbose and print flag enabled
	report, err = CompareFiles(origin, dest, true, true)
	if err != nil {
		t.Fatalf("Error while running CompareFiles: %v", err)
	}

	// Check if report contains expected differences or not
	if len(report) != 0 {
		t.Errorf("Expected no differences, found differences in the files while printing with verbose")
	}

	// Test with non-existent files
	_, err = CompareFiles("nonexistent_file1.txt", "nonexistent_file2.txt", false, false)
	if err == nil {
		t.Fatalf("Expected error for non-existent files, got nil")
	}
}

// Test case for function CompareFilesFromRemote
func TestCompareFilesFromRemote(t *testing.T) {
	origin := "origin.txt"
	dest := "dest.txt"
	originRemoteCmd := "ssh user@server cat origin.txt"
	destRemoteCmd := "ssh user@server cat dest.txt"
	verbose := true

	err := CompareFilesFromRemote(origin, dest, originRemoteCmd, destRemoteCmd, verbose)
	if err != nil {
		t.Errorf("Error running CompareFilesFromRemote: %v", err)
	}

	// Add additional test cases for edge cases and scenarios

	// Test case where GetConfigFromRemote returns an error
	err = CompareFilesFromRemote("invalidOrigin", "invalidDest", originRemoteCmd, destRemoteCmd, verbose)
	if err == nil {
		t.Error("Expected an error but got nil")
	}
}

