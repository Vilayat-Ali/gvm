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
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Go versions",
	Long: `List all Go versions currently installed on your system.

This command shows all Go versions that have been installed using GVM,
highlighting the currently active version and indicating which versions
are set as the system default.`,
	Run: func(cmd *cobra.Command, args []string) {
		showDownloaded, _ := cmd.Flags().GetBool("downloaded")
		showCurrent, _ := cmd.Flags().GetBool("current")

		if showDownloaded {
			config, err := internal.LoadConfig()
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
				os.Exit(1)
			}

			currentVersion, err := internal.GetCurrentGolangVersion()
			if err != nil {
				fmt.Printf("Error: %s", err.Error())
				os.Exit(1)
			}

			ltsFound := false

			for version := range config.DownloadedVersions {
				version_print_stmt := version

				isReleaseCandidate := strings.Contains(version, "rc")
				isCurrentVersion := version == *currentVersion

				if !ltsFound && !isReleaseCandidate {
					version_print_stmt += " [LTS] "
					ltsFound = true
				}

				if isCurrentVersion {
					version_print_stmt += " (current)"
				}

				if strings.Contains(version_print_stmt, "[LTS]") || isCurrentVersion {
					color.Green(version_print_stmt)
				} else if isReleaseCandidate {
					color.Red(version_print_stmt)
				} else {
					color.Magenta(version_print_stmt)
				}
			}

			return
		}

		if showCurrent {
			currentVersion, err := internal.GetCurrentGolangVersion()
			if err != nil {
				color.Red("Error: %s", err.Error())
				os.Exit(1)
			}

			color.Green(*currentVersion)
			return
		}

		config, err := internal.LoadConfig()
		if err != nil {
			color.Red("Error: %s", err.Error())
			os.Exit(1)
		}

		currentVersion, err := internal.GetCurrentGolangVersion()
		if err != nil {
			color.Red("Error: %s", err.Error())
			os.Exit(1)
		}

		ltsFound := false

		for _, remoteVersion := range config.AvailableVersions {
			version_print_stmt := remoteVersion.Version

			isReleaseCandidate := strings.Contains(remoteVersion.Version, "rc")
			isCurrentVersion := remoteVersion.Version == *currentVersion

			if !ltsFound && !isReleaseCandidate {
				version_print_stmt += " [LTS] "
				ltsFound = true
			}

			if isCurrentVersion {
				version_print_stmt += " (current)"
			}

			if strings.Contains(version_print_stmt, "[LTS]") || isCurrentVersion {
				color.Green(version_print_stmt)
			} else if isReleaseCandidate {
				color.Red(version_print_stmt)
			} else {
				color.Magenta(version_print_stmt)
			}
		}
	},
}

var updateListCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates available Go versions list",
	Long: `Updates all Go versions currently enlisted on your system for download.

This command updates the available list of all Go versions that can be downloaded`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := internal.LoadConfig()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		if err := config.UpdateAvailableVersions(); err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		if err := config.Save(); err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	listCmd.AddCommand(updateListCmd)
	rootCmd.AddCommand(listCmd)

	// Define flags for the list command
	listCmd.Flags().BoolP("downloaded", "d", false, "Show downloaded versions only")
	listCmd.Flags().BoolP("current", "c", false, "Show current active version only")
}
