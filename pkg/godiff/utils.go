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
	"reflect"

	"github.com/go-yaml/yaml"
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

func compareJSON(orgData, destData interface{}, path string) ([]string, error) {
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
				compareJSON(value, value2, path+"."+key)
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
			compareJSON(orgData[i], destData[i], fmt.Sprintf("%s[%d]", path, i))
		}
	default:
		if !reflect.DeepEqual(orgData, destData) {
			//fmt.Println("Value mismatch at %s: %v != %v\n", path, orgData, destData)
			return diff, nil
		}
	}
	return diff, nil
}
