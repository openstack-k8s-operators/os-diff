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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-ini/ini"
	"github.com/go-yaml/yaml"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"

type CompareFileNames struct {
	Origin      string
	Destination string
	DiffReport  []string
}

func writeReport(content []string, reportPath string) error {

	path, _ := filepath.Split(reportPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0700)
	}
	reportContent := strings.Join(content, "")
	err := ioutil.WriteFile(reportPath, []byte(reportContent), 0644)
	log.Info("Write diff file: ", reportPath)
	if err != nil {
		log.Error("Failed to write diff file in: ", reportPath)
		return errors.New("Failed to write report file: '" + reportPath + "'. " + err.Error())
	}
	return nil
}

func (f *CompareFileNames) Compare(origin []byte, destination []byte) error {
	log.Info("Start basic line by line comparison")
	// Split both files into lines
	file1 := strings.Split(string(origin), "\n")
	file2 := strings.Split(string(destination), "\n")

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
				log.Warn("Line: ", line1, " not found in: ", f.Destination, " line: ", i)
				msg = fmt.Sprintf("@@ line: %d\n", i)
				if !stringInSlice(msg, f.DiffReport) {
					f.DiffReport = append(f.DiffReport, msg)
				}
				msg = fmt.Sprintf("+%s\n", line1)
				f.DiffReport = append(f.DiffReport, msg)
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
				log.Warn("Line: ", line2, " not found in: ", f.Origin, " line: ", i)
				msg = fmt.Sprintf("@@ line: %d\n", i)
				if !stringInSlice(msg, f.DiffReport) {
					f.DiffReport = append(f.DiffReport, msg)
					msg = fmt.Sprintf("-%s\n", line2)
					f.DiffReport = append(f.DiffReport, msg)
				} else {
					index = sliceIndex(msg, f.DiffReport)
					//lineNb = strings.Split(f.DiffReport[index], " ")
					msg = fmt.Sprintf("-%s\n", line2)
					f.DiffReport = append(f.DiffReport[:index+2], msg)
				}
				diffFound = true
			}
		}
	}
	if diffFound {
		log.Warn("File: ", f.Origin, " has difference with: ", f.Destination)
		msg := fmt.Sprintf("Source file path: %s, difference with: %s\n", f.Origin, f.Destination)
		f.DiffReport = append([]string{msg}, f.DiffReport...)
	}
	return nil
}

func (f *CompareFileNames) CompareJsonFiles(origin []byte, dest []byte) error {
	// Unmarshal the JSON files into interface{}
	var originData, destData interface{}
	err := json.Unmarshal(origin, &originData)
	if err != nil {
		return fmt.Errorf("Error unmarshalling %s, error: %s", origin, err)
	}
	err = json.Unmarshal(dest, &destData)
	if err != nil {
		return fmt.Errorf("Error unmarshalling %s, error: %s", dest, err)
	}
	compareJSON(originData, destData, "")
	return nil
}

func (f *CompareFileNames) CompareYamlFiles(origin []byte, dest []byte) error {
	// Unmarshal the YAML files into maps
	var map1, map2 map[interface{}]interface{}
	err := yaml.Unmarshal(origin, &map1)
	if err != nil {
		return fmt.Errorf("Error unmarshalling %s, error: %s", origin, err)
	}
	err = yaml.Unmarshal(dest, &map2)
	if err != nil {
		return fmt.Errorf("Error unmarshalling %s, error: %s", dest, err)
	}
	// Compare the maps and print the differences
	msg := string("")
	if !reflect.DeepEqual(map1, map2) {
		//printMapDiff(map1, map2)
		for key, val1 := range map1 {
			val2, isEqual := map2[key]
			if !isEqual {
				msg = fmt.Sprintf("+%v: %v\n", key, val1)
				f.DiffReport = append(f.DiffReport, msg)
			} else if !reflect.DeepEqual(val1, val2) {
				// data1, _ := yaml.Marshal(&val1)
				// data2, _ := yaml.Marshal(&val2)
				// msg = fmt.Sprintf("+%v: %s\n-%v:%s\n", key, data1, key, data2)
				data, _ := yaml.Marshal(&map1)
				ioutil.WriteFile("/tmp/test.yaml", data, 0644)
				msg = fmt.Sprintf("+%v: %s\n-%v:%s\n", key, val1, key, val2)
				f.DiffReport = append(f.DiffReport, msg)
			}
		}
		// Loop through the keys in map2 and check for any missing keys
		for key := range map2 {
			_, isEqual := map1[key]
			if !isEqual {
				msg = fmt.Sprintf("-%v: %v\n", key, map2[key])
				f.DiffReport = append(f.DiffReport, msg)
			}
		}
	}
	return nil
}

