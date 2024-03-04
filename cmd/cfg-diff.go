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
package cmd

import (
	"github.com/openstack-k8s-operators/os-diff/pkg/servicecfg"

	"github.com/spf13/cobra"
)

// diff patch commands
var service string
var configPatchFile string
var configFile string
var frompod bool
var frompodman bool
var podname string
var sidebyside bool

var cfgDiffCmd = &cobra.Command{
	Use:   "cdiff",
	Short: "Print diff between an OpenShift config patch and an OpenStack service config file",
	Long: `Print diff from an OpenShift config patch file and an OpenStack sercice config file.
For example:
./os-diff cdiff -s cinder --configpatch examples/cinder/cinder.patch --configfile examples/cinder/cinder.conf
or
./os-diff cdiff --service cinder --configpatch cinder.patch --configfile /etc/cinder.conf --frompod --podname cinder-api`,
	Run: func(cmd *cobra.Command, args []string) {
		if frompod {
			if podname == "" {
				panic("Please provide a pod name with --frompod option.")
			}
			servicecfg.DiffServiceConfigFromPod(service, configPatchFile, configFile, podname)
		} else if frompodman {
			if podname == "" {
				panic("Please provide a pod name with --frompodman option.")
			}
			servicecfg.DiffServiceConfigFromPodman(service, configPatchFile, configFile, podname)
		} else {
			servicecfg.DiffServiceConfig(service, configPatchFile, configFile, false)
		}
	},
}

func init() {
	cfgDiffCmd.Flags().StringVarP(&configPatchFile, "configpatch", "o", "", "OpenShift configmap patch file path.")
	cfgDiffCmd.Flags().StringVarP(&configFile, "configfile", "c", "", "OpenStack service INI config file path.")
	cfgDiffCmd.Flags().StringVarP(&service, "service", "s", "", "OpenStack service, could be one of: Cinder, Glance...")
	cfgDiffCmd.Flags().BoolVar(&frompod, "frompod", false, "Get config file directly from OpenShift service Pod.")
	cfgDiffCmd.Flags().BoolVar(&frompodman, "frompodman", false, "Get config file directly from OpenStack podman container.")
	cfgDiffCmd.Flags().StringVarP(&podname, "podname", "p", "", "Name of the pod of the service: cinder-api.")
	rootCmd.AddCommand(cfgDiffCmd)
}
