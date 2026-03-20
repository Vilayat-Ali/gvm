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
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "see what Go versions are available 🔍",
	Long: `Browse through available and installed Go versions.

Examples:
  gvm list           → see available versions
  gvm list -d        → your downloaded versions
  gvm list -c        → current active version
  gvm list update    → refresh the version list`,
	Run: func(cmd *cobra.Command, args []string) {
		showDownloaded, _ := cmd.Flags().GetBool("downloaded")
		showCurrent, _ := cmd.Flags().GetBool("current")

		if showDownloaded {
			showDownloadedVersions()
			return
		}

		if showCurrent {
			showCurrentVersion()
			return
		}

		showAvailableVersions()
	},
}

func isCurrentVersion(configVersion string, currentVersion *string) bool {
	if currentVersion == nil {
		return false
	}
	normalizedConfig := internal.NormalizeVersion(configVersion)
	normalizedCurrent := internal.NormalizeVersion(*currentVersion)
	return normalizedConfig == normalizedCurrent
}

func isSystemGoInstalled() bool {
	goBinPath := path.Join("/usr", "local", "go", "bin", "go")
	if _, err := os.Stat(goBinPath); err == nil {
		return true
	}
	return false
}

func showDownloadedVersions() {
	config, err := internal.LoadConfig()
	if err != nil {
		color.Red("❌ Oops! Couldn't load config: %s", err.Error())
		os.Exit(1)
	}

	currentVersion, _ := internal.GetCurrentGolangVersion()
	systemInstalled := isSystemGoInstalled()

	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Cyan("          📦 INSTALLED GO VERSIONS")
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	hasVersions := false
	ltsFound := false

	for _, downloadVersion := range *config.GetDownloadedVersions() {
		hasVersions = true
		version := downloadVersion.Version
		isRC := strings.Contains(version, "rc")
		isCurrent := isCurrentVersion(version, currentVersion)

		isSystemVersion := systemInstalled && currentVersion != nil &&
			strings.TrimPrefix(version, "go") == *currentVersion

		var marker string
		if isSystemVersion {
			marker = " ▶  ACTIVE (GVM)"
		} else if isCurrent {
			marker = " ▶  ACTIVE"
		} else if !ltsFound && !isRC {
			marker = " ✨ RECOMMENDED"
			ltsFound = true
		} else {
			marker = " •  INSTALLED"
		}

		if isSystemVersion || isCurrent {
			color.Green("  %s%s", version, marker)
		} else if strings.Contains(marker, "RECOMMENDED") {
			color.Magenta("  %s%s", version, marker)
		} else if isRC {
			color.Yellow("  %s%s", version, marker)
		} else {
			color.White("  %s%s", version, marker)
		}
	}

	if !hasVersions && !systemInstalled {
		color.Yellow("📦 No Go versions installed yet")
		color.Cyan("  Run 'gvm download 1.22.0' to get started!")
		fmt.Println()
		return
	}

	if !hasVersions && systemInstalled && currentVersion != nil {
		color.Green("  %s ▶  ACTIVE (SYSTEM)", *currentVersion)
		fmt.Println()
		color.HiBlack("Legend: ▶ = Currently using | ✨ = Recommended | • = Installed | GVM = Managed by gvm | SYSTEM = Pre-installed")
		return
	}

	fmt.Println()
	color.HiBlack("Legend: ▶ = Currently using | ✨ = Recommended | • = Installed | GVM = Managed by gvm | SYSTEM = Pre-installed")
}

