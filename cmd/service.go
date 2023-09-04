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
var podname string

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Print diff between an Openshift Config spec and an Openstack service config file",
	Long: `Print diff from an Openshit config spec file and an Openstack sercice config file.
		   For example:
           ./os-diff service --service cinder --ocp cinder.patch --serviceconfig cinder.conf
		   or
		   ./os-diff service --service cinder --ocp cinder.patch --serviceconfig /etc/cinder.conf --frompod --podname cinder-api`,
	Run: func(cmd *cobra.Command, args []string) {
		if service == "cinder" {
			if frompod {
				if podname == "" {
					panic("Please provide a pod name with --frompod option.")
				}
				servicecfg.DiffCinderConfigFromPod(ocp, config, podname)
			} else {
				servicecfg.DiffCinderConfig(ocp, config)
			}
		} else {
			panic("Unknown service...")
		}
	},
}

func init() {
	serviceCmd.Flags().StringVarP(&ocp, "ocp", "o", "", "Openshift config spec file path.")
	serviceCmd.Flags().StringVarP(&config, "config", "c", "", "Openstack service config file path.")
	serviceCmd.Flags().StringVarP(&service, "service", "s", "", "Openstack service, could be one of: Cinder, Glance...")
	serviceCmd.Flags().BoolVar(&frompod, "frompod", false, "Get config file directly from OpenShift service Pod.")
	serviceCmd.Flags().StringVarP(&podname, "podname", "p", "", "Name of the pod of the service: cinder-api.")
	rootCmd.AddCommand(serviceCmd)
}
