/*
Copyright © 2025 Syed Vilayat Ali Rizvi

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "1.2.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gvm",
	Short: "Go Version Manager - Install and manage multiple Go versions",
	Long: `GVM - Go Version Manager

A simple, powerful command-line tool for managing multiple Go versions on your system.

GVM allows you to:
  • Install multiple versions of Go
  • Switch between Go versions seamlessly
  • Set default Go versions globally or per-project
  • List available and installed Go versions
  • Keep your system Go installation untouched

Examples:
  gvm install 1.20.3        # Install a specific Go version
  gvm use 1.19.8            # Switch to Go 1.19.8
  gvm list                  # List all installed versions
  gvm list-remote           # List all available versions
  gvm default 1.20.3        # Set Go 1.20.3 as default

Documentation: https://vilayat-ali.github.io/gvm
Source Code:   https://github.com/vilayat-ali/gvm`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no arguments provided
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")

	// Set up version template
	rootCmd.SetVersionTemplate(`GVM - Go Version Manager v{{.Version}}
Built with ❤️ for the Go community
`)
}
