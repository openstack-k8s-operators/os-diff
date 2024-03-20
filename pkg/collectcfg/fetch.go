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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/openstack-k8s-operators/os-diff/pkg/common"

	"gopkg.in/yaml.v3"
)

var config common.Config

// TripleO information structures:
type PodmanContainer struct {
	Image string   `json:"Image"`
	ID    string   `json:"ID"`
	Names []string `json:"Names"`
}

func dumpConfigFile(configPath string) error {
	// Write updated data to config.yaml file
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, yamlData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func PullConfigs(configDir string, tripleo bool, sshCmd string, undercloud string, filters []string) error {
	// Pull configuration service by service
	filterMap := make(map[string]struct{})
	for _, filter := range filters {
		filterMap[filter] = struct{}{}
	}
	for service := range config.Services {
		if _, ok := filterMap[service]; ok || len(filters) == 0 {
			if tripleo && (config.Services[service].PodmanName == "" || config.Services[service].PodmanId == "") {
				PullConfigFromHosts(service, configDir, sshCmd, undercloud)
			} else {
				fullCmd := sshCmd + " " + undercloud
				PullConfig(service, tripleo, configDir, fullCmd)
			}
		}
	}
	return nil
}

func PullConfig(serviceName string, tripleo bool, configDir string, sshCmd string) error {
	// Pull configuration from TripleO Podman or OCP Pods
	if tripleo {
		var podmanId string
		if config.Services[serviceName].PodmanId != "" {
			podmanId = config.Services[serviceName].PodmanId
		} else {
			podmanId, _ = GetPodmanId(config.Services[serviceName].PodmanName, sshCmd)
		}
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

func GetPodmanIds(sshCmd string, all bool) ([]byte, error) {
	var cmd string
	if all {
		cmd = sshCmd + " podman ps -a --format json"
	} else {
		cmd = sshCmd + " podman ps --format json"
	}
	output, err := exec.Command("bash", "-c", cmd).Output()
	return output, err
}

func PullConfigFromHosts(service string, configDir string, sshCmd string, undercloud string) error {
	// Pull confugiration for a given service non hosted on Podman and OCP containers
	if len(config.Services[service].Hosts) != 0 {
		// if the services are not on the Undercloud/Director node
		for _, h := range config.Services[service].Hosts {
			fullCmd := sshCmd + " " + h
			// check if its config files or command output
			if config.Services[service].ServiceCommand != "" && config.Services[service].CatOutput {
				for _, path := range config.Services[service].Path {
					GetCommandOutput(config.Services[service].ServiceCommand, configDir+"/"+service+"/"+h+"/"+path, fullCmd)
				}
			} else {
				// else if config files
				for _, path := range config.Services[service].Path {
					PullLocalFiles(path, configDir+"/"+service+"/"+h+"/"+path, fullCmd)
				}
			}
		}
	} else {
		fullCmd := sshCmd + " " + undercloud
		// check if its config files or command output
		if config.Services[service].ServiceCommand != "" && config.Services[service].CatOutput {
			for _, path := range config.Services[service].Path {
				GetCommandOutput(config.Services[service].ServiceCommand, configDir+"/"+service+"/"+undercloud+"/"+path, fullCmd)
			}
		} else {
			// else if config files
			for _, path := range config.Services[service].Path {
				PullLocalFiles(path, configDir+"/"+service+"/"+undercloud+"/"+path, fullCmd)
			}
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

func GetCommandOutput(command string, localPath string, sshCmd string) error {
	cmd := sshCmd + " " + command + " > " + localPath
	output, err := common.ExecComplexCmd(cmd)
	if err != nil {
		fmt.Println(output)
		return err
	}
	fmt.Println(output)
	return nil
}

func PullLocalFiles(orgPath string, destPath string, sshCmd string) error {
	cmd := sshCmd + " cp -R " + orgPath + " " + destPath
	_, err := common.ExecCmd(cmd)
	if err != nil {
		return err
	}
	return nil
}

func PullPodmanFiles(podmanId string, remotePath string, localPath string, sshCmd string) error {
	cmd := sshCmd + " podman cp " + podmanId + ":" + remotePath + " " + localPath
	_, err := common.ExecCmd(cmd)
	if err != nil {
		return err
	}
	return nil
}

func PullPodFiles(podId string, containerName string, remotePath string, localPath string) error {
	// Test OC connexion
	cmd := "oc cp -c " + containerName + " " + podId + ":" + remotePath + " " + localPath
	_, err := common.ExecCmd(cmd)
	if err != nil {
		return err
	}
	return nil
}

func SyncConfigDir(localPath string, remotePath string, sshCmd string, undercloud string) error {
	hosts := GetListHosts(undercloud)
	for _, h := range hosts {
		cmd := "rsync -a -e '" + sshCmd + " " + h + "' :" + remotePath + " " + localPath
		common.ExecCmd(cmd)
	}
	return nil
}

func GetListHosts(undercloud string) []string {
	var hosts []string
	hosts = append(hosts, undercloud)
	for service := range config.Services {
		for _, h := range config.Services[service].Hosts {
			if !common.StringInSlice(h, hosts) {
				hosts = append(hosts, h)
			}
		}
	}
	return hosts
}

func CleanUp(remotePath string, sshCmd string) error {
	if remotePath == "" || remotePath == "/" {
		return fmt.Errorf("Clean up Error - Empty or wrong path: " + remotePath + ". Please make sure you provided a correct path.")
	}
	cmd := sshCmd + " rm -rf " + remotePath
	common.ExecCmd(cmd)
	return nil
}

func CreateServicesTrees(configDir string, sshCmd string, undercloud string, filters []string) (string, error) {
	filterMap := make(map[string]struct{})
	for _, filter := range filters {
		filterMap[filter] = struct{}{}
	}

	for service := range config.Services {
		if _, ok := filterMap[service]; ok || len(filters) == 0 {
			if len(config.Services[service].Hosts) != 0 {
				for _, h := range config.Services[service].Hosts {
					// Create trees for each hosts describe in config Yaml file
					fullCmd := sshCmd + " " + h
					for _, path := range config.Services[service].Path {
						output, err := CreateServiceTree(service, path, configDir, fullCmd, h)
						if err != nil {
							return output, err
						}
					}
				}
			} else {
				fullCmd := sshCmd + " " + undercloud
				for _, path := range config.Services[service].Path {
					output, err := CreateServiceTree(service, path, configDir, fullCmd, "")
					if err != nil {
						return output, err
					}
				}
			}
		}
	}
	return "", nil
}

func CreateServiceTree(serviceName string, path string, configDir string, sshCmd string, host string) (string, error) {
	fullPath := configDir + "/" + serviceName + "/" + host + "/" + getDir(path)
	cmd := sshCmd + " mkdir -p " + fullPath
	output, err := common.ExecCmdSimple(cmd)
	return output, err
}

func getDir(s string) string {
	return path.Dir(s)
}

func FetchConfigFromEnv(configPath string,
	localDir string, remoteDir string, tripleo bool, connection, sshCmd string, undercloud string, filters []string) error {

	var local bool
	cfg, err := common.LoadServiceConfigFile(configPath)
	if err != nil {
		return err
	}
	config = cfg

	if connection == "local" {
		local = true
	} else {
		local = false
	}

	if local {
		output, err := CreateServicesTrees(localDir, sshCmd, undercloud, filters)
		if err != nil {
			fmt.Println(output)
			return err
		}
		PullConfigs(localDir, tripleo, sshCmd, undercloud, filters)
	} else {
		output, err := CreateServicesTrees(remoteDir, sshCmd, undercloud, filters)
		if err != nil {
			fmt.Println(output)
			return err
		}
		PullConfigs(remoteDir, tripleo, sshCmd, undercloud, filters)
		SyncConfigDir(localDir, remoteDir, sshCmd, undercloud)
		CleanUp(remoteDir, sshCmd)
	}
	return nil
}

func buildPodmanInfo(output []byte, filters []string) (map[string]map[string]string, error) {

	filterMap := make(map[string]struct{})
	for _, filter := range filters {
		filterMap[filter] = struct{}{}
	}
	var containers []PodmanContainer
	err := json.Unmarshal(output, &containers)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	data := make(map[string]map[string]string)
	for _, container := range containers {
		for _, name := range container.Names {
			if _, ok := filterMap[name]; ok || len(filters) == 0 {
				data[name] = map[string]string{
					"containerid": container.ID[:12],
					"image":       container.Image,
				}
			}
		}
	}
	return data, nil
}

func SetTripleODataEnv(configPath string, sshCmd string, filters []string, all bool) error {
	// Get Podman informations:
	output, err := GetPodmanIds(sshCmd, all)
	if err != nil {
		return err
	}
	data, _ := buildPodmanInfo(output, filters)
	// Load config.yaml
	config, err = common.LoadServiceConfigFile(configPath)
	if err != nil {
		return err
	}
	// Update or add data to config
	for name, info := range data {
		if _, ok := config.Services[name]; !ok {
			config.Services[name] = common.Service{}
		}
		if entry, ok := config.Services[name]; ok {
			entry.PodmanId = info["containerid"]
			entry.PodmanImage = info["image"]
			entry.PodmanName = name
			config.Services[name] = entry
		}
	}

	err = dumpConfigFile(configPath)
	if err != nil {
		return err
	}
	return nil
}
