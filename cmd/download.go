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
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/vilayat-ali/gvm/internal"
)

// downloadCmd represents the download command
var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
			if rv.Version == strings.TrimSpace(requestedVersion) {
				remoteVersion = &rv
				remoteVersionIdx = idx
			}
		}

		if remoteVersion == nil || remoteVersionIdx == -1 {
			color.Red(fmt.Sprintf("Download Error: Failed to download '%s'. Couldn't find in config available versions", requestedVersion))
			os.Exit(1)
		}

		color.Green(fmt.Sprintf("Downloading %s\n", requestedVersion))
		path, err := remoteVersion.Download(PrintProgress())
		if err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}

		if err := config.MarkVersionAsDownloaded(uint(remoteVersionIdx), *path); err != nil {
			color.Red(err.Error())
			os.Exit(1)
		}
	},
}

func PrintProgress() internal.ProgressFunc {
	start := time.Now()
	last := time.Now()

	return func(downloaded, total int64) {
		if time.Since(last) < 200*time.Millisecond {
			return
		}
		last = time.Now()

		elapsed := time.Since(start).Seconds()
		speed := float64(downloaded) / elapsed / 1024 / 1024

		if total > 0 {
			percent := float64(downloaded) / float64(total) * 100
			color.Green(fmt.Sprintf(
				"\r%.2f%% | %.2f / %.2f MB | %.2f MB/s",
				percent,
				float64(downloaded)/1024/1024,
				float64(total)/1024/1024,
				speed,
			))
		} else {
			color.Green(fmt.Sprintf(
				"\rDownloaded %.2f MB | %.2f MB/s",
				float64(downloaded)/1024/1024,
				speed,
			))
		}
	}
}

func init() {
	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ltsVersion, err := config.GetLTSVersion()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	downloadCmd.Flags().AddFlag(&pflag.Flag{
		Name:      "version",
		Shorthand: "g",
		Usage:     "gvm download --version 1.25.5",
		DefValue:  *ltsVersion,
	})
	rootCmd.AddCommand(downloadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downloadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downloadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
