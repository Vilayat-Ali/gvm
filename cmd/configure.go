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

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "setup gvm for the first time ✨",
	Long: `╔══════════════════════════════════════════════════════════════╗
║                      First time? Let's fix that!                  ║
╚══════════════════════════════════════════════════════════════╝

This wizard sets up everything gvm needs:
  📁 Config: ~/.config/gvm/config.json
  📦 Versions: /usr/local/gvm/go-versions/

Just run 'gvm configure' and we'll handle the rest.

Pro tip: You only need to run this once!`,
	Run: func(cmd *cobra.Command, args []string) {
		if internal.ConfigExists() {
			color.Green("✅ gvm is already configured!")
			color.Cyan("   Run 'gvm --help' to see what you can do")
			return
		}

		fmt.Println()
		color.Cyan("🔧 Setting up gvm for the first time...")
		fmt.Println()

		if err := internal.SetupConfig(); err != nil {
			color.Red("❌ Setup failed: %s", err.Error())
			os.Exit(1)
		}

		color.Green("✨ Boom! gvm is ready to roll!")
		fmt.Println()
		color.Yellow("Next steps:")
		color.White("  1. %s gvm list%s - check available Go versions", color.BoldString("run"), color.ResetString(""))
		color.White("  2. %s gvm download 1.22.0%s - grab a version", color.BoldString("run"), color.ResetString(""))
		color.White("  3. %s gvm use 1.22.0%s - start coding!", color.BoldString("run"), color.ResetString(""))
		fmt.Println()
		color.Green("Happy coding! 🎉")
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
