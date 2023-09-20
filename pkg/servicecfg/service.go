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

func DiffServiceConfig(service string, ocpConfig string, serviceConfig string, sidebyside bool) error {
	var servicePatch string
	// Get ocpConfig
	if service == "cinder" {
		servicePatch = LoadCinderOpenshiftConfig(ocpConfig)
	} else if service == "glance" {
		servicePatch = LoadGlanceOpenshiftConfig(ocpConfig)
	} else {
		msg := `Service not supported, please implement it.
			Follow the instructions to add new Openstack services here:
			https://github.com/openstack-k8s-operators/os-diff#add-service`
		panic(msg)
	}

	// Get service Config
	osConfig, err := LoadServiceConfig(serviceConfig)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig([]byte(servicePatch), osConfig, ocpConfig, serviceConfig)
	if err != nil {
		panic(err)
	}
	if sidebyside {
		_, err = CompareIniConfig(osConfig, []byte(servicePatch), serviceConfig, ocpConfig)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func DiffServiceConfigFromPod(service string, ocpConfig string, serviceConfig string, containerName string) error {
	var servicePatch string
	var podName string
	// Get ocpConfig
	if service == "cinder" {
		podName = "cinder"
		servicePatch = LoadCinderOpenshiftConfig(ocpConfig)
	} else if service == "glance" {
		// @todo: should be move a config spec file, users must be describe their env in a file.cfg.
		podName = "glance-external-api"
		servicePatch = LoadGlanceOpenshiftConfig(ocpConfig)
	} else {
		msg := `Service not supported, please implement it.
			Follow the instructions to add new Openstack services here:
			https://github.com/openstack-k8s-operators/os-diff#add-service`
		panic(msg)
	}
	// Get service Config
	podConfig, err := GetConfigFromPod(serviceConfig, podName, containerName)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig([]byte(servicePatch), podConfig, ocpConfig, serviceConfig)
	if err != nil {
		panic(err)
	}
	return nil
}

func DiffServiceConfigFromPodman(service string, ocpConfig string, serviceConfig string, podname string) error {
	var servicePatch string
	// Get ocpConfig
	if service == "cinder" {
		servicePatch = LoadCinderOpenshiftConfig(ocpConfig)
	} else if service == "glance" {
		servicePatch = LoadGlanceOpenshiftConfig(ocpConfig)
	} else {
		msg := `Service not supported, please implement it.
			Follow the instructions to add new Openstack services here:
			https://github.com/openstack-k8s-operators/os-diff#add-service`
		panic(msg)
	}
	// Get service Config
	osConfig, err := GetConfigFromPodman(serviceConfig, podname)
	if err != nil {
		panic(err)
	}

	_, err = CompareIniConfig([]byte(servicePatch), osConfig, ocpConfig, serviceConfig)
	if err != nil {
		panic(err)
	}
	return nil
}
