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

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a Go version",
	Long: `Download a specific version of Go.

Examples:
  gvm download --version 1.25.5
  gvm download -g 1.25.5`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !internal.ConfigExists() {
			return fmt.Errorf("configuration not found. Please run 'gvm configure' first")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		config, err := internal.LoadConfig()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		requestedVersion, err := cmd.Flags().GetString("version")
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		var remoteVersion *internal.RemoteVersion
		var remoteVersionIdx int = -1

		for idx, rv := range config.AvailableVersions {
			if strings.Replace(rv.Version, "go", "", 1) == strings.TrimSpace(requestedVersion) {
				remoteVersion = &rv
				remoteVersionIdx = idx
			}
		}

		if remoteVersion == nil || remoteVersionIdx == -1 {
			color.Red(fmt.Sprintf("Download Error: Failed to download '%s'. Couldn't find in config available versions", requestedVersion))
			os.Exit(1)
		}

		color.Green(fmt.Sprintf("Downloading %s\n", requestedVersion))
		path, err := remoteVersion.Download()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		if err := config.MarkVersionAsDownloaded(uint(remoteVersionIdx), *path); err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		color.Green(fmt.Sprintf("\nGo version %s was downloaded and saved in %s", remoteVersion.Version, *path))
	},
}

func init() {
	downloadCmd.Flags().StringP("version", "g", "", "Go version to download (e.g., 1.25.5)")
	rootCmd.AddCommand(downloadCmd)
}
