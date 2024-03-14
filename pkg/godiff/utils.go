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
package godiff

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"strings"

	"github.com/go-ini/ini"
	"gopkg.in/yaml.v3"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func sliceIndex(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

func isIni(data []byte) bool {
	if data[0] == '[' {
		return true
	}
	return false
}

func isYaml(data []byte) bool {
	var yamlData interface{}
	err := yaml.Unmarshal(data, &yamlData)
	if err == nil {
		return true
	}
	return false
}

func isJson(data []byte) bool {
	var jsonData interface{}
	err := json.Unmarshal(data, &jsonData)
	if err == nil {
		// fmt.Errorf("Faild to unmarshal json file %s", err)
		return true
	}
	return false
}

func CompareYAML(origin []byte, dest []byte) ([]string, error) {

	var report []string
	var map1, map2 map[interface{}]interface{}
	err := yaml.Unmarshal(origin, &map1)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling %s, error: %s", origin, err)
	}
	err = yaml.Unmarshal(dest, &map2)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling %s, error: %s", dest, err)
	}
	// Compare the maps and print the differences
	msg := string("")
	if !reflect.DeepEqual(map1, map2) {
		//printMapDiff(map1, map2)
		for key, val1 := range map1 {
			val2, isEqual := map2[key]
			if !isEqual {
				msg = fmt.Sprintf("+%v: %v\n", key, val1)
				report = append(report, msg)
			} else if !reflect.DeepEqual(val1, val2) {
				// data1, _ := yaml.Marshal(&val1)
				// data2, _ := yaml.Marshal(&val2)
				// msg = fmt.Sprintf("+%v: %s\n-%v:%s\n", key, data1, key, data2)
				data, _ := yaml.Marshal(&map1)
				ioutil.WriteFile("/tmp/test.yaml", data, 0644)
				msg = fmt.Sprintf("+%v: %s\n-%v:%s\n", key, val1, key, val2)
				report = append(report, msg)
			}
		}
		// Loop through the keys in map2 and check for any missing keys
		for key := range map2 {
			_, isEqual := map1[key]
			if !isEqual {
				msg = fmt.Sprintf("-%v: %v\n", key, map2[key])
				report = append(report, msg)
			}
		}
	}
	return report, nil
}

func CompareJSON(orgData, destData interface{}, path string) ([]string, error) {
	if reflect.TypeOf(orgData) != reflect.TypeOf(destData) {
		//fmt.Println("Type mismatch at %s: %T != %T\n", path, orgData, destData)
		return nil, fmt.Errorf("Type mismatch at %s: %T != %T\n", path, orgData, destData)
	}

	var diff []string
	msg := string("")
	switch orgData := orgData.(type) {
	case map[string]interface{}:
		destData := destData.(map[string]interface{})
		for key, value := range orgData {
			if value2, ok := destData[key]; ok {
				CompareJSON(value, value2, path+"."+key)
			} else {
				//fmt.Println("Key %s not found in second JSON\n", path+"."+key)
				msg = fmt.Sprintf("+%s", key)
				diff = append(diff, msg)
			}
		}
		for key := range destData {
			if _, ok := orgData[key]; !ok {
				//fmt.Println("Key %s not found in first JSON\n", path+"."+key)
				msg = fmt.Sprintf("-%s", key)
				diff = append(diff, msg)
			}
		}
	case []interface{}:
		destData := destData.([]interface{})
		if len(orgData) != len(destData) {
			//fmt.Println("Array length mismatch at %s: %d != %d\n", path, len(orgData), len(destData))
			return diff, fmt.Errorf("Array length mismatch at %s: %d != %d\n", path, len(orgData), len(destData))
		}
		for i := range orgData {
			CompareJSON(orgData[i], destData[i], fmt.Sprintf("%s[%d]", path, i))
		}
	default:
		if !reflect.DeepEqual(orgData, destData) {
			//fmt.Println("Value mismatch at %s: %v != %v\n", path, orgData, destData)
			return diff, nil
		}
	}
	return diff, nil
}

