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
	"log"
	"os"
	"os-diff/pkg/godiff"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ParentStruct struct {
	Spec SimpleServiceStruct `yaml:"spec"`
}

type SimpleServiceStruct map[string]struct {
	Enabled  bool `yaml:"enabled"`
	Template struct {
		CustomServiceConfig string `yaml:"customServiceConfig"`
	} `yaml:"template"`
}

type KeystoneConfigMapStruct struct {
	Data struct {
		CustomConf            string `yaml:"custom.conf"`
		HttpdConf             string `yaml:"httpd.conf"`
		KeystoneAPIConfigJSON string `yaml:"keystone-api-config.json"`
		KeystoneConf          string `yaml:"keystone.conf"`
	} `yaml:"data"`
}

type ConfigMapDataStruct struct {
	Data map[string]string `yaml:"data"`
}

type ConfigMapConf string

func DiffServiceConfig(service string, ocpConfig string, serviceConfig string, sidebyside bool) error {
	var servicePatch string
	// Get ocpConfig
	if service == "cinder" {
		servicePatch = LoadCinderOpenShiftConfig(ocpConfig)
	} else if service == "glance" {
		servicePatch = LoadGlanceOpenShiftConfig(ocpConfig)
	} else {
		msg := `Service not supported, please implement it.
			Follow the instructions to add new OpenStack services here:
			https://github.com/openstack-k8s-operators/os-diff#add-service`
		panic(msg)
	}

	// Get service Config
	osConfig, err := LoadServiceConfig(serviceConfig)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig(osConfig, []byte(servicePatch), serviceConfig, ocpConfig)
	if err != nil {
		panic(err)
	}
	if sidebyside {
		_, err = CompareIniConfig(osConfig, []byte(servicePatch), serviceConfig, ocpConfig)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func DiffServiceConfigFromPod(service string, ocpConfig string, serviceConfig string, containerName string) error {
	var servicePatch string
	var podName string
	// Get ocpConfig
	if service == "cinder" {
		podName = "cinder"
		servicePatch = LoadCinderOpenShiftConfig(ocpConfig)
	} else if service == "glance" {
		// @todo: should be move a config spec file, users must be describe their env in a file.cfg.
		podName = "glance-external-api"
		servicePatch = LoadGlanceOpenShiftConfig(ocpConfig)
	} else {
		msg := `Service not supported, please implement it.
			Follow the instructions to add new OpenStack services here:
			https://github.com/openstack-k8s-operators/os-diff#add-service`
		panic(msg)
	}
	// Get service Config
	podConfig, err := GetConfigFromPod(serviceConfig, podName, containerName)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig(podConfig, []byte(servicePatch), serviceConfig, ocpConfig)
	if err != nil {
		panic(err)
	}
	return nil
}

func DiffServiceConfigFromPodman(service string, ocpConfig string, serviceConfig string, podname string) error {
	var servicePatch string
	// Get ocpConfig
	if service == "cinder" {
		servicePatch = LoadCinderOpenShiftConfig(ocpConfig)
	} else if service == "glance" {
		servicePatch = LoadGlanceOpenShiftConfig(ocpConfig)
	} else {
		msg := `Service not supported, please implement it.
			Follow the instructions to add new OpenStack services here:
			https://github.com/openstack-k8s-operators/os-diff#add-service`
		panic(msg)
	}
	// Get service Config
	osConfig, err := GetConfigFromPodman(serviceConfig, podname)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig(osConfig, []byte(servicePatch), serviceConfig, ocpConfig)
	if err != nil {
		panic(err)
	}
	return nil
}

func GenerateConfigPatchFromIni(serviceName string, configFile string, outputFile string, serviceEnable bool) error {
	config, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	return GenerateConfigPatch(serviceName, config, outputFile, serviceEnable)
}

func GenerateConfigPatchFromRemote(serviceName string, configFile string, outputFile string, serviceEnable bool, podname string) error {
	// Get service Config
	osConfig, err := GetConfigFromPodman(configFile, podname)
	if err != nil {
		panic(err)
	}
	return GenerateConfigPatch(serviceName, osConfig, outputFile, serviceEnable)
}

func GenerateConfigPatch(serviceName string, config []byte, outputFile string, serviceEnable bool) error {
	configStr := strings.Split(string(config), "\n")
	var configClean []string
	for _, line := range configStr {
		if !strings.HasPrefix(line, "#") && len(line) > 0 && (strings.Contains(line, "=") || strings.HasPrefix(line, "[")) {
			configClean = append(configClean, line)
		}
	}
	// Service structure
	parentStruct := ParentStruct{}
	configStruct := SimpleServiceStruct{}

	service := configStruct[serviceName]

	service.Enabled = serviceEnable
	service.Template.CustomServiceConfig = string(strings.Join(configClean[:], "\n"))
	configStruct[serviceName] = service

	parentStruct.Spec = configStruct

	yamlData, err := yaml.Marshal(&parentStruct)
	if err != nil {
		fmt.Printf("Error marshaling YAML: %v\n", err)
		return err
	}
	err = os.WriteFile(outputFile, yamlData, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return nil
	}

	fmt.Println("YAML file generated: ", outputFile)
	return nil
}

func DiffConfigMap(configMapName string, orgConfigPath string, fromRemote bool, remoteCmd string) error {
	var config []byte
	var err error
	var isDir bool
	var isConfigNameisDir bool
	// Get configMap
	configMapStat, err := os.Stat(configMapName)
	if err != nil {
		config, err = GetOCConfigMap(configMapName)
		if err != nil {
			return err
		}
	} else if !configMapStat.IsDir() {
		config, err = os.ReadFile(configMapName)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Wrong configmap arguments, need file or oc get configmap/<name> instead.")
	}

	isDir = false
	if fromRemote {
		isDir, err = RemoteStatDir(remoteCmd, orgConfigPath)
		if err != nil {
			fmt.Println("Error while trying to stat remote:", orgConfigPath, "no such file or directory.")
			return err
		}
	} else {
		configPathStat, err := os.Stat(orgConfigPath)
		if err != nil {
			fmt.Println(err)
			return err
		}
		isDir = configPathStat.IsDir()
	}
	// Start processing data
	var configMapdata ConfigMapDataStruct
	err = yaml.Unmarshal(config, &configMapdata)
	if err != nil {
		log.Fatal(err)
	}
	for key, _ := range configMapdata.Data {
		if isDir {
			// Check if orgConfigPath and confName exists
			confPath := filepath.Join(orgConfigPath, key)
			if fromRemote {
				isConfigNameisDir, err = RemoteStatDir(remoteCmd, confPath)
				if err != nil {
					continue
				}
			} else {
				configNameStat, err := os.Stat(confPath)
				if err != nil {
					continue
				}
				isConfigNameisDir = configNameStat.IsDir()
			}
			if !isConfigNameisDir {
				configMapPath := filepath.Join(configMapName, key)
				compareIniFromFileAndStringBuilder(configMapdata.Data[key], confPath, configMapPath, fromRemote, remoteCmd)
			} else {
				continue
			}
		} else {
			fileName := filepath.Base(orgConfigPath)
			if fileName == key {
				configMapPath := filepath.Join(configMapName, key)
				compareIniFromFileAndStringBuilder(configMapdata.Data[key], orgConfigPath, configMapPath, fromRemote, remoteCmd)
			}
		}
	}
	return nil
}

func compareIniFromFileAndStringBuilder(configString string, configFile string, path1 string, remote bool, remoteCmd string) error {
	var sb strings.Builder
	var configContent []byte
	var err error
	sb.WriteString(configString)
	if !remote {
		configContent, err = os.ReadFile(configFile)
		if err != nil {
			return err
		}
	} else {
		configContent, err = godiff.GetConfigFromRemote(remoteCmd, configFile)
		if err != nil {
			return err
		}
	}
	_, err = CompareIniConfig([]byte(sb.String()), configContent, path1, configFile)
	if err != nil {
		panic(err)
	}
	return nil
}
