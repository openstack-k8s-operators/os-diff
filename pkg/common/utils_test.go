package common

import (
	"testing"
	"reflect"
	"github.com/stretchr/testify/assert"
)

// Test case for function ExecCmd
func TestExecCmd(t *testing.T) {
	testCases := []struct {
		cmd          string
		expectedOut  []string
		expectedErr  error
	}{
		{
			cmd:         "echo hello",
			expectedOut: []string{"hello"},
			expectedErr: nil,
		},
		{
			cmd:         "ls -l",
			expectedOut: []string{"", "...", "file1.txt", "file2.txt", "...", ""},
			expectedErr: nil,
		},
		{
			// Edge case: empty command
			cmd:         "",
			expectedOut: []string{""},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		output, err := ExecCmd(tc.cmd)

		// Check if the output matches the expected output
		for i, line := range output {
			if line != tc.expectedOut[i] {
				t.Errorf("ExecCmd(%s) returned incorrect output, got: %v, want: %v", tc.cmd, output, tc.expectedOut)
				break
			}
		}

		// Check if the error matches the expected error
		if err != tc.expectedErr {
			t.Errorf("ExecCmd(%s) returned incorrect error, got: %v, want: %v", tc.cmd, err, tc.expectedErr)
		}
	}
}

// Test case for function ExecCmdSimple
func TestExecCmdSimple(t *testing.T) {
	testCases := []struct {
		name     string
		cmd      string
		expected string
		err      error
	}{
		{"Valid command", "echo 'Hello, World!'", "Hello, World!\n", nil},
		{"Invalid command", "invalid_command", "", err},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ExecCmdSimple(tc.cmd)

			if output != tc.expected {
				t.Errorf("Expected output: %s, but got: %s", tc.expected, output)
			}

			if !reflect.DeepEqual(err, tc.err) {
				t.Errorf("Expected error: %v, but got: %v", tc.err, err)
			}
		})
	}
}

// Test case for function TestOCConnection
func TestTestOCConnection_Success(t *testing.T) {
	// Arrange

	// Act
	result := common.TestOCConnection()

	// Assert
	assert.True(t, result, "Expected TestOCConnection to return true for successful connection")
}

func TestTestOCConnection_Failure(t *testing.T) {
	// Arrange

	// Act
	result := common.TestOCConnection()

	// Assert
	assert.False(t, result, "Expected TestOCConnection to return false for failed connection")
}

// Test case for function TestSshConnection
func TestTestSshConnection(t *testing.T) {
	tests := []struct {
		name     string
		sshCmd   string
		expected bool
	}{
		{
			name:     "Valid SSH command",
			sshCmd:   "ssh user@hostname",
			expected: true,
		},
		{
			name:     "Invalid SSH command",
			sshCmd:   "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestSshConnection(tt.sshCmd)
			if result != tt.expected {
				t.Errorf("Expected %v, but got %v", tt.expected, result)
			}
		})
	}
}

// Test case for function TestOCConnection
func TestTestOCConnection_Success(t *testing.T) {
	testSuccessExecCmd := func(cmd string) ([]byte, error) {
		return []byte("user123"), nil
	}

	ExecCmd = testSuccessExecCmd

	result := TestOCConnection()

	if !result {
		t.Errorf("Expected TestOCConnection to return true for successful connection, but got false")
	}
}

func TestTestOCConnection_Failure(t *testing.T) {
	testFailureExecCmd := func(cmd string) ([]byte, error) {
		return nil, fmt.Errorf("Error executing cmd")
	}

	ExecCmd = testFailureExecCmd

	result := TestOCConnection()

	if result {
		t.Errorf("Expected TestOCConnection to return false for failed connection, but got true")
	}
}

// Test case for function TestOCConnection
func TestTestOCConnection_Success(t *testing.T) {
	mockExecCmdSuccess := func(cmd string) (string, error) {
		return "user123", nil
	}

	ExecCmd = mockExecCmdSuccess

	result := TestOCConnection()
	if result != true {
		t.Error("Expected true, got false")
	}
}

func TestTestOCConnection_Failure(t *testing.T) {
	mockExecCmdFailure := func(cmd string) (string, error) {
		return "", errors.New("error executing command")
	}

	ExecCmd = mockExecCmdFailure
	
	result := TestOCConnection()
	if result != false {
		t.Error("Expected false, got true")
	}
}

// Test case for function TestOCConnection
func TestTestOCConnection_Success(t *testing.T) {
	// Mocking the ExecCmd function to return nil error
	oldExecCmd := ExecCmd
	defer func() { ExecCmd = oldExecCmd }()
	ExecCmd = func(cmd string) (string, error) {
		return "test_user", nil
	}

	result := TestOCConnection()
	if !result {
		t.Errorf("TestOCConnection returned false for a successful command execution")
	}
}

func TestTestOCConnection_Error(t *testing.T) {
	// Mocking the ExecCmd function to return an error
	oldExecCmd := ExecCmd
	defer func() { ExecCmd = oldExecCmd }()
	ExecCmd = func(cmd string) (string, error) {
		return "", fmt.Errorf("Error executing command")
	}

	result := TestOCConnection()
	if result {
		t.Errorf("TestOCConnection returned true for an error scenario")
	}
}

