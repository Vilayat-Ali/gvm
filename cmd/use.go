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

// useCmd represents the use command
var useCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Select a Go version for use on this machine",
	Long: `Select a Go toolchain version to be used across the system or in the current shell.

This command will ensure the requested Go version is installed (if supported),
configure the environment, and set it as the active Go version.

Examples:
  mycli use 1.22.2
  mycli use latest
  mycli use 1.20`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("Arg Error: Expected positional arguement 'golang version'. Example gvm download 1.25.5")
			os.Exit(1)
		}

		requestedVersion := args[0]
		if !internal.ValidateGoVersion(requestedVersion) {
			color.Red(fmt.Sprintf("Input Error: Version '%s' is not a valid golang version", requestedVersion))
			os.Exit(1)
		}

		gvmConfig, err := internal.LoadConfig()
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		if !internal.ValidateGoVersion(requestedVersion) {
			color.Red("Input Error: Invalid golang version requested")
			os.Exit(1)
		}

		var isAvailable bool = false
		var requiredDownloadedVersion *internal.DownloadVersion

		for _, downloadedVersion := range gvmConfig.DownloadedVersions {
			raw_version := strings.Replace(downloadedVersion.Version, "go", "", 1)

			if requestedVersion == raw_version {
				requiredDownloadedVersion = &downloadedVersion
				break
			}
		}

		for _, availableVersion := range gvmConfig.AvailableVersions {
			raw_version := strings.Replace(availableVersion.Version, "go", "", 1)

			if requestedVersion == raw_version {
				isAvailable = true
				break
			}
		}

		if !isAvailable {
			color.Red("Input Error: Invalid version %s. Version not available for download.", requestedVersion)
			color.Blue("Run `gvm list update` to update the version list available")
			os.Exit(1)
		}

		if !isAvailable && requiredDownloadedVersion == nil {
			color.Red(fmt.Sprintf("Input Error: Invalid version %s was asked to be used. Version neither available nor downloaded.", requestedVersion))
			os.Exit(1)
		}

		if requiredDownloadedVersion == nil {
			color.Yellow(fmt.Sprintf("Version %s not downloaded yet. Downloading now...", requestedVersion))
			downloadCmd.Run(cmd, []string{requestedVersion})

			// downloading the golang version and updating config for evaluation
			gvmConfig, err = internal.LoadConfig()
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}

			// get the DownloadedVersion instance from the newly updated config
			for _, downloadedVersion := range gvmConfig.DownloadedVersions {
				raw_version := strings.Replace(downloadedVersion.Version, "go", "", 1)

				if requestedVersion == raw_version {
					requiredDownloadedVersion = &downloadedVersion
					break
				}
			}
		} else {
			color.Green(fmt.Sprintf("Version %s is already downloaded", requestedVersion))
		}

		// Setup guide:
		// https://go.dev/doc/install

		// delete current golang installation
		internal.PurgeCurrentGolangInstallation()

		// decompress tarball
		if _, err := internal.ExecShellCommand(fmt.Sprintf("tar -C /usr/local -xzf %s", requiredDownloadedVersion.TarPath)); err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		// set the path
		if _, err := internal.ExecShellCommand("export PATH=$PATH:/usr/local/go/bin"); err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		color.Green(fmt.Sprintf("Now using go version %s. Run go version to confirm", requestedVersion))
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
