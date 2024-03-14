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
	"io/ioutil"
	"strings"

	"github.com/openstack-k8s-operators/os-diff/pkg/godiff"
	"gopkg.in/yaml.v3"
)

type Cinder struct {
	Spec struct {
		Cinder struct {
			Enabled  bool `yaml:"enabled"`
			Template struct {
				CinderAPI struct {
					CustomServiceConfig string `yaml:"customServiceConfig"`
					ExternalEndpoints   []struct {
						Endpoint        string   `yaml:"endpoint"`
						IPAddressPool   string   `yaml:"ipAddressPool"`
						LoadBalancerIPs []string `yaml:"loadBalancerIPs"`
					} `yaml:"externalEndpoints"`
					Replicas int `yaml:"replicas"`
				} `yaml:"cinderAPI"`
				CinderBackup struct {
					NetworkAttachments []string `yaml:"networkAttachments"`
					Replicas           int      `yaml:"replicas"`
				} `yaml:"cinderBackup"`
				CinderScheduler struct {
					CustomServiceConfig string `yaml:"customServiceConfig"`
					Replicas            int    `yaml:"replicas"`
				} `yaml:"cinderScheduler"`
				CinderVolumes struct {
					NetworkAttachments []string `yaml:"networkAttachments"`
					TripleoIscsi       struct {
						CustomServiceConfig string `yaml:"customServiceConfig"`
					} `yaml:"tripleo-iscsi"`
				} `yaml:"cinderVolumes"`
				CustomServiceConfig string   `yaml:"customServiceConfig"`
				DatabaseInstance    string   `yaml:"databaseInstance"`
				Secret              string   `yaml:"secret"`
				ServiceUser         []string `yaml:"serviceUser"`
			} `yaml:"template"`
		} `yaml:"cinder"`
	} `yaml:"spec"`
}

func LoadCinderOpenShiftConfig(configPath string) string {

	// String builder for Cinder Config
	var sb strings.Builder
	// Cinder structure
	var cinder Cinder

	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &cinder)
	if err != nil {
		panic(err)
	}
	if strings.HasPrefix(cinder.Spec.Cinder.Template.CinderAPI.CustomServiceConfig, "[") {
		sb.WriteString(cinder.Spec.Cinder.Template.CinderAPI.CustomServiceConfig)
	}
	if strings.HasPrefix(cinder.Spec.Cinder.Template.CinderScheduler.CustomServiceConfig, "[") {
		sb.WriteString(cinder.Spec.Cinder.Template.CinderScheduler.CustomServiceConfig)
	}
	if strings.HasPrefix(cinder.Spec.Cinder.Template.CinderVolumes.TripleoIscsi.CustomServiceConfig, "[") {
		sb.WriteString(cinder.Spec.Cinder.Template.CinderVolumes.TripleoIscsi.CustomServiceConfig)
	}
	if strings.HasPrefix(cinder.Spec.Cinder.Template.CustomServiceConfig, "[") {
		sb.WriteString(cinder.Spec.Cinder.Template.CustomServiceConfig)
	}

	return godiff.CleanIniSections(sb.String())
}
