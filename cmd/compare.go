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

// compareCmd represents the compare command
var output string
var reverse bool

var compareCmd = &cobra.Command{
	Use:   "compare [PATH1] [PATH2]",
	Short: "Compare files and directories from PATH1 and PATH2",
	Long: `Compare files or directories from two different paths. For example:
		./os-diff compare $PATH1 $PATH2

		./os-diff compare tests/podman/keystone.conf tests/ocp/keystone.conf --output=output.txt
		or
		os-diff compare tests/podman-containers/ tests/ocp-pods/ --output=output.txt`,
	Run: func(cmd *cobra.Command, args []string) {
		path1 := args[0]
		path2 := args[1]
		goDiff := &godiff.GoDiffDataStruct{
			Origin:      path1,
			Destination: path2,
		}
		err := goDiff.ProcessDirectories(reverse)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	compareCmd.Flags().StringVar(&output, "output", "output.txt", "Output file (default is $PWD/output.txt)")
	compareCmd.Flags().BoolVar(&reverse, "reverse", false, "Search difference in both directories: origin and destination.")
	rootCmd.AddCommand(compareCmd)
}
