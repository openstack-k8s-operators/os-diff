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

type Glance struct {
	Spec struct {
		Glance struct {
			Enabled  bool `yaml:"enabled"`
			Template struct {
				DatabaseInstance    string `yaml:"databaseInstance"`
				ContainerImage      string `yaml:"containerImage"`
				CustomServiceConfig string `yaml:"customServiceConfig"`
				StorageClass        string `yaml:"storageClass"`
				StorageRequest      string `yaml:"storageRequest"`
				GlanceAPIInternal   struct {
					ExternalEndpoints []struct {
						Endpoint        string   `yaml:"endpoint"`
						IPAddressPool   string   `yaml:"ipAddressPool"`
						LoadBalancerIPs []string `yaml:"loadBalancerIPs"`
					} `yaml:"externalEndpoints"`
					NetworkAttachments []string `yaml:"networkAttachments"`
				} `yaml:"glanceAPIInternal"`
				GlanceAPIExternal struct {
					NetworkAttachments []string `yaml:"networkAttachments"`
				} `yaml:"glanceAPIExternal"`
			} `yaml:"template"`
		} `yaml:"glance"`
		ExtraMounts []struct {
			ExtraVol []struct {
				Propagation  []string `yaml:"propagation"`
				ExtraVolType string   `yaml:"extraVolType"`
				Volumes      []struct {
					Name      string `yaml:"name"`
					Projected struct {
						Sources []struct {
							Secret struct {
								Name string `yaml:"name"`
							} `yaml:"secret"`
						} `yaml:"sources"`
					} `yaml:"projected"`
				} `yaml:"volumes"`
				Mounts []struct {
					Name      string `yaml:"name"`
					MountPath string `yaml:"mountPath"`
					ReadOnly  bool   `yaml:"readOnly"`
				} `yaml:"mounts"`
			} `yaml:"extraVol"`
		} `yaml:"extraMounts"`
	} `yaml:"spec"`
}

func LoadGlanceOpenShiftConfig(configPath string) string {
	var sb strings.Builder
	// Service structure
	var service Glance

	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &service)
	if err != nil {
		panic(err)
	}
	if strings.HasPrefix(service.Spec.Glance.Template.CustomServiceConfig, "[") {
		sb.WriteString(service.Spec.Glance.Template.CustomServiceConfig)
	}

	return godiff.CleanIniSections(sb.String())
}
