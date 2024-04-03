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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type AnsibleHostStruct struct {
	AnsibleHost              string `yaml:"ansible_ssh_host,omitempty"`
	AnsibleUser              string `yaml:"ansible_user,omitempty"`
	AnsiblePort              string `yaml:"ansible_port,omitempty"`
	AnsibleSSHPrivateKeyFile string `yaml:"ansible_ssh_private_key_file,omitempty"`
}

type Group struct {
	Hosts map[string]AnsibleHostStruct `yaml:"hosts"`
	Vars  map[string]interface{}       `yaml:"vars"`
}

// type Inventory struct {
// 	All struct {
// 		Children  map[string]Group             `yaml:"children"`
// 		Ungrouped map[string]AnsibleHostStruct `yaml:"ungrouped"`
// 	} `yaml:"all"`
// }

type Inventory map[string]Group

// Host structure for ssh config file
type Host struct {
	Name                  string
	HostName              string
	IdentityFile          string
	Port                  string
	User                  string
	AdditionalLines       []string
	StrictHostKeyChecking string
	UserKnownHostsFile    string
}

func BuildSshConfigFile(inventoryFile string, sshConfigFile string, yaml bool, etc bool) error {
	if yaml {
		return BuildSshConfigFileFromYaml(inventoryFile, sshConfigFile)
	} else if etc {
		return BuildSshConfigFileFromEtcHosts(inventoryFile, sshConfigFile)
	}
	return BuildSshConfigFileFromIni(inventoryFile, sshConfigFile)
}

func BuildSshConfigFileFromEtcHosts(etcHostsFile string, sshConfigFile string) error {
	// Open the /etc/hosts file
	data, err := ioutil.ReadFile(etcHostsFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}

	// Create SSH config file
	sshCfgFile, err := os.Create(sshConfigFile)
	if err != nil {
		fmt.Println("Error creating ssh config file:", err)
		return err
	}
	defer sshCfgFile.Close()

	var sshConfig *Host
	// Split the file content into lines
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}
		// Split the line into fields
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if len(fields) == 3 {
			sshConfig = &Host{Name: fields[2]}
		} else {
			sshConfig = &Host{Name: fields[1]}
		}
		sshConfig.HostName = fields[1]
		writeHostConfig(sshCfgFile, sshConfig)
	}

	return nil
}

func BuildSshConfigFileFromYaml(inventoryFile string, sshConfigFile string) error {
	// Open the inventory file
	data, err := ioutil.ReadFile(inventoryFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}

	// Parse the YAML file
	var inventory Inventory
	err = yaml.Unmarshal(data, &inventory)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		return err
	}

	// Create SSH config file
	sshCfgFile, err := os.Create(sshConfigFile)
	if err != nil {
		fmt.Println("Error creating ssh config file:", err)
		return err
	}
	var sshConfig *Host
	// Iterate over the groups and hosts
	for _, group := range inventory {
		for hostName, host := range group.Hosts {
			// Write the host configuration to the SSH config file
			sshConfig = &Host{Name: hostName}
			if host.AnsibleHost != "" {
				sshConfig.HostName = host.AnsibleHost
			} else {
				sshConfig.HostName = hostName
			}
			writeHostConfig(sshCfgFile, sshConfig)
		}
	}
	return nil
}

func BuildSshConfigFileFromIni(inventoryFile string, sshConfigFile string) error {
	file, err := os.Open(inventoryFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return err
	}
	sshCfgFile, err := os.Create(sshConfigFile)
	if err != nil {
		fmt.Println("Error creating ssh config file:", err)
		return err
	}
	// Read the inventory file line by line
	scanner := bufio.NewScanner(file)
	var sshConfig *Host
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			continue
		}

		// Split host and parameters
		parts := strings.Split(line, " ")
		hostName := parts[0]
		parameters := parts[1:]

		if sshConfig != nil {
			writeHostConfig(sshCfgFile, sshConfig)
		}
		sshConfig = &Host{Name: hostName}
		// Set host name and static parameters
		sshConfig.HostName = hostName
		for _, param := range parameters {
			if strings.HasPrefix(param, "ansible_ssh_private_key_file=") {
				sshConfig.IdentityFile = strings.Split(param, "=")[1]
			} else if strings.HasPrefix(param, "ansible_port=") {
				sshConfig.Port = strings.Split(param, "=")[1]
			} else if strings.HasPrefix(param, "ansible_user=") {
				sshConfig.User = strings.Split(param, "=")[1]
			} else {
				sshConfig.AdditionalLines = append(sshConfig.AdditionalLines, param)
			}
		}

	}
	// Write configuration for the last host
	if sshConfig != nil {
		writeHostConfig(sshCfgFile, sshConfig)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}
	return nil
}

func writeHostConfig(file *os.File, host *Host) {
	file.WriteString(fmt.Sprintf("Host %s\n", host.Name))
	file.WriteString(fmt.Sprintf("  HostName %s\n", host.HostName))
	if host.IdentityFile != "" {
		file.WriteString(fmt.Sprintf("  IdentityFile %s\n", host.IdentityFile))
	}
	if host.Port != "" {
		file.WriteString(fmt.Sprintf("  Port %s\n", host.Port))
	}
	if host.User != "" {
		file.WriteString(fmt.Sprintf("  User %s\n", host.User))
	} else {
		file.WriteString("  User root\n")
	}
	if host.StrictHostKeyChecking == "" {
		file.WriteString("  StrictHostKeyChecking no\n")
	} else {
		file.WriteString(fmt.Sprintf("  StrictHostKeyChecking %s\n", host.StrictHostKeyChecking))
	}
	if host.UserKnownHostsFile == "" {
		file.WriteString("  UserKnownHostsFile /dev/null\n")
	} else {
		file.WriteString(fmt.Sprintf("  StrictHostKeyChecking %s\n", host.UserKnownHostsFile))
	}
	for _, line := range host.AdditionalLines {
		file.WriteString(fmt.Sprintf("  %s\n", line))
	}
	file.WriteString("\n")
}
