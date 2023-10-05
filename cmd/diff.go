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
	"os-diff/pkg/godiff"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var source string
var dest string
var debug bool
var remote bool
var sourceCmd string
var destCmd string

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Print diff for two specific files",
	Long: `Print diff for files provided via the command line: For example:
./os-diff os-diff diff --source=tests/podman/keystone.conf --destination=tests/ocp/keystone.conf
Example for remote diff:
export CMD1="ssh -F ssh.config standalone podman exec a6e1ca049eee"
export CMD2="oc exec glance-external-api-6cf6c98564-blggc -c glance-api --"
./os-diff diff --dest-cmd "$CMD2" --orgin-cmd "$CMD1" -o /etc/glance/glance-api.conf -d /etc/glance/glance.conf.d/00-config.conf --remote`,
	Run: func(cmd *cobra.Command, args []string) {
		if remote {
			godiff.CompareFilesFromRemote(source, dest, sourceCmd, destCmd, debug)
		} else {
			godiff.CompareFiles(source, dest, true, debug)
		}
	},
}

func init() {
	diffCmd.Flags().StringVarP(&source, "source", "o", "", "Source file.")
	diffCmd.Flags().StringVarP(&dest, "destination", "d", "", "Destination file.")
	diffCmd.Flags().StringVarP(&sourceCmd, "source-cmd", "", "", "Remote command for the source configuration file.")
	diffCmd.Flags().StringVarP(&destCmd, "dest-cmd", "", "", "Remote command for the destination configuration file.")
	diffCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug.")
	diffCmd.Flags().BoolVar(&remote, "remote", false, "Run the diff remotely.")
	rootCmd.AddCommand(diffCmd)
}