func (f *CompareFileNames) CompareIniFiles(origin string, dest string) error {
	// Load the INI files
	cfg1, err := ini.Load(origin)
	if err != nil {
		log.Error("Error while loading file: ", origin, err)
		return fmt.Errorf("Error while loading file %s: %s", origin, err)
	}
	cfg2, err := ini.Load(dest)
	if err != nil {
		log.Error("Error while loading file: ", dest, err)
		return fmt.Errorf("Erro while loading file %s: %s", dest, err)
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
			if !stringInSlice(msg, f.DiffReport) {
				diffFound = true
				log.Warn("Difference detected. Section: ", sec1.Name(), " not found in:", dest)
				f.DiffReport = append(f.DiffReport, msg)
				break
			}
		}
		for _, key1 := range sec1.Keys() {
			key2, err := sec2.GetKey(key1.Name())
			if err != nil {
				if !sectionFound {
					sectionFound = true
					msg = fmt.Sprintf("[%s]\n-%s=%s\n", sec1.Name(), key1.Name(), key1.Value())
				} else {
					msg = fmt.Sprintf("-%s=%s\n", key1.Name(), key1.Value())
				}
				if !stringInSlice(msg, f.DiffReport) {
					diffFound = true
					log.Warn("Difference detected. Section: ", sec1.Name(), " Key ", key1.Name(), " not found in:", dest)
					f.DiffReport = append(f.DiffReport, msg)
				}
			} else {
				if key1.Value() != key2.Value() {
					if !sectionFound {
						sectionFound = true
						msg = fmt.Sprintf(
							"[%s]\n+%s=%s\n-%s=%s\n",
							sec1.Name(),
							key1.Name(),
							key1.Value(),
							key2.Name(),
							key2.Value(),
						)
					} else {
						msg = fmt.Sprintf("+%s=%s\n-%s=%s\n", key1.Name(), key1.Value(), key2.Name(), key2.Value())
					}
					if !stringInSlice(msg, f.DiffReport) {
						diffFound = true
						log.Warn("Difference detected: Values are not equal: ",
							key1.Value(), " and ", key2.Value(),
							"Section: ", sec1.Name(), " Key ", key1.Name(), dest)
						f.DiffReport = append(f.DiffReport, msg)
					}
				}
			}
		}
		// Look for missing keys in Origin:
		for _, key2 := range sec2.Keys() {
			_, err := sec1.GetKey(key2.Name())
			if err != nil {
				if !sectionFound {
					sectionFound = true
					msg = fmt.Sprintf("[%s]\n-%s=%s\n", sec2.Name(), key2.Name(), key2.Value())
				} else {
					msg = fmt.Sprintf("-%s=%s\n", key2.Name(), key2.Value())
				}
				if !stringInSlice(msg, f.DiffReport) {
					diffFound = true
					log.Warn("Difference detected -- Section: ", sec2.Name(), " Key ", key2.Name(), " not found in:", dest)
					f.DiffReport = append(f.DiffReport, msg)
				}
			}
		}
	}
	if diffFound {
		log.Warn("File: ", origin, " has difference with: ", dest)
		msg := fmt.Sprintf("Source file path: %s, difference with: %s\n", origin, dest)
		f.DiffReport = append([]string{msg}, f.DiffReport...)
	}
	return nil
}

func (f *CompareFileNames) CompareFiles() ([]string, error) {
	// Read the files
	log.Info("Start to compare file contents for: ", f.Origin, " and: ", f.Destination)
	orgContent, err := ioutil.ReadFile(f.Origin)
	if err != nil {
		log.Error("Failed to read file", f.Origin, "\n")
		return nil, errors.New("Failed to open file: '" + f.Origin + "'. " + err.Error())
	}
	destContent, err := ioutil.ReadFile(f.Destination)
	if err != nil {
		log.Error("Failed to read file", f.Origin, "\n")
		return nil, errors.New("Failed to open file: '" + f.Destination + "'. " + err.Error())
	}
	// Detect type
	if isIni(orgContent) && isIni(destContent) {
		log.Info("Files detected as Ini files, start to process contents")
		err := f.CompareIniFiles(f.Origin, f.Destination)
		// if error occur, try to make a basic diff
		if err != nil {
			log.Warn(
				"Error while processing files: ",
				f.Origin, " and ",
				f.Destination, " try to compare as a standard type...")
			f.Compare(orgContent, destContent)
		}
	} else if isJson(orgContent) && isJson(destContent) {
		log.Info("Files detected as JSON files, start to process contents")
		f.CompareJsonFiles(orgContent, destContent)
	} else if isYaml(orgContent) && isYaml(destContent) {
		log.Info("Files detected as YAML files, start to process contents")
		f.Compare(orgContent, destContent)
	} else {
		log.Info("No specific type detected, process to a standard line by line comparison...")
		// Check for differences
		f.Compare(orgContent, destContent)
	}
	filePath := f.Origin + ".diff"
	if len(f.DiffReport) != 0 {
		err = writeReport(f.DiffReport, filePath)
		if err != nil {
			log.Error("Error while trying to create diff file in the file system: ", filePath)
			fmt.Println(err)
		}
	}
	return f.DiffReport, nil
}

func (f *CompareFileNames) DiffFiles() error {
	// Drop logging
	log.SetOutput(ioutil.Discard)
	// Read the files
	orgContent, err := ioutil.ReadFile(f.Origin)
	if err != nil {
		log.Error("Failed to read file", f.Origin, "\n")
		return errors.New("Failed to open file: '" + f.Origin + "'. " + err.Error())
	}
	destContent, err := ioutil.ReadFile(f.Destination)
	if err != nil {
		log.Error("Failed to read file", f.Origin, "\n")
		return errors.New("Failed to open file: '" + f.Destination + "'. " + err.Error())
	}
	if isIni(orgContent) && isIni(destContent) {
		f.CompareIniFiles(f.Origin, f.Destination)
	} else {
		f.Compare(orgContent, destContent)
	}

	var output []string
	f.DiffReport = strings.Split(strings.Join(f.DiffReport, ""), "\n")
	for _, line := range f.DiffReport {
		if strings.HasPrefix(line, "+") {
			output = append(output, fmt.Sprintf("%s%s%s\n", Green, line, Reset))
		} else if strings.HasPrefix(line, "-") {
			output = append(output, fmt.Sprintf("%s%s%s\n", Red, line, Reset))
		} else {
			output = append(output, fmt.Sprintf("%s\n", line))
		}
	}
	fmt.Println(strings.Join(output, ""))
	return nil
}
