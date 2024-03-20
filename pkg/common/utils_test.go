package common

import (
	"reflect"
	"testing"
)

// Test case for function ExecCmd
func TestExecCmd(t *testing.T) {
	testCases := []struct {
		cmd         string
		expectedOut []string
		expectedErr error
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
