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
	"strings"
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
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			return out, err
		}
		return []byte(out), nil

	} else {
		return nil, fmt.Errorf("OC is not connected, you need to logged in before.")
	}
}

func GetConfigFromPodman(serviceConfigPath string, podmanName string) ([]byte, error) {

	cmd := exec.Command("ssh", "-F", "ssh.config", "standalone", "podman", "exec", podmanName, "cat ", serviceConfigPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return out, err
	}
	return []byte(out), nil
}

func GenerateOpenshiftConfig(outputConfigPath string, serviceConfigPath string) error {
	return nil
}

func TestOCConnection() bool {
	cmd := exec.Command("oc", "whoami")
	_, err := cmd.Output()
	if err != nil {
		return false
	}
	return true
}

func LoadServiceConfig(file string) ([]byte, error) {
	serviceConfig, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return serviceConfig, nil
}

func cleanIniSections(config string) string {
	lines := strings.Split(config, "\n")
	sectionMap := make(map[string][]string)
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check if line is a section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
			continue
		}
		// Skip empty lines or lines without '='
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		// Append key-value pairs to section map
		if currentSection != "" {
			sectionMap[currentSection] = append(sectionMap[currentSection], line)
		}
	}
	var sb strings.Builder
	// Build updated INI string
	for section, lines := range sectionMap {
		sb.WriteString(fmt.Sprintf("[%s]\n", section))
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("%s\n", line))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
