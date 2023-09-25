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
	"os-diff/pkg/servicecfg"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var service string
var ocp string
var config string
var frompod bool
var frompodman bool
var podname string
var sidebyside bool

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Print diff between an OpenShift Config spec and an OpenStack service config file",
	Long: `Print diff from an OpenShift config spec file and an OpenStack sercice config file.
		   For example:
           ./os-diff service --service cinder --ocp examples/cinder/cinder.patch --serviceconfig examples/cinder/cinder.conf
		   or
		   ./os-diff service --service cinder --ocp cinder.patch --serviceconfig /etc/cinder.conf --frompod --podname cinder-api`,
	Run: func(cmd *cobra.Command, args []string) {
		if frompod {
			if podname == "" {
				panic("Please provide a pod name with --frompod option.")
			}
			servicecfg.DiffServiceConfigFromPod(service, ocp, config, podname)
		} else if frompodman {
			if podname == "" {
				panic("Please provide a pod name with --frompodman option.")
			}
			servicecfg.DiffServiceConfigFromPodman(service, ocp, config, podname)
		} else {
			servicecfg.DiffServiceConfig(service, ocp, config, sidebyside)
		}
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&ocp, "ocp", "o", "", "OpenShift config spec file path.")
	serviceCmd.Flags().StringVarP(&config, "config", "c", "", "OpenStack service config file path.")
	serviceCmd.Flags().StringVarP(&service, "service", "s", "", "OpenStack service, could be one of: Cinder, Glance...")
	serviceCmd.Flags().BoolVar(&frompod, "frompod", false, "Get config file directly from OpenShift service Pod.")
	serviceCmd.Flags().BoolVar(&frompodman, "frompodman", false, "Get config file directly from OpenStack podman container.")
	serviceCmd.Flags().BoolVar(&sidebyside, "sidebyside", false, "Compare both: source->dest and dest->source.")
	serviceCmd.Flags().StringVarP(&podname, "podname", "p", "", "Name of the pod of the service: cinder-api.")
	rootCmd.AddCommand(serviceCmd)
}