func CompareIni(rawdata1 []byte, rawdata2 []byte, origin string, dest string, verbose bool) ([]string, error) {
	if !verbose {
		log.SetOutput(ioutil.Discard)
	}
	var report []string
	// Load the INI files
	cfg1, err := ini.Load(rawdata1)
	if err != nil {
		return nil, fmt.Errorf("Error while loading file %s: %s", origin, err)
	}
	cfg2, err := ini.Load(rawdata2)
	if err != nil {
		log.Error("Error while loading file: ", dest, err)
		return nil, fmt.Errorf("Erro while loading file %s: %s", dest, err)
	}

	diffFound := false
	sectionFound := false
	msg := string("")
	// Compare the sections and keys in each file
	for _, sec1 := range cfg1.Sections() {
		sectionFound = false
		sec2, err := cfg2.GetSection(sec1.Name())
		if err != nil {
			msg := fmt.Sprintf("-[%s]\n", sec1.Name())
			if !stringInSlice(msg, report) {
				diffFound = true
				log.Warn("Difference detected. Section: ", sec1.Name(), " not found in:", dest)
				report = append(report, msg)
			}
		}
		for _, key1 := range sec1.Keys() {
			if sec2 == nil {
				if !sectionFound {
					sectionFound = true
					msg = fmt.Sprintf("[%s]\n-%s=%s\n", sec1.Name(), key1.Name(), key1.Value())
				} else {
					msg = fmt.Sprintf("-%s=%s\n", key1.Name(), key1.Value())
				}
				if !stringInSlice(msg, report) {
					diffFound = true
					log.Warn("Difference detected. Section: ", sec1.Name(), " Key ", key1.Name(), " not found in:", dest)
					report = append(report, msg)
				}
			} else {
				key2, err := sec2.GetKey(key1.Name())
				if err != nil {
					// key2 not found
					if !sectionFound {
						sectionFound = true
						msg = fmt.Sprintf("[%s]\n-%s=%s\n", sec1.Name(), key1.Name(), key1.Value())
					} else {
						msg = fmt.Sprintf("-%s=%s\n", key1.Name(), key1.Value())
					}
					if !stringInSlice(msg, report) {
						diffFound = true
						log.Warn("Difference detected. Section: ", sec1.Name(), " Key ", key1.Name(), " not found in:", dest)
						report = append(report, msg)
					}
				} else if key1.Value() != key2.Value() {
					if !sectionFound {
						sectionFound = true
						msg = fmt.Sprintf(
							"[%s]\n-%s=%s\n+%s=%s\n",
							sec1.Name(),
							key1.Name(),
							key1.Value(),
							key2.Name(),
							key2.Value(),
						)
					} else {
						msg = fmt.Sprintf("-%s=%s\n+%s=%s\n", key1.Name(), key1.Value(), key2.Name(), key2.Value())
					}
					if !stringInSlice(msg, report) {
						diffFound = true
						log.Warn("Difference detected: Values are not equal: ",
							key1.Value(), " and ", key2.Value(),
							"Section: ", sec1.Name(), " Key ", key1.Name(), dest)
						report = append(report, msg)
					}
				}
			}
		}
		// Look for missing keys in Origin:
		if sec2 != nil {
			for _, key2 := range sec2.Keys() {
				_, err := sec1.GetKey(key2.Name())
				if err != nil {
					if !sectionFound {
						sectionFound = true
						msg = fmt.Sprintf("[%s]\n+%s=%s\n", sec2.Name(), key2.Name(), key2.Value())
					} else {
						msg = fmt.Sprintf("+%s=%s\n", key2.Name(), key2.Value())
					}
					if !stringInSlice(msg, report) {
						diffFound = true
						log.Warn("Difference detected -- Section: ", sec2.Name(), " Key ", key2.Name(), " not found in:", dest)
						report = append(report, msg)
					}
				}
			}
		}
	}
	for _, sec2 := range cfg2.Sections() {
		sectionFound = false
		_, err := cfg1.GetSection(sec2.Name())
		if err != nil {
			msg := fmt.Sprintf("-[%s]\n", sec2.Name())
			if !stringInSlice(msg, report) {
				diffFound = true
				log.Warn("Difference detected. Section: ", sec2.Name(), " not found in:", dest)
				report = append(report, msg)
			}
			for _, key2 := range sec2.Keys() {
				msg = fmt.Sprintf("-%s=%s\n", key2.Name(), key2.Value())
				if !stringInSlice(msg, report) {
					log.Warn("Difference detected -- Section: ", sec2.Name(), " Key ", key2.Name(), " not found in:", dest)
					report = append(report, msg)
				}
			}
		}
	}
	if diffFound {
		log.Warn("File: ", origin, " has difference with: ", dest)
		msg := fmt.Sprintf("Source file path: %s, difference with: %s\n", origin, dest)
		report = append([]string{msg}, report...)
	}
	return report, nil
}

func CompareRawData(rawdata1 []byte, rawdata2 []byte, origin string, dest string) ([]string, error) {
	var report []string
	log.Info("Start basic line by line comparison")
	// Split both files into lines
	file1 := strings.Split(string(rawdata1), "\n")
	file2 := strings.Split(string(rawdata2), "\n")

	found := false
	diffFound := false
	msg := string("")
	for i, line1 := range file1 {
		found = false
		// Skip comments
		if !strings.HasPrefix(line1, "#") && len(line1) > 0 {
			for _, line2 := range file2 {
				if !strings.HasPrefix(line2, "#") && len(line2) > 0 {
					if line1 == line2 {
						found = true
						break
					}
				}
			}
			if !found {
				log.Warn("Line: ", line1, " not found in: ", dest, " line: ", i)
				msg = fmt.Sprintf("@@ line: %d\n", i)
				if !stringInSlice(msg, report) {
					report = append(report, msg)
				}
				msg = fmt.Sprintf("+%s\n", line1)
				report = append(report, msg)
				diffFound = true
			}
		}
	}
	//var lineNb []string
	var index int
	for i, line2 := range file2 {
		found = false
		// Skip comments
		if !strings.HasPrefix(line2, "#") && len(line2) > 0 {
			for _, line1 := range file1 {
				if !strings.HasPrefix(line1, "#") && len(line1) > 0 {
					if line1 == line2 {
						found = true
						break
					}
				}
			}
			if !found {
				log.Warn("Line: ", line2, " not found in: ", origin, " line: ", i)
				msg = fmt.Sprintf("@@ line: %d\n", i)
				if !stringInSlice(msg, report) {
					report = append(report, msg)
					msg = fmt.Sprintf("-%s\n", line2)
					report = append(report, msg)
				} else {
					index = sliceIndex(msg, report)
					msg = fmt.Sprintf("-%s\n", line2)
					report = append(report[:index+2], msg)
				}
				diffFound = true
			}
		}
	}
	if diffFound {
		log.Warn("File: ", origin, " has difference with: ", dest)
		msg := fmt.Sprintf("Source file path: %s, difference with: %s\n", origin, dest)
		report = append([]string{msg}, report...)
	}
	return report, nil
}

func GetConfigFromRemote(remoteCmd string, configPath string) ([]byte, error) {
	// Build command:
	cmd := remoteCmd + " cat " + configPath
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println(string(out))
		return out, err
	}
	return []byte(out), nil
}

func CleanIniSections(config string) string {
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
