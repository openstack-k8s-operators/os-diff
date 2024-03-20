package godiff

import (
	"bytes"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogOutput(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.Info("test message")

	expected := "INFO[0000] test message\n"
	if buf.String() != expected {
		t.Errorf("Log output does not match expected. Got: %s, want: %s", buf.String(), expected)
	}
}

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
	result, err := filesEqual(file1, file2)
	assert.NoError(t, err)
	assert.True(t, result)

	// Modify content of file2 to make files different
	err = os.WriteFile(file2, []byte("different content"), 0644)
	if err != nil {
		t.Fatalf("Error modifying file2: %v", err)
	}

	// Test filesEqual function with different files
	result, err = filesEqual(file1, file2)
	assert.NoError(t, err)
	assert.False(t, result)
}
