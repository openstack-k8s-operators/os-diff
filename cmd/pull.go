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
	"os-diff/pkg/ansible"

	"github.com/spf13/cobra"
)

var inventory string
var cloud_engine string
var output_dir string
var play string
var verbose bool
var extraVars map[string]string

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull configurations from Podman or OCP",
	Long: `This command pulls configuration files by services from Podman
	environment or OCP. For example:
    os-diff pull --cloud_engine=ocp --inventory=$PWD/hosts --output-dir=/tmp`,
	Run: func(cmd *cobra.Command, args []string) {

		ansiblePlaybookConnectionOptions := &ansible.AnsiblePlaybookConnectionOptions{
			Connection: "local",
		}

		envMap := make(map[string]interface{})
		for key, value := range extraVars {
			envMap[key] = value
		}
		ansiblePlaybookOptions := &ansible.AnsiblePlaybookOptions{
			Inventory: inventory,
			Verbosity: verbose,
			ExtraVars: envMap,
		}

		if cloud_engine == "ocp" {
			play = "playbooks/collect_ocp_config.yaml"
		} else {
			play = "playbooks/collect_podman_config.yaml"
		}

		playbook := &ansible.AnsiblePlaybookCmd{
			Playbook:          play,
			ConnectionOptions: ansiblePlaybookConnectionOptions,
			Options:           ansiblePlaybookOptions,
		}

		err := playbook.Run()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	pullCmd.Flags().StringVarP(&inventory, "inventory", "i", "hosts", "Ansible inventory hosts file.")
	pullCmd.Flags().StringVarP(&cloud_engine, "cloud_engine", "c", "ocp", "Service engine, could be: ocp or podman.")
	pullCmd.Flags().StringVar(&output_dir, "output_dir", "/tmp", "Output directory for the configuration files.")
	pullCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable Ansible verbosity.")
	pullCmd.Flags().StringToStringVarP(&extraVars, "extra-vars", "e", nil, "Ansible extra vars")
	rootCmd.AddCommand(pullCmd)
}
