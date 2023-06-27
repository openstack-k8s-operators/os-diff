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
package ansible

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type DefaultExecute struct {
	Write io.Writer
}

// Execute takes a command and args and runs it, streaming output to stdout
func (e *DefaultExecute) Execute(command string, args []string, prefix string) error {

	stderr := &bytes.Buffer{}

	if e.Write == nil {
		return errors.New("(DefaultExecute::Execute) A writer must be defined")
	}

	cmd := exec.Command(command, args...)
	cmd.Stderr = stderr

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return errors.New("(DefaultExecute::Execute) -> " + err.Error())
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Fprintf(e.Write, "%s =>  %s\n", prefix, scanner.Text())
		}
	}()

	timeInit := time.Now()
	err = cmd.Start()
	if err != nil {
		return errors.New("(DefaultExecute::Execute) -> " + err.Error())
	}

	err = cmd.Wait()
	elapsedTime := time.Since(timeInit)
	if err != nil {
		return errors.New("(DefaultExecute::Execute) -> " + stderr.String())
	}

	fmt.Fprintf(e.Write, "Duration: %s\n", elapsedTime.String())

	return nil
}
