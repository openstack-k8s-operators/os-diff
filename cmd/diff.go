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
	"os"

	"github.com/openstack-k8s-operators/os-diff/pkg/godiff"
	"github.com/openstack-k8s-operators/os-diff/pkg/servicecfg"

	"github.com/spf13/cobra"
)

// Diff parameters
var debug bool
var remote bool
var quiet bool
var file1Cmd string
var file2Cmd string
var crd bool
var serviceCfgFile string

var diffCmd = &cobra.Command{
	Use:   "diff [path1] [path2]",
	Short: "Compare two files or directories",
	Long: `Print diff for paths provided via the command line: For example:

Example for two files:

./os-diff diff tests/podman/keystone.conf tests/ocp/keystone.conf

Example for remote diff:

CMD1="ssh -F ssh.config standalone podman exec a6e1ca049eee"
CMD2="oc exec glance-external-api-6cf6c98564-blggc -c glance-api --"
./os-diff diff /etc/glance/glance-api.conf /etc/glance/glance.conf.d/00-config.conf --file1-cmd "$CMD1" --file2-cmd "$CMD2" --remote

OR, here only file 1 is remote:
CMD1=oc exec -t neutron-cd94d8ccb-vq2gk -c neutron-api --
./os-diff diff /etc/neutron/neutron.conf /tmp/collect_tripleo_configs/neutron/etc/neutron/neutron.conf --file1-cmd="$CMD1" --remote

Example for directories:

./os-diff diff tests/podman-containers/ tests/ocp-pods/


./os-diff diff ovs_external_ids.json edpm.crd --crd edpm

/!\ Important: remote option is only available for files comparison.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: Insufficient arguments. Please provide at least two file names.")
			return
		}
		path1 := args[0]
		path2 := args[1]
		if crd {
			servicecfg.DiffEdpmCrdFromFile(path1, path2, "ovs_external_ids", serviceCfgFile)
			return
		}
		if remote {
			godiff.CompareFilesFromRemote(path1, path2, file1Cmd, file2Cmd, debug)
		} else {
			fi1, err := os.Stat(path1)
			if err != nil {
				fmt.Println(err)
				return
			}
			fi2, err := os.Stat(path2)
			if err != nil {
				fmt.Println(err)
				return
			}
			if fi1.IsDir() || fi2.IsDir() || quiet {
				goDiff := &godiff.GoDiffDataStruct{
					Origin:      path1,
					Destination: path2,
				}
				err := goDiff.ProcessDirectories(false)
				if err != nil {
					return
				}
			} else {
				godiff.CompareFiles(path1, path2, true, debug)
			}
		}
	},
}

func init() {
	diffCmd.Flags().StringVarP(&file1Cmd, "file1-cmd", "", "", "Remote command for the file1 configuration file.")
	diffCmd.Flags().StringVarP(&file2Cmd, "file2-cmd", "", "", "Remote command for the file2 configuration file.")
	diffCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug.")
	diffCmd.Flags().BoolVar(&quiet, "quiet", false, "Do not print difference on the console and use logs report, only for files comparison")
	diffCmd.Flags().BoolVar(&remote, "remote", false, "Run the diff remotely.")
	diffCmd.Flags().BoolVar(&crd, "crd", false, "Compare a CRDs with a config file.")
	diffCmd.Flags().StringVarP(&serviceCfgFile, "service-config", "f", "config.yaml", "Path for the Yaml config where the services are described.")
	rootCmd.AddCommand(diffCmd)
}
