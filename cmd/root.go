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

	"github.com/spf13/cobra"
)

const version = "2.0.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gvm",
	Short: "gvm - switch between Go versions like a pro 🧑‍💻",
	Long: `╔══════════════════════════════════════════════════════════════╗
║                    gvm - Go Version Manager                    ║
║          seamlessly switch between Go versions 🔄             ║
╚══════════════════════════════════════════════════════════════╝

GVM is your one-stop shop for managing multiple Go versions without
breaking a sweat. No more manual installs, no more PATH wrestling.

Quick Start:
  gvm configure          → setup gvm for the first time
  gvm list               → see what Go versions are available
  gvm download 1.22.0    → grab a Go version
  gvm use 1.22.0         → start using it like a boss

Features:
  ✨ Download any Go version instantly
  🔀 Switch between versions on the fly
  📋 Keep track of what's installed
  🔒 Verified downloads with checksums
  🚀 Lightweight and blazing fast

Made with 💜 for the Go community
https://github.com/vilayat-ali/gvm`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
			os.Exit(0)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")

	rootCmd.SetVersionTemplate(`gvm v{{.Version}} 🚀
The ultimate Go version manager
Run 'gvm --help' to get started
`)
}
