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

	"github.com/openstack-k8s-operators/os-diff/pkg/collectcfg"
	"github.com/openstack-k8s-operators/os-diff/pkg/common"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cloud string
var output_dir string
var update bool
var updateOnly bool
var serviceConfig string
var filters []string

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull configurations from Podman or OCP",
	Long: `This command pulls configuration files by services from Podman
environment or OCP. For example:
./os-diff pull --env=tripleo
You can set configuration in your os-diff.cfg or provide output directory via the command line:
./os-diff pull -e ocp -o /tmp/myconfigdir -s my-service-config-file
You can also update the config.yaml file with the information from your TripleO environment:
./os-diff pull --update-only
This command will add the podman and image IDs in the config.yaml or also:
./os-pull pull --update
This command will populate the config.yaml file with the podman and image Ids and pull the config too.
`,
	Run: func(cmd *cobra.Command, args []string) {

		// Get config:
		config := viper.Get("config").(*common.ODConfig)
		if serviceConfig == "" {
			serviceConfig = config.Default.ServiceConfigFile
		}
		configPath := CheckFilesPresence(serviceConfig)

		if cloud == "ocp" {
			// Test OCP connection:
			if !common.TestOCConnection() {
				fmt.Println("OC not connected, you need to logged in before running this command...")
				return
			}
			// OCP Settings
			localOCPDir := config.Openshift.OcpLocalConfigPath
			err := collectcfg.FetchConfigFromEnv(configPath, localOCPDir, "", false, config.Openshift.Connection, "", "", filters)
			if err != nil {
				fmt.Println("Error while collecting config: ", err)
				return
			}
		} else if cloud == "tripleo" {
			// TRIPLEO Settings:
			sshCmd := config.Tripleo.SshCmd
			fullCmd := sshCmd + " " + config.Tripleo.DirectorHost
			remoteConfigDir := config.Tripleo.RemoteConfigPath
			localConfigDir := config.Tripleo.LocalConfigPath
			if !common.TestSshConnection(fullCmd) {
				fmt.Println("Please check your SSH configuration: " + fullCmd)
				return
			}
			if update || updateOnly {
				collectcfg.SetTripleODataEnv(configPath, fullCmd, filters, true)
				if updateOnly {
					return
				}
			}
			err := collectcfg.FetchConfigFromEnv(configPath, localConfigDir, remoteConfigDir, true, config.Tripleo.Connection, sshCmd, config.Tripleo.DirectorHost, filters)
			if err != nil {
				fmt.Println("Error while collecting config: ", err)
				return
			}
		} else {
			fmt.Println("Error unkown cloud", cloud)
			return
		}
	},
}

func init() {
	pullCmd.Flags().StringVarP(&cloud, "env", "e", "tripleo", "Service engine, could be: ocp or tripleo.")
	pullCmd.Flags().StringVarP(&output_dir, "output_dir", "o", "", "Output directory for the configuration files.")
	pullCmd.Flags().StringVarP(&serviceConfig, "service_config", "s", "", "File where the service configurations are describe.")
	pullCmd.Flags().StringSliceVar(&filters, "filters", []string{}, "Filter Openstack services: --filters glance_api,nova_api,keystone ..")
	pullCmd.Flags().BoolVar(&update, "update", false, "Update config.yaml with Podman informations.")
	pullCmd.Flags().BoolVar(&updateOnly, "update-only", false, "Update only config.yaml with Podman informations and not pull configurations from services.")
	rootCmd.AddCommand(pullCmd)
}
