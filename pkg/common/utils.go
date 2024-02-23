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

package common

import (
	"os/exec"
	"strings"
)

// Shell execution functions:
func ExecCmd(cmd string) ([]string, error) {
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return strings.Split(string(output), "\n"), err
	}
	return strings.Split(string(output), "\n"), nil
}

func ExecCmdSimple(cmd string) (string, error) {
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

func TestOCConnection() bool {
	cmd := "oc whoami"
	_, err := ExecCmd(cmd)
	if err != nil {
		return false
	}
	return true
}

func TestSshConnection(sshCmd string) bool {
	cmd := sshCmd + " ls"
	_, err := ExecCmd(cmd)
	if err != nil {
		return false
	}
	return true
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