func showCurrentVersion() {
	faint := color.New(color.Faint)
	currentVersion, err := internal.GetCurrentGolangVersion()
	if err != nil {
		color.Yellow("🤔 No Go version detected")
		color.Cyan("  Run 'gvm use <version>' to set one up!")
		return
	}

	systemInstalled := isSystemGoInstalled()

	fmt.Println()
	color.Green("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Green("          ⚡ CURRENT GO VERSION")
	color.Green("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.New(color.FgGreen, color.Bold).Printf("\n     %s\n", *currentVersion)
	if systemInstalled {
		faint.Printf("     (System installation)")
	}
	fmt.Println()
}

func showAvailableVersions() {
	config, err := internal.LoadConfig()
	if err != nil {
		color.Red("❌ Oops! Couldn't load config: %s", err.Error())
		os.Exit(1)
	}

	if len(config.AvailableVersions) == 0 {
		fmt.Println()
		color.Yellow("📦 No versions in cache")
		color.Cyan("  Run 'gvm list update' to fetch the latest!")
		return
	}

	currentVersion, _ := internal.GetCurrentGolangVersion()

	fmt.Println()
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	color.Cyan("          📋 AVAILABLE GO VERSIONS")
	color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	ltsFound := false
	versionCount := 0

	for _, remoteVersion := range config.AvailableVersions {
		version := remoteVersion.Version
		isReleaseCandidate := strings.Contains(version, "rc")
		isCurrent := isCurrentVersion(version, currentVersion)

		var marker string
		if isCurrent {
			marker = " ▶  IN USE"
		} else if !ltsFound && !isReleaseCandidate {
			marker = " ✨ LTS"
			ltsFound = true
		} else if isReleaseCandidate {
			marker = " 🧪 RC"
		} else {
			marker = " •  AVAILABLE"
		}

		if isCurrent {
			color.Green("  %s%s", version, marker)
		} else if strings.Contains(marker, "LTS") {
			color.Magenta("  %s%s", version, marker)
		} else if isReleaseCandidate {
			color.Yellow("  %s%s", version, marker)
		} else {
			color.White("  %s%s", version, marker)
		}

		versionCount++
		if versionCount >= 10 {
			if len(config.AvailableVersions) > 10 {
				color.HiBlack("\n  ... and %d more versions", len(config.AvailableVersions)-10)
			}
			break
		}
	}

	fmt.Println()
	color.Cyan("Quick actions:")
	color.White("  gvm download %s → grab the LTS version", config.AvailableVersions[0].Version)
	fmt.Println()
}

var updateListCmd = &cobra.Command{
	Use:   "update",
	Short: "refresh the version list 🔄",
	Long: `Fetches the latest Go versions from the official Go repository.

Run this periodically to see newly released Go versions!`,
	Run: func(cmd *cobra.Command, args []string) {
		faint := color.New(color.Faint)

		fmt.Println()
		color.Cyan("🔄 Fetching latest Go versions from the matrix...")
		fmt.Println()

		config, err := internal.LoadConfig()
		if err != nil {
			color.Red("❌ Couldn't load config: %s", err.Error())
			os.Exit(1)
		}

		faint.Printf("  Connecting to go.dev...")

		if err := config.UpdateAvailableVersions(); err != nil {
			color.Red("❌ Fetch failed: %s", err.Error())
			os.Exit(1)
		}

		color.Green("✅ Got the latest versions!")
		color.Cyan("  Found %d Go versions ready for download", len(config.AvailableVersions))
		fmt.Println()
		color.White("  Run 'gvm list' to see what's new! 🎉")
		fmt.Println()
	},
}

var deleteVersionFromListCmd = &cobra.Command{
	Use:   "delete <version>",
	Short: "nuke a downloaded version 🗑️",
	Long: `Remove a downloaded Go version from your system.

This deletes the tarball but keeps the version info in case you want to re-download later.

Examples:
  gvm delete 1.21.0    → remove Go 1.21.0`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("❌ Missing version! Usage: gvm delete <version>")
			color.Cyan("  Example: gvm delete 1.22.0")
			os.Exit(1)
		}

		versionToBeDeleted := args[0]
		if !internal.ValidateGoVersion(versionToBeDeleted) {
			color.Red("❌ '%s' doesn't look like a valid Go version", versionToBeDeleted)
			os.Exit(1)
		}

		config, err := internal.LoadConfig()
		if err != nil {
			color.Red("❌ Couldn't load config: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println()
		color.Yellow("🗑️  Removing Go %s...", versionToBeDeleted)

		if err := config.RemoveDownloadedVersion(versionToBeDeleted); err != nil {
			color.Red("❌ Delete failed: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println()
		color.Green("✅ Poof! Go %s has been removed", versionToBeDeleted)
		color.Cyan("  Run 'gvm list -d' to see your remaining versions")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(updateListCmd)
	listCmd.AddCommand(deleteVersionFromListCmd)

	listCmd.Flags().BoolP("downloaded", "d", false, "show downloaded versions")
	listCmd.Flags().BoolP("current", "c", false, "show current version")
}
