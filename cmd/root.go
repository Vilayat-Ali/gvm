/*
Copyright © 2026 Syed Vilayat Ali Rizvi

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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var rootCmd = &cobra.Command{
	Use:   "gvm",
	Short: "Manage multiple Go toolchains",
	Long: `gvm installs Go toolchains into your own home directory and switches
between them by moving a single symlink. It never modifies system
directories such as /usr/local/go, and every download is verified
against the checksum published by go.dev.

Quick start:
  gvm configure       prepare the gvm directory and version catalog
  gvm use latest      install and activate the newest stable Go
  gvm list            show what is installed
  gvm doctor          check that your shell is wired up correctly`,
	Version:       internal.AppVersion,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if disable, _ := cmd.Flags().GetBool("no-color"); disable {
			color.NoColor = true
		}
		if err := internal.GuardRoot(); err != nil {
			return err
		}
		if warning := internal.RootWarning(); warning != "" {
			warn("%s", warning)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "gvm: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("gvm {{.Version}}\n")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable coloured output")
}
