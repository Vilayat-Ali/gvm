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
	"slices"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

func findDownloadedVersion(config *internal.Config, version string) *internal.DownloadVersion {
	normalizedVersion := internal.NormalizeVersion(version)
	for _, downloadedVersion := range config.DownloadedVersions {
		if internal.NormalizeVersion(downloadedVersion.Version) == normalizedVersion {
			return &downloadedVersion
		}
	}
	return nil
}

func isVersionAvailable(config *internal.Config, version string) bool {
	normalizedVersion := internal.NormalizeVersion(version)
	for _, availableVersion := range config.AvailableVersions {
		if internal.NormalizeVersion(availableVersion.Version) == normalizedVersion {
			return true
		}
	}
	return false
}

func getDownloadedVersion(config *internal.Config, version string) *internal.DownloadVersion {
	normalizedVersion := internal.NormalizeVersion(version)
	for _, downloadedVersion := range config.DownloadedVersions {
		if internal.NormalizeVersion(downloadedVersion.Version) == normalizedVersion {
			return &downloadedVersion
		}
	}
	return nil
}

var useCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "switch to a Go version 🔄",
	Long: `Activate a Go version and start coding!

This command sets the specified Go version as your active version.
If you haven't downloaded it yet, gvm will grab it for you first.

Examples:
  gvm use 1.22.0    → switch to Go 1.22.0
  gvm use 1.21.0    → switch to Go 1.21.0`,
	Run: func(cmd *cobra.Command, args []string) {
		faint := color.New(color.Faint)

		if len(args) == 0 {
			color.Red("❌ Oops! You forgot the version")
			color.Cyan("  Usage: gvm use <version>")
			color.White("  Example: gvm use 1.22.0")
			os.Exit(1)
		}

		requestedVersion := args[0]
		if !internal.ValidateGoVersion(requestedVersion) {
			color.Red("❌ '%s' doesn't look like a valid Go version", requestedVersion)
			os.Exit(1)
		}

		gvmConfig, err := internal.LoadConfig()
		if err != nil {
			color.Red("❌ Couldn't load config: %s", err.Error())
			os.Exit(1)
		}

		if !isVersionAvailable(gvmConfig, requestedVersion) {
			color.Red("❌ Version %s is not available", requestedVersion)
			color.Cyan("  Run 'gvm list update' to refresh available versions")
			os.Exit(1)
		}

		requiredDownloadedVersion := findDownloadedVersion(gvmConfig, requestedVersion)

		if requiredDownloadedVersion == nil {
			color.Yellow("📦 Version not downloaded yet, grabbing it for you...")
			downloadCmd.Run(cmd, []string{requestedVersion})

			gvmConfig, err = internal.LoadConfig()
			if err != nil {
				color.Red("❌ Couldn't reload config: %s", err.Error())
				os.Exit(1)
			}

			requiredDownloadedVersion = getDownloadedVersion(gvmConfig, requestedVersion)
			if requiredDownloadedVersion == nil {
				color.Red("❌ Something went wrong - couldn't find the downloaded version")
				os.Exit(1)
			}
		} else {
			fmt.Println()
			color.Green("✨ Found cached version, installing...")
		}

		fmt.Println()
		color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		color.Cyan("          ⚙️  SWITCHING TO GO %s", strings.ToUpper(requestedVersion))
		color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		faint.Println("  Removing old installation...")

		if err := internal.PurgeCurrentGolangInstallation(); err != nil {
			color.Red("❌ Couldn't remove old version: %s", err.Error())
			os.Exit(1)
		}

		faint.Println("  Extracting new version...")

		if _, err := internal.ExecShellCommand("tar", "-C", "/usr/local", "-xzf", requiredDownloadedVersion.TarPath); err != nil {
			color.Red("❌ Extraction failed: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println()
		color.Green("✅ Switched to Go %s!", requestedVersion)

		pathVars := os.Getenv("PATH")
		pathIncluded := slices.Contains(strings.Split(pathVars, ":"), "/usr/local/go/bin")

		if !pathIncluded {
			color.Yellow("⚠️  Don't forget to update your PATH!")
			fmt.Println()
			color.White("  Add this to your shell config (~/.bashrc, ~/.zshrc, etc.):")
			color.Cyan("  export PATH=$PATH:/usr/local/go/bin")
		} else {
			fmt.Println()
			color.Green("🎉 Ready to code! Run 'go version' to verify")
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
