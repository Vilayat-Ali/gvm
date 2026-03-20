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

var downloadCmd = &cobra.Command{
	Use:   "download <version>",
	Short: "download a Go version 📥",
	Long: `Grab a Go version and save it for later use.

Downloaded versions are stored locally so you can switch to them
instantly without re-downloading.

Examples:
  gvm download 1.22.0    → download Go 1.22.0
  gvm download latest    → download the newest version`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !internal.ConfigExists() {
			return fmt.Errorf("configuration not found. Please run 'gvm configure' first")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("❌ Oops! You forgot the version")
			color.Cyan("  Usage: gvm download <version>")
			color.White("  Example: gvm download 1.22.0")
			os.Exit(1)
		}

		faint := color.New(color.Faint)

		requestedVersion := args[0]
		if !internal.ValidateGoVersion(requestedVersion) {
			color.Red("❌ '%s' doesn't look like a valid Go version", requestedVersion)
			faint.Println("  Go versions look like: 1.21.0, 1.22.0, 1.23.0-rc1")
			os.Exit(1)
		}

		gvmConfig, err := internal.LoadConfig()
		if err != nil {
			color.Red("❌ Couldn't load config: %s", err.Error())
			os.Exit(1)
		}

		var remoteVersion *internal.RemoteVersion

		for _, rv := range gvmConfig.AvailableVersions {
			if internal.NormalizeVersion(rv.Version) == strings.TrimSpace(requestedVersion) {
				remoteVersion = &rv
				break
			}
		}

		if remoteVersion == nil {
			color.Red("❌ Version %s not found in available versions", requestedVersion)
			color.Cyan("  Run 'gvm list update' to refresh the list")
			os.Exit(1)
		}

		fmt.Println()
		color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		color.Cyan("          📥 DOWNLOADING GO %s", strings.ToUpper(requestedVersion))
		color.Cyan("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		path, err := remoteVersion.Download()
		if err != nil {
			color.Red("❌ Download failed: %s", err.Error())
			os.Exit(1)
		}

		if err := gvmConfig.MarkVersionAsDownloaded(remoteVersion, *path); err != nil {
			color.Red("❌ Couldn't save to config: %s", err.Error())
			os.Exit(1)
		}

		fmt.Println()
		color.Green("✅ Download complete!")
		color.White("  📍 Location: %s", *path)
		fmt.Println()
		color.Cyan("Next: Run 'gvm use %s' to start using it!", requestedVersion)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
