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
	"path/filepath"

	"github.com/openstack-k8s-operators/os-diff/pkg/common"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var osDiffConfig string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "os-diff",
	Short: "Tool for pulling and inspecting config files for OpenStack services",
	Long: `Pull and compare OpenStack services configuration files from pods
or podman containers. For example:

You can pull configuration from a Keystone container and compare
to a new Keystone pod which has been migrated.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.os-diff.cfg)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// Initialize the config and bind it to the root command
	err := initConfig(rootCmd)
	if err != nil {
		fmt.Println("Error initializing config:", err)
		return
	}
}

func initConfig(cmd *cobra.Command) error {
	// Bind the loaded config to a persistent flag
	cmd.PersistentFlags().StringVarP(&osDiffConfig, "config", "c", "os-diff.cfg", "Config file (default is $PWD/os-diff.cfg)")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	path := CheckFilesPresence(osDiffConfig)
	config, err := common.LoadOSDiffConfig(path)
	if err != nil {
		return err
	}
	// Store the loaded config in Viper for access from within the command
	viper.Set("config", config)

	return nil
}

func CheckFilesPresence(configFile string) string {
	_, err := os.Stat(configFile)
	if err == nil {
		fmt.Println("Found " + configFile + " in the current working directory.")
		return configFile
	}

	// If config.yaml doesn't exist in the current working directory,
	// check if it exists in /etc
	etcConfigFile := filepath.Join("/etc/os-diff/", configFile)
	_, err = os.Stat(etcConfigFile)
	if err == nil {
		fmt.Println("Found os-diff.cfg in /etc/os-diff/.")
		return etcConfigFile
	}

	// If config.yaml doesn't exist in both locations, raise an error and end the program
	fmt.Println("Error: config.yaml not found in the current working directory or /etc/os-diff.")
	os.Exit(1)
	return ""
}
