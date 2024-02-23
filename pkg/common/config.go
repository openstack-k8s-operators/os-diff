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
	"fmt"

	"github.com/go-ini/ini"
)

var oDConfig ODConfig

// OS Diff Config Structure
type ODConfig struct {
	Default struct {
		LocalConfigDir    string `ini:"local_config_dir"`
		ServiceConfigFile string `ini:"service_config_file"`
	} `ini:"Default"`

	Tripleo struct {
		SshCmd           string `ini:"ssh_cmd"`
		DirectorHost     string `ini:"director_host"`
		ContainerEngine  string `ini:"container_engine"`
		Connection       string `ini:"connection"`
		RemoteConfigPath string `ini:"remote_config_path"`
		LocalConfigPath  string `ini:"local_config_path"`
	} `ini:"Tripleo"`

	Openshift struct {
		OcpLocalConfigPath string `ini:"ocp_local_config_path"`
		Connection         string `ini:"connection"`
	} `ini:"Openshift"`
}

func LoadOSDiffConfig(configFileName string) (*ODConfig, error) {
	cfg, err := ini.Load(configFileName)
	if err != nil {
		fmt.Println("Error loading config file:", err)
		return nil, err
	}

	err = cfg.MapTo(&oDConfig)
	if err != nil {
		fmt.Println("Error mapping config file:", err)
		return nil, err
	}
	return &oDConfig, nil
}
