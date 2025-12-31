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
				color.Red("✗ Error loading configuration: %s", err.Error())
				os.Exit(1)
			}

			if len(config.DownloadedVersions) == 0 {
				color.Yellow("📭 No downloaded Go versions found.")
				color.Cyan("\nTry: gvm list      # to see available versions")
				color.Cyan("     gvm install   # to install a version")
				return
			}

			currentVersion, err := internal.GetCurrentGolangVersion()
			if err != nil {
				color.Red("✗ Error detecting current version: %s", err.Error())
				os.Exit(1)
			}

			fmt.Println()
			color.Cyan("📦 Downloaded Go Versions")
			fmt.Println(strings.Repeat("─", 50))

			ltsFound := false

			for version := range config.DownloadedVersions {
				version_print_stmt := version

				isReleaseCandidate := strings.Contains(version, "rc")
				isCurrentVersion := version == *currentVersion

				if !ltsFound && !isReleaseCandidate {
					version_print_stmt += " 🏷️ LTS"
					ltsFound = true
				}

				if isCurrentVersion {
					version_print_stmt += " ✅"
				}

				// Create bullet point with colored text
				bullet := "  • "
				if isCurrentVersion {
					bullet = "  ▶ "
				}

				if isCurrentVersion {
					color.New(color.FgGreen, color.Bold).Printf("%s%s\n", bullet, version_print_stmt)
				} else if strings.Contains(version_print_stmt, "🏷️ LTS") {
					color.New(color.FgCyan, color.Bold).Printf("%s%s\n", bullet, version_print_stmt)
				} else if isReleaseCandidate {
					color.New(color.FgYellow).Printf("%s%s\n", bullet, version_print_stmt)
				} else {
					color.New(color.FgMagenta).Printf("%s%s\n", bullet, version_print_stmt)
				}
			}

			fmt.Println()
			color.HiBlack("Legend: ✅ = Current | 🏷️ = LTS | • = Installed")
			return
		}

		if showCurrent {
			currentVersion, err := internal.GetCurrentGolangVersion()
			if err != nil {
				color.Red("✗ Error detecting current version: %s", err.Error())
				os.Exit(1)
			}

			fmt.Println()
			color.Cyan("⚡ Current Go Version")
			fmt.Println(strings.Repeat("─", 30))
			color.New(color.FgGreen, color.Bold).Printf("  %s\n", *currentVersion)
			fmt.Println()
			return
		}

		config, err := internal.LoadConfig()
		if err != nil {
			color.Red("✗ Error loading configuration: %s", err.Error())
			os.Exit(1)
		}

		if len(config.AvailableVersions) == 0 {
			color.Yellow("📭 No Go versions available in cache.")
			color.Cyan("\nTry: gvm list update  # to update the versions list")
			return
		}

		currentVersion, err := internal.GetCurrentGolangVersion()
		if err != nil {
			color.Red("✗ Error detecting current version: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println()
		color.Cyan("📚 Available Go Versions")
		fmt.Println(strings.Repeat("─", 50))

		ltsFound := false
		versionCount := 0

		for _, remoteVersion := range config.AvailableVersions {
			version_print_stmt := remoteVersion.Version

			isReleaseCandidate := strings.Contains(remoteVersion.Version, "rc")
			isCurrentVersion := remoteVersion.Version == *currentVersion

			if !ltsFound && !isReleaseCandidate {
				version_print_stmt += " 🏷️ LTS"
				ltsFound = true
			}

			if isCurrentVersion {
				version_print_stmt += " ✅"
			}

			// Create bullet point with colored text
			bullet := "  • "
			if isCurrentVersion {
				bullet = "  ▶ "
			}

			if isCurrentVersion {
				color.New(color.FgGreen, color.Bold).Printf("%s%s\n", bullet, version_print_stmt)
			} else if strings.Contains(version_print_stmt, "🏷️ LTS") {
				color.New(color.FgCyan, color.Bold).Printf("%s%s\n", bullet, version_print_stmt)
			} else if isReleaseCandidate {
				color.New(color.FgYellow).Printf("%s%s\n", bullet, version_print_stmt)
			} else {
				color.New(color.FgMagenta).Printf("%s%s\n", bullet, version_print_stmt)
			}

			versionCount++
			if versionCount >= 10 { // Show only first 10 versions
				if len(config.AvailableVersions) > 10 {
					color.HiBlack("\n  ... and %d more versions", len(config.AvailableVersions)-10)
					color.HiBlack("  Use 'gvm list -d' to see downloaded versions")
				}
				break
			}
		}

		fmt.Println()
		color.HiBlack("Legend: ✅ = Current | 🏷️ = LTS | ▶ = Active | • = Available")
		fmt.Println()
		color.Cyan("💡 Tips:")
		color.Cyan("  • Use 'gvm list -d' to see downloaded versions")
		color.Cyan("  • Use 'gvm list -c' to see current version only")
		color.Cyan("  • Use 'gvm list update' to refresh available versions")
	},
}

var updateListCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates available Go versions list",
	Long: `Updates all Go versions currently enlisted on your system for download.

This command updates the available list of all Go versions that can be downloaded`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		color.Cyan("🔄 Updating Go versions list...")
		fmt.Println(strings.Repeat("─", 30))

		config, err := internal.LoadConfig()
		if err != nil {
			color.Red("✗ Error loading configuration: %s", err.Error())
			os.Exit(1)
		}

		color.Blue("  Fetching latest versions from GitHub...")
		if err := config.UpdateAvailableVersions(); err != nil {
			color.Red("✗ Failed to update versions: %s", err.Error())
			os.Exit(1)
		}

		color.Blue("  Saving updated list...")
		if err := config.Save(); err != nil {
			color.Red("✗ Failed to save configuration: %s", err.Error())
			os.Exit(1)
		}

		color.Green("✓ Successfully updated versions list!")
		color.Cyan("\n📊 Found %d Go versions available for download", len(config.AvailableVersions))
		color.Cyan("\nRun 'gvm list' to see the updated list")
		fmt.Println()
	},
}

func init() {
	listCmd.AddCommand(updateListCmd)
	rootCmd.AddCommand(listCmd)

	// Define flags for the list command
	listCmd.Flags().BoolP("downloaded", "d", false, "Show downloaded versions only")
	listCmd.Flags().BoolP("current", "c", false, "Show current active version only")
}
