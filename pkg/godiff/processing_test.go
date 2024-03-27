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
	"os"
	"testing"

	"github.com/openstack-k8s-operators/os-diff/pkg/godiff"
	"github.com/stretchr/testify/assert"
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
	result, err := godiff.FilesEqual(file1, file2)
	assert.NoError(t, err)
	assert.True(t, result)

	// Modify content of file2 to make files different
	err = os.WriteFile(file2, []byte("different content"), 0644)
	if err != nil {
		t.Fatalf("Error modifying file2: %v", err)
	}

	// Test filesEqual function with different files
	result, err = godiff.FilesEqual(file1, file2)
	assert.NoError(t, err)
	assert.False(t, result)
}
