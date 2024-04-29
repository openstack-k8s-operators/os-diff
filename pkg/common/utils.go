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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

var config Config

// Service YAML Config Structure
type Service struct {
	Enable             bool              `yaml:"enable"`
	PodmanId           string            `yaml:"podman_id"`
	PodmanImage        string            `yaml:"podman_image"`
	PodmanName         string            `yaml:"podman_name"`
	PodName            string            `yaml:"pod_name"`
	ContainerName      string            `yaml:"container_name"`
	StrictPodNameMatch bool              `yaml:"strict_pod_name_match"`
	Path               []string          `yaml:"path"`
	Hosts              []string          `yaml:"hosts"`
	ServiceCommand     string            `yaml:"service_command"`
	CatOutput          bool              `yaml:"cat_output"`
	ConfigMapping      map[string]string `yaml:"config_mapping"`
}

type Config struct {
	Services map[string]Service `yaml:"services"`
}

// Shell execution functions:
func ExecCmd(cmd string) ([]string, error) {
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return strings.Split(string(output), "\n"), err
	}
	return strings.Split(string(output), "\n"), nil
}

func ExecCmdSimple(cmd string) (string, error) {
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

func ExecComplexCmd(cmd string) (string, error) {
	// Format Shel command before execute
	args := FormatShellCommand(cmd)
	output, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		fmt.Println(err)
		return string(output), err
	}
	return string(output), nil
}

func TestOCConnection() bool {
	cmd := "oc whoami"
	_, err := ExecCmd(cmd)
	if err != nil {
		return false
	}
	return true
}

func TestSshConnection(sshCmd string) bool {
	cmd := sshCmd + " ls"
	_, err := ExecCmd(cmd)
	if err != nil {
		return false
	}
	return true
}

func TestEqualSlice(a []string, b []string) bool {
	if len(a) != len(b) {
		fmt.Println(len(a), len(b))
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			fmt.Println(a[i], b[i])
			return false
		}
	}
	return true
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ToLowerSlice(data []string) []string {
	var lowerData []string
	for _, d := range data {
		lowerData = append(lowerData, strings.ToLower(d))
	}
	return lowerData
}

func SliceIndex(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

func GetNestedFieldValue(data interface{}, keyName string) interface{} {
	val := reflect.ValueOf(data)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	field := val.FieldByName(keyName)
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

func LoadServiceConfigFile(configPath string) (Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return config, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding YAML:", err)
		return config, err
	}
	return config, nil
}

func ConvertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case []string:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	var result string
	for _, part := range parts {
		result += strings.Title(part)
	}
	return result
}

func IsIni(data []byte) bool {
	if data[0] == '[' {
		return true
	}
	return false
}

func IsYaml(data []byte) bool {
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		fmt.Println("Error unmarshaling YAML:", err)
		return false
	}
	return true
}

func IsJson(data []byte) bool {
	var jsonData interface{}
	err := json.Unmarshal(data, &jsonData)
	return err == nil
}

func DetectType(value []byte) string {
	switch {
	case IsIni(value):
		return "ini"
	case IsYaml(value):
		return "yaml"
	case IsJson(value):
		return "json"
	default:
		return "raw"
	}
}

func FormatShellCommand(input string) []string {
	var tokens []string
	var currentToken string
	inQuote := false
	quoteChar := rune(0)

	for _, char := range input {
		switch char {
		case '"':
			if !inQuote || quoteChar == '"' {
				inQuote = !inQuote
				quoteChar = '"'
			}
			currentToken += string(char)
		case '\'':
			if !inQuote || quoteChar == '\'' {
				inQuote = !inQuote
				quoteChar = '\''
			}
			currentToken += string(char)
		case ' ', '\t':
			if !inQuote {
				if currentToken != "" {
					tokens = append(tokens, currentToken)
					currentToken = ""
				}
			} else {
				currentToken += string(char)
			}
		default:
			currentToken += string(char)
		}
	}
	// Add the last token
	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}
	return tokens
}

func BuildFullSshCmd(sshCmd string, host string) (string, string, error) {
	sshCmd = strings.Join(strings.Fields(sshCmd), " ")
	atIndex := strings.LastIndex(sshCmd, "@")
	if atIndex == -1 {
		fIndex := strings.Index(sshCmd, "-F")
		if fIndex != -1 {
			fParts := strings.SplitN(sshCmd[fIndex:], " ", 4)
			if len(fParts) < 3 {
				return sshCmd + " " + host, host, nil
			} else if len(fParts) == 3 && (host == "" || fParts[2] == host) {
				fmt.Printf("director_host option is already set in ssh_cmd or empty, using ssh_cmd as full command: %s and: %s as director host...", sshCmd, fParts[2])
				return sshCmd, fParts[2], nil
			} else if len(fParts) == 3 && fParts[2] != host {
				fmt.Printf("Error: The host in the ssh_cmd: %s does not match the director_host: %s\n", fParts[2], host)
				return "", "", fmt.Errorf("error: The host in the ssh_cmd: %s does not match the director_host: %s", fParts[2], host)
			}
			if len(fParts) > 3 {
				fmt.Println("Error: Too many arguments after -F option")
				return "", "", errors.New("error: Too many arguments after -F option")
			}
		}
		return sshCmd + " " + host, host, nil
	} else if atIndex == len(sshCmd)-1 {
		return sshCmd + host, host, nil
	} else {
		cmdHost := sshCmd[atIndex+1:]
		if host == "" {
			fmt.Printf("director_host option is already set in ssh_cmd or empty, using ssh_cmd as full command: %s and: %s as director host...", sshCmd, cmdHost)
			return sshCmd, cmdHost, nil
		}
		if cmdHost != host {
			fmt.Printf("Error: The host in the sshCmd %s does not match the DirectorHost %s\n", cmdHost, host)
			return "", "", fmt.Errorf("error: The host in the sshCmd %s does not match the directorHost %s", cmdHost, host)
		}
		return sshCmd, cmdHost, nil
	}
}
