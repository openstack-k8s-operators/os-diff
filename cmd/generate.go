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

// represents the generate command
var serviceName string
var configFileName string
var outputFile string
var serviceEnable bool
var pullRemote bool
var podmanContainerName string

var generateCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate config patch from an ini config file",
	Long: `Config helpers, generate config patch a config file, example:
	./os-diff gen --service glance --config my-conf.ini --output glance.patch`,
	Run: func(cmd *cobra.Command, args []string) {
		if pullRemote {
			servicecfg.GenerateConfigPatchFromRemote(serviceName, configFileName, outputFile, serviceEnable, podmanContainerName)
		} else {
			err := servicecfg.GenerateConfigPatchFromIni(serviceName, configFileName, outputFile, serviceEnable)
			if err != nil {
				panic(err)
			}
		}
	},
}

func init() {
	generateCmd.Flags().StringVarP(&serviceName, "service", "s", "", "OpenStack service, could be one of: Cinder, Glance...")
	generateCmd.Flags().StringVarP(&configFileName, "config", "c", "", "Configuration file from which you want to generate config patch.")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file name for the config patch.")
	generateCmd.Flags().BoolVar(&serviceEnable, "enable", false, "Enable the service.")
	generateCmd.Flags().BoolVar(&pullRemote, "remote", false, "Cat config file from a remote podman container.")
	generateCmd.Flags().StringVarP(&podmanContainerName, "container-id", "i", "", "Podman container id.")
	rootCmd.AddCommand(generateCmd)
}
