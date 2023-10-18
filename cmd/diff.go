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
	"os-diff/pkg/godiff"

	"github.com/spf13/cobra"
)

// Diff parameters
var debug bool
var remote bool
var quiet bool
var file1Cmd string
var file2Cmd string

var diffCmd = &cobra.Command{
	Use:   "diff [path1] [path2]",
	Short: "Compare two files or directories",
	Long: `Print diff for paths provided via the command line: For example:

Example for two files:

./os-diff diff tests/podman/keystone.conf tests/ocp/keystone.conf

Example for remote diff:

export CMD1="ssh -F ssh.config standalone podman exec a6e1ca049eee"
export CMD2="oc exec glance-external-api-6cf6c98564-blggc -c glance-api --"
./os-diff diff /etc/glance/glance-api.conf /etc/glance/glance.conf.d/00-config.conf --file1-cmd "$CMD1" --file2-cmd "$CMD2" --remote

Example for directories:

./os-diff diff tests/podman-containers/ tests/ocp-pods/

/!\ Important: remote option is only available for files comparison.`,
	Run: func(cmd *cobra.Command, args []string) {
		path1 := args[0]
		path2 := args[1]

		fi1, err := os.Stat(path1)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fi2, err := os.Stat(path2)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		if fi1.IsDir() || fi2.IsDir() || quiet {
			goDiff := &godiff.GoDiffDataStruct{
				Origin:      path1,
				Destination: path2,
			}
			err := goDiff.ProcessDirectories(false)
			if err != nil {
				panic(err)
			}
		} else {
			if remote {
				godiff.CompareFilesFromRemote(path1, path2, file1Cmd, file2Cmd, debug)
			} else {
				godiff.CompareFiles(path1, path2, true, debug)
			}

		}
	},
}

func init() {
	diffCmd.Flags().StringVarP(&file1Cmd, "file1-cmd", "", "", "Remote command for the file1 configuration file.")
	diffCmd.Flags().StringVarP(&file1Cmd, "file2-cmd", "", "", "Remote command for the file2 configuration file.")
	diffCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug.")
	diffCmd.Flags().BoolVar(&quiet, "quiet", false, "Do not print difference on the console and use logs report, only for files comparison")
	diffCmd.Flags().BoolVar(&remote, "remote", false, "Run the diff remotely.")
	rootCmd.AddCommand(diffCmd)
}
