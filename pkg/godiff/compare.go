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
	"strings"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"

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

func PrintReport(report []string) error {
	var output []string
	report = strings.Split(strings.Join(report, ""), "\n")
	for _, line := range report {
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

func CompareJSONFiles(origin []byte, dest []byte) ([]string, error) {
	// Unmarshal the JSON files into interface{}
	var originData, destData interface{}
	err := json.Unmarshal(origin, &originData)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling %s, error: %s", origin, err)
	}
	err = json.Unmarshal(dest, &destData)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling %s, error: %s", dest, err)
	}
	report, err := CompareJSON(originData, destData, "")
	if err != nil {
		return nil, err
	}
	return report, nil
}

func CompareFiles(origin string, dest string, print bool, verbose bool) ([]string, error) {
	var report []string
	if print {
		log.SetOutput(ioutil.Discard)
	}
	// Read the files
	log.Info("Start to compare file contents for: ", origin, " and: ", dest)
	orgContent, err := ioutil.ReadFile(origin)
	if err != nil {
		log.Error("Failed to read file", origin, "\n")
		return nil, errors.New("Failed to open file: '" + origin + "'. " + err.Error())
	}
	destContent, err := ioutil.ReadFile(dest)
	if err != nil {
		log.Error("Failed to read file", origin, "\n")
		return nil, errors.New("Failed to open file: '" + dest + "'. " + err.Error())
	}
	// Detect type
	if isIni(orgContent) && isIni(destContent) {
		log.Info("Files detected as Ini files, start to process contents")
		report, err = CompareIni(orgContent, destContent, origin, dest, verbose)
		// if error occur, try to make a basic diff
		if err != nil {
			log.Warn(
				"Error while processing files: ",
				origin, " and ",
				dest, " try to compare as a standard type...")
			report, _ = CompareRawData(orgContent, destContent, origin, dest)
		}
	} else if isJson(orgContent) && isJson(destContent) {
		log.Info("Files detected as JSON files, start to process contents")
		report, err = CompareJSONFiles(orgContent, destContent)
		if err != nil {
			log.Warn(
				"Error while processing files: ",
				origin, " and ",
				dest, " try to compare as a standard type...")
			report, _ = CompareRawData(orgContent, destContent, origin, dest)
		}
	} else if isYaml(orgContent) && isYaml(destContent) {
		log.Info("Files detected as YAML files, start to process contents")
		report, err = CompareYAML(orgContent, destContent)
		if err != nil {
			log.Warn(
				"Error while processing files: ",
				origin, " and ",
				dest, " try to compare as a standard type...")
			report, _ = CompareRawData(orgContent, destContent, origin, dest)
		}
	} else {
		log.Info("No specific type detected, process to a standard line by line comparison...")
		// Check for differences
		report, _ = CompareRawData(orgContent, destContent, origin, dest)
	}
	filePath := origin + ".diff"
	if len(report) != 0 {
		err = writeReport(report, filePath)
		if err != nil {
			log.Error("Error while trying to create diff file in the file system: ", filePath)
			fmt.Println(err)
		}
		if print {
			PrintReport(report)
		}
	}
	return report, nil
}
