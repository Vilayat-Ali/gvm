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
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Set up initial configuration for gvm (Go Version Manager)",
	Long: `The configure command initializes gvm by creating the necessary configuration files and directory structure.

This command performs the following actions:
1. Creates a configuration file at ~/.config/gvm/config.json
2. Sets up the required directories for storing Go versions at /usr/local/gvm/go-versions/

If configuration already exists, this command will inform you that gvm is already set up.

Examples:
  gvm configure      # Initializes gvm with default settings
  gvm configure -h   # Shows help information for this command`,
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.ConfigExists() {
			color.Blue("Setting up gvm configuration...")
			if err := internal.SetupConfig(); err != nil {
				color.Red("Failed to configure gvm: %s", err.Error())
				os.Exit(1)
			}
			color.Green("✓ gvm configured successfully!")
			color.Cyan("\nNext steps:")
			color.Cyan("  • Run 'gvm list' to see available Go versions")
			color.Cyan("  • Run 'gvm download <version>' to install a Go version")
			color.Cyan("  • Run 'gvm use <version>' to switch to a specific Go version")
		} else {
			color.Yellow("gvm is already configured. Run 'gvm --help' to see available commands.")
		}
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
