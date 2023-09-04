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
package servicecfg

import (
	"fmt"
	"io/ioutil"
	"os-diff/pkg/godiff"
	"os/exec"
)

func CompareIniConfig(rawdata1 []byte, rawdata2 []byte, ocpConfig string, serviceConfig string) ([]string, error) {

	report, err := godiff.CompareIni(rawdata1, rawdata2, ocpConfig, serviceConfig, false)
	if err != nil {
		panic(err)
	}
	godiff.PrintReport(report)
	return report, nil
}

func GetConfigFromPod(serviceConfigPath string, podname string) ([]byte, error) {

	if TestOCConnection() {
		cmd := exec.Command("oc", "exec", podname, "--", "cat", serviceConfigPath)
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return []byte(out), nil

	} else {
		return nil, fmt.Errorf("OC is not connected, you need to logged in before.")
	}
}

func GenerateOpenshiftConfig(outputConfigPath string, serviceConfigPath string) error {
	return nil
}

func TestOCConnection() bool {
	cmd := exec.Command("oc", "whoiam")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	fmt.Println("Output: ", string(out))
	return true
}

func LoadServiceConfig(file string) ([]byte, error) {
	serviceConfig, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return serviceConfig, nil
}
