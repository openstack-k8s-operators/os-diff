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
	"github.com/openstack-k8s-operators/os-diff/pkg/common"
	"github.com/spf13/cobra"
)

// represents the generate command
var inventoryFile string
var sshConfig string
var yaml bool
var etc bool

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure os-diff before running the diff command",
	Long: `Config helpers, configure os-diff before running the diff command, examples:
	from txt/ini format:
	./os-diff configure --inventory inventory --output ssh_config
	from /etc/hosts format:
	./os-diff configure --inventory inventory --output ssh_config --etc
	from yaml format:
	./os-diff configure --inventory inventory --output ssh_config --yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		common.BuildSshConfigFile(inventoryFile, sshConfig, yaml, etc)
	},
}

func init() {
	configureCmd.Flags().StringVarP(&inventoryFile, "inventory", "i", "", "Inventory file")
	configureCmd.Flags().StringVarP(&sshConfig, "ouput", "o", "", "Ssh config output file")
	configureCmd.Flags().BoolVar(&yaml, "yaml", false, "Set this if the inventory is in yaml format.")
	configureCmd.Flags().BoolVar(&etc, "etc", false, "Set this if the inventory is from /etc/hosts format.")
	rootCmd.AddCommand(configureCmd)
}
