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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/openstack-k8s-operators/os-diff/pkg/common"
	"github.com/openstack-k8s-operators/os-diff/pkg/godiff"

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

func ExtractCustomServiceConfig(yamlData string) ([]string, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlData), &data); err != nil {
		return nil, err
	}

	var customServiceConfigs []string
	for _, value := range data {
		spec, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		for _, v := range spec {
			template, ok := v.(map[string]interface{})["template"].(map[string]interface{})
			if !ok {
				continue
			}

			customServiceConfig, ok := template["customServiceConfig"].(string)
			if !ok {
				continue
			}
			customServiceConfigs = append(customServiceConfigs, customServiceConfig)
		}
	}
	return customServiceConfigs, nil
}

func DiffServiceConfigWithCRD(service string, crdFile string, configFile string, serviceCfgFile string) error {
	// Load config
	var config common.Config
	config, _ = common.LoadServiceConfigFile(serviceCfgFile)
	//Load files
	src, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(err)
		return err
	}
	yamlFile, err := ioutil.ReadFile(crdFile)
	if err != nil {
		return err
	}
	// Make sure crdFile is Yaml
	if common.DetectType([]byte(crdFile)) != "yaml" {
		fmt.Println("Error, file2 is not a Yaml or a CRD file. Please provide a correct file.")
		return fmt.Errorf("wrong file2 type")
	}
	if service != "" {
		if config.Services[service].ConfigMapping != nil {
			var fileMap map[string]string
			if service == "ovs_external_ids" {
				fileMap = LoadOvsExternalIds(configFile)
			} else {
				if common.DetectType(src) == "raw" {
					fileMap, _ = LoadFilesIntoMap(configFile)
				} else {
					fmt.Println("File type not supported, only support format as: key=value or key: value.")
					return nil
				}
			}
			var edpmService OpenStackDataPlaneNodeSet
			err = yaml.Unmarshal(yamlFile, &edpmService)
			if err != nil {
				panic(err)
			}
			fmt.Println("Start to compare file contents for: " + configFile + " and " + crdFile)
			return CompareMappingConfig(fileMap, config.Services[service].ConfigMapping, edpmService)
		}
	}

	if common.DetectType(src) == "ini" {
		customServiceConfigs, err := ExtractCustomServiceConfig(string(yamlFile))
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		_, err = CompareIniConfig(src, []byte(strings.Join(customServiceConfigs, "")), configFile, crdFile)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Unsupported config file type, only support INI file.")
	}
	return nil
}

func DiffServiceConfigFromPod(service string, crdFile string, configFile string, serviceCfgFile string) error {
	var config common.Config
	config, _ = common.LoadServiceConfigFile(serviceCfgFile)
	yamlFile, err := ioutil.ReadFile(crdFile)
	if err != nil {
		return err
	}
	customServiceConfigs, err := ExtractCustomServiceConfig(string(yamlFile))
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	// Get service Config
	podConfig, err := GetConfigFromPod(configFile, config.Services[service].PodName, config.Services[service].ContainerName)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig(podConfig, []byte(strings.Join(customServiceConfigs, "")), configFile, crdFile)
	if err != nil {
		panic(err)
	}
	return nil
}

func DiffServiceConfigFromPodman(service string, crdFile string, configFile string, serviceCfgFile string) error {
	var config common.Config
	config, _ = common.LoadServiceConfigFile(serviceCfgFile)
	// Get ocpConfig
	yamlFile, err := ioutil.ReadFile(crdFile)
	if err != nil {
		return err
	}
	customServiceConfigs, err := ExtractCustomServiceConfig(string(yamlFile))
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Get service Config
	osConfig, err := GetConfigFromPodman(configFile, config.Services[service].PodmanName)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig(osConfig, []byte(strings.Join(customServiceConfigs, "")), configFile, crdFile)
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
