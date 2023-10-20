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
	"fmt"
	"os-diff/pkg/servicecfg"

	"github.com/spf13/cobra"
)

// diff patch commands
var configMap string
var configPath string
var fromRemote bool
var remoteCmd string

var cfgMapDiffCmd = &cobra.Command{
	Use:   "cfgmap-diff",
	Short: "Print diff between OpenShift configmap and OpenStack/TripleO config files",
	Long: `Print diff from OpenShift configmap and OpenStack/TripleO config files.
For example:
./os-diff cfgmap-diff --configmap keystone-config-data --config /tmp/collect_tripleo_configs/keystone/etc/keystone
or
from a configmap file:
./os-diff cfgmap-diff --configmap keystone-config-data.yaml --config /tmp/collect_tripleo_configs/keystone/etc/keystone
or
CMD1="ssh -F ssh.config standalone podman exec a6e1ca049eee"
./os-diff cfgmap-diff --configmap keystone-config-data --config /etc/keystone --remote --remode-cmd $CMD`,
	Run: func(cmd *cobra.Command, args []string) {
		if fromRemote {
			if remoteCmd == "" {
				fmt.Println("Error: you must provide a --remote-cmd with --remote")
				return
			}
		}
		err := servicecfg.DiffConfigMap(configMap, configPath, fromRemote, remoteCmd)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	cfgMapDiffCmd.Flags().StringVarP(&configMap, "configmap", "m", "", "OpenShift configmap: oc get configmap/<name>")
	cfgMapDiffCmd.Flags().StringVarP(&configPath, "config", "c", "", "OpenStack service INI config file path.")
	cfgMapDiffCmd.Flags().BoolVar(&fromRemote, "remote", false, "Get Tripleo config remotely.")
	cfgMapDiffCmd.Flags().StringVarP(&remoteCmd, "remote-cmd", "", "", "Remote Ssh command for pulling Tripleo config.")
	rootCmd.AddCommand(cfgMapDiffCmd)
}
