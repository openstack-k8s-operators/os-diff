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
	"os-diff/pkg/collectcfg"
	"os-diff/pkg/common"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cloud string
var output_dir string
var verbose bool
var serviceConfig string

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull configurations from Podman or OCP",
	Long: `This command pulls configuration files by services from Podman
environment or OCP. For example:
./os-diff pull --env=tripleo
You can set configuration in your os-diff.cfg or provide output directory via the command line:
./os-diff pull -e ocp -o /tmp/myconfigdir -s my-service-config-file`,
	Run: func(cmd *cobra.Command, args []string) {

		// Get config:
		config := viper.Get("config").(*common.ODConfig)
		if serviceConfig == "" {
			serviceConfig = config.Default.ServiceConfigFile
		}

		if cloud == "ocp" {
			// Test OCP connection:
			if !common.TestOCConnection() {
				fmt.Println("OC not connected, you need to logged in before running this command...")
				return
			}
			// OCP Settings
			localOCPDir := config.Openshift.OcpLocalConfigPath
			err := collectcfg.FetchConfigFromEnv(serviceConfig, localOCPDir, "", false, config.Openshift.Connection, "")
			if err != nil {
				fmt.Println("Error while collecting config: ", err)
				return
			}
		} else if cloud == "tripleo" {
			// TRIPLEO Settings:
			standaloneSsh := config.Tripleo.SshCmd
			remoteConfigDir := config.Tripleo.RemoteConfigPath
			localConfigDir := config.Default.LocalConfigDir
			if !common.TestSshConnection(standaloneSsh) {
				fmt.Println("Please check your SSH configuration: " + standaloneSsh)
				return
			}
			err := collectcfg.FetchConfigFromEnv(serviceConfig, localConfigDir, remoteConfigDir, true, config.Tripleo.Connection, standaloneSsh)
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
	pullCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable Ansible verbosity.")
	rootCmd.AddCommand(pullCmd)
}
