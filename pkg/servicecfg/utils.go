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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/openstack-k8s-operators/os-diff/pkg/common"
	"github.com/openstack-k8s-operators/os-diff/pkg/godiff"
)

func CompareIniConfig(rawdata1 []byte, rawdata2 []byte, ocpConfig string, serviceConfig string) ([]string, error) {

	report, err := godiff.CompareIni(rawdata1, rawdata2, ocpConfig, serviceConfig, false)
	if err != nil {
		panic(err)
	}
	godiff.PrintReport(report)
	return report, nil
}

func GetConfigFromPod(serviceConfigPath string, podName string, containerName string) ([]byte, error) {

	if common.TestOCConnection() {
		fullName, err := GetPodFullName(podName)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command("oc", "exec", fullName, "-c", containerName, "--", "cat", serviceConfigPath)
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

func GenerateOpenShiftConfig(outputConfigPath string, serviceConfigPath string) error {
	return nil
}

func GetPodFullName(podName string) (string, error) {
	// Get full pod name
	cmd := "oc get pod | grep " + podName + " | grep -i running | cut -f 1 -d' '"
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return string(output), err
	}
	return string(output[:len(output)-1]), nil
}

func GetOCConfigMap(configMapName string) ([]byte, error) {
	if common.TestOCConnection() {
		// Get full pod name
		cmd := "oc get configmap/" + configMapName + " -o yaml"
		output, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			return output, err
		}
		return output, nil
	}
	return nil, fmt.Errorf("OC is not connected, you need to logged in before.")
}

func RemoteStatDir(sshCmd string, path string) (bool, error) {
	// Get full pod name
	cmd := sshCmd + " stat --printf='%F' " + path
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return false, err
	}
	if strings.Contains("regular file", string(output)) {
		return false, nil
	} else if strings.Contains("directory", string(output)) {
		return true, nil
	}
	return false, fmt.Errorf("Unable to stat: ", path)
}

func LoadServiceConfig(file string) ([]byte, error) {
	serviceConfig, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return serviceConfig, nil
}

func LoadFilesIntoMap(fileName string) (map[string]string, error) {
	result := make(map[string]string)

	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			parts = strings.SplitN(line, ":", 2)
		}
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
