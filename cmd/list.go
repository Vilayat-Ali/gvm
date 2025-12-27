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

	"github.com/spf13/cobra"
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
		showRemote, _ := cmd.Flags().GetBool("remote")

		if showRemote {
			fmt.Println("Show remote")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Define flags for the list command
	listCmd.Flags().BoolP("all", "a", false, "Show all versions (installed + available)")
	listCmd.Flags().BoolP("remote", "r", false, "Show remote available versions")
	listCmd.Flags().BoolP("installed", "i", true, "Show installed versions only")
	listCmd.Flags().BoolP("current", "c", false, "Show current active version only")
	listCmd.Flags().BoolP("verbose", "v", false, "Verbose output with additional details")
}
