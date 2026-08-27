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
package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestRetirementMessage(t *testing.T) {
	msg := retirementMessage()
	if !strings.Contains(msg, "retired") || !strings.Contains(msg, "no longer supported") {
		t.Fatalf("unexpected message: %q", msg)
	}
}

func TestMainExitsOne(t *testing.T) {
	if os.Getenv("OS_DIFF_TEST_MAIN") == "1" {
		main()
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainExitsOne")
	cmd.Env = append(os.Environ(), "OS_DIFF_TEST_MAIN=1")
	err := cmd.Run()
	if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 1 {
		t.Fatalf("expected exit 1, got %v", err)
	}
}
