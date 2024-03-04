package common

import (
	"testing"
	"os"
	"io/ioutil"
	"path/filepath"
	"github.com/stretchr/testify/assert"
)

// Test case for function LoadOSDiffConfig
func TestLoadOSDiffConfig(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_config.ini")
	
	// Create a sample config file
	configData := []byte(`
[ODConfig]
key1 = value1
key2 = value2
`)
	if err := ioutil.WriteFile(tempFile, configData, 0644); err != nil {
		t.Fatalf("Error creating temp config file: %v", err)
	}

	// Test case for successful loading of config file
	expectedConfig := ODConfig{Key1: "value1", Key2: "value2"}
	actualConfig, err := LoadOSDiffConfig(tempFile)
	assert.Nil(t, err, "Expected no error, but got: %v", err)
	assert.Equal(t, expectedConfig, *actualConfig, "Loaded config does not match expected config")

	// Test case for error scenario - file not found
	_, err = LoadOSDiffConfig("non_existent_file.ini")
	assert.NotNil(t, err, "Expected error, but got nil")

	// Test case for error scenario - invalid file format
	invalidFile := filepath.Join(tempDir, "invalid_file.txt")
	_, err = os.Create(invalidFile)
	if err != nil {
		t.Fatalf("Error creating invalid temp file: %v", err)
	}
	_, err = LoadOSDiffConfig(invalidFile)
	assert.NotNil(t, err, "Expected error, but got nil")
}

