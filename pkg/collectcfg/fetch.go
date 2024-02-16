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

package collectcfg

import (
	"fmt"
	"os"
	"os-diff/pkg/common"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

var config Config

// Service YAML Config Structure
type Service struct {
	Enable             bool     `yaml:"enable"`
	PodmanName         string   `yaml:"podman_name"`
	PodName            string   `yaml:"pod_name"`
	ContainerName      string   `yaml:"container_name"`
	StrictPodNameMatch bool     `yaml:"strict_pod_name_match"`
	Path               []string `yaml:"path"`
}

type Config struct {
	Services map[string]Service `yaml:"services"`
}

func LoadServiceConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding YAML:", err)
		return err
	}
	return nil
}

func PullConfigs(configDir string, podman bool, sshCmd string) error {
	for service, _ := range config.Services {
		PullConfig(service, podman, configDir, sshCmd)
	}
	return nil
}

func PullConfig(serviceName string, podman bool, configDir string, sshCmd string) error {
	if podman {
		podmanId, _ := GetPodmanId(config.Services[serviceName].PodmanName, sshCmd)
		if len(strings.TrimSpace(podmanId)) > 0 {
			for _, path := range config.Services[serviceName].Path {
				dirPath := getDir(strings.TrimRight(path, "/"))
				PullPodmanFiles(podmanId, path, configDir+"/"+serviceName+"/"+dirPath, sshCmd)
			}
		} else {
			fmt.Println("Error, Podman name not found, skipping ..." + config.Services[serviceName].PodmanName)
		}
	} else {
		podId, _ := GetPodId(config.Services[serviceName].PodName)
		if len(strings.TrimSpace(podId)) > 0 {
			for _, path := range config.Services[serviceName].Path {
				PullPodFiles(podId, config.Services[serviceName].ContainerName, path, configDir+"/"+serviceName+"/"+path)
			}
		} else {
			fmt.Println("Error, Pod name not found, skipping ..." + config.Services[serviceName].PodName)
		}
	}
	return nil
}

func GetPodmanId(containerName string, sshCmd string) (string, error) {
	cmd := sshCmd + " podman ps -a | awk '/" + containerName + "$/  {print $1}'"
	output, err := common.ExecCmd(cmd)
	return output[0], err
}

func GetPodId(podName string) (string, error) {
	cmd := "oc get pods --field-selector status.phase=Running | awk '/" + podName + "-[a-f0-9-]/ {print $1}'"
	output, err := common.ExecCmd(cmd)
	return output[0], err
}

func PullPodmanFiles(podmanId string, remotePath string, localPath string, sshCmd string) error {
	cmd := sshCmd + " podman cp " + podmanId + ":" + remotePath + " " + localPath
	common.ExecCmd(cmd)
	return nil
}

func PullPodFiles(podId string, containerName string, remotePath string, localPath string) error {
	// Test OC connexion
	cmd := "oc cp -c " + containerName + " " + podId + ":" + remotePath + " " + localPath
	common.ExecCmd(cmd)
	return nil
}

func SyncConfigDir(localPath string, remotePath string, sshCmd string) error {
	cmd := "rsync -a -e '" + sshCmd + "' :" + remotePath + " " + localPath
	common.ExecCmd(cmd)
	return nil
}

func CleanUp(remotePath string, sshCmd string) error {
	if remotePath == "" || remotePath == "/" {
		return fmt.Errorf("Clean up Error - Empty or wrong path: " + remotePath + ". Please make sure you provided a correct path.")
	}
	cmd := sshCmd + " rm -rf " + remotePath
	common.ExecCmd(cmd)
	return nil
}

func CreateServicesTrees(configDir string, sshCmd string) (string, error) {
	for service, _ := range config.Services {
		for _, path := range config.Services[service].Path {
			output, err := CreateServiceTree(service, path, configDir, sshCmd)
			if err != nil {
				return output, err
			}
		}
	}
	return "", nil
}

func CreateServiceTree(serviceName string, path string, configDir string, sshCmd string) (string, error) {
	fullPath := configDir + "/" + serviceName + "/" + getDir(path)
	cmd := sshCmd + " mkdir -p " + fullPath
	output, err := common.ExecCmdSimple(cmd)
	return output, err
}

func getDir(s string) string {
	return path.Dir(s)
}

func FetchConfigFromEnv(configPath string,
	localDir string, remoteDir string, podman bool, connection, sshCmd string) error {

	var local bool
	err := LoadServiceConfig(configPath)
	if err != nil {
		return err
	}

	if connection == "local" {
		local = true
	} else {
		local = false
	}

	if local {
		output, err := CreateServicesTrees(localDir, sshCmd)
		if err != nil {
			fmt.Println(output)
			return err
		}
		PullConfigs(localDir, podman, sshCmd)
	} else {
		output, err := CreateServicesTrees(remoteDir, sshCmd)
		if err != nil {
			fmt.Println(output)
			return err
		}
		PullConfigs(remoteDir, podman, sshCmd)
		SyncConfigDir(localDir, remoteDir, sshCmd)
		CleanUp(remoteDir, sshCmd)
	}
	return nil
}
