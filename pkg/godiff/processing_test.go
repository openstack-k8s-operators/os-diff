package godiff

import (
	"testing"
	"os"
	"github.com/sirupsen/logrus"
	"bytes"
)

// Test case for function init
func TestInit(t *testing.T) {
	// Test init function with successful file open
	os.Remove("results.log") // Remove any existing log file
	init()
	_, err := os.Stat("results.log")
	if err != nil {
		t.Errorf("results.log file not created: %v", err)
	}

	// Test init function with failed file open
	fakeFile, _ := os.Create("fakefile.log") // Create a file without write permission
	fakeFile.Chmod(0)
	fakeFile.Close()

	os.Remove("results.log") // Remove results.log created in previous test case
	init() // Try to open file with permission issue
	_, err = os.Stat("results.log")
	if err == nil {
		t.Errorf("results.log file created despite error: %v", err)
	}

	os.Remove("fakefile.log") // Clean up created fake file
}

func TestLogOutput(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	init()
	logrus.Info("test message")

	expected := "INFO[0000] test message\n"
	if buf.String() != expected {
		t.Errorf("Log output does not match expected. Got: %s, want: %s", buf.String(), expected)
	}
}

// Test case for function filesEqual
import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/your-package-path/godiff" // Import the package where the filesEqual function is defined
)

func TestFilesEqual(t *testing.T) {
	// Create temporary test files with same content
	file1 := "file1.txt"
	file2 := "file2.txt"
	content := []byte("test content")
	err := os.WriteFile(file1, content, 0644)
	if err != nil {
		t.Fatalf("Error creating file1: %v", err)
	}
	defer os.Remove(file1)
	err = os.WriteFile(file2, content, 0644)
	if err != nil {
		t.Fatalf("Error creating file2: %v", err)
	}
	defer os.Remove(file2)

	// Test filesEqual function with identical files
	result, err := godiff.filesEqual(file1, file2)
	assert.NoError(t, err)
	assert.True(t, result)

	// Modify content of file2 to make files different
	err = os.WriteFile(file2, []byte("different content"), 0644)
	if err != nil {
		t.Fatalf("Error modifying file2: %v", err)
	}

	// Test filesEqual function with different files
	result, err = godiff.filesEqual(file1, file2)
	assert.NoError(t, err)
	assert.False(t, result)
}

// Test case for function checkFile
func TestCheckFile(t *testing.T) {
	// Test case where files are equal
	equalPath := "testdata/equal.txt"
	if _, err := os.Create(equalPath); err != nil {
		t.Fatalf("error creating test file: %v", err)
	}
	defer os.Remove(equalPath)

	if _, err := os.OpenFile(equalPath, os.O_RDWR, os.ModeAppend); err != nil {
		t.Fatalf("error opening test file: %v", err)
	}

	equalResult, equalErr := checkFile(equalPath, equalPath)
	if equalErr != nil {
		t.Fatalf("unexpected error: %v", equalErr)
	}
	if equalResult != true {
		t.Fatalf("expected files to be equal, got false")
	}

	// Test case where files are not equal
	unequalPath1 := "testdata/unequal1.txt"
	if _, err := os.Create(unequalPath1); err != nil {
		t.Fatalf("error creating test file: %v", err)
	}
	defer os.Remove(unequalPath1)

	unequalPath2 := "testdata/unequal2.txt"
	if _, err := os.Create(unequalPath2); err != nil {
		t.Fatalf("error creating test file: %v", err)
	}
	defer os.Remove(unequalPath2)

	unequalResult, unequalErr := checkFile(unequalPath1, unequalPath2)
	if unequalErr != nil {
		t.Fatalf("unexpected error: %v", unequalErr)
	}
	if unequalResult != false {
		t.Fatalf("expected files to be unequal, got true")
	}

	// Test case where one file path is a directory
	dirPath := "testdata"
	dirResult, dirErr := checkFile(equalPath, dirPath)
	if dirErr == nil {
		t.Fatalf("expected error for directory path, got nil")
	}
}

