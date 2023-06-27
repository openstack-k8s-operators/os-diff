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

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Print diff for two specific files",
	Long: `Print diff for files provided via the command line: For example:
os-diff diff --origin=tests/podman/keystone.conf --destination=tests/ocp/keystone.conf`,
	Run: func(cmd *cobra.Command, args []string) {
		goDiff := &godiff.CompareFileNames{
			Origin:      source,
			Destination: dest,
		}

		err := goDiff.DiffFiles()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	diffCmd.Flags().StringVarP(&source, "origin", "o", "", "Source file.")
	diffCmd.Flags().StringVarP(&dest, "destination", "d", "", "Destination file.")
	rootCmd.AddCommand(diffCmd)
}
