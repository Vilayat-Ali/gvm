package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check that gvm and your shell are wired up correctly",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		d := internal.Diagnose()

		heading("Environment")
		plain("    root         %s", d.Root)
		plain("    shims        %s", d.ShimDir)
		plain("    installed    %d version(s)", d.Installed)

		if d.CurrentErr == nil {
			plain("    active       %s", internal.DisplayVersion(d.Current))
		} else {
			plain("    active       none")
		}
		if d.ResolvedGo != "" {
			plain("    go on PATH   %s (%s)", d.ResolvedGo, internal.DisplayVersion(d.ActiveGo))
		} else {
			plain("    go on PATH   not found")
		}

		blank()
		for _, problem := range d.Problems {
			accent(yellow, "  ! %s", problem)
		}
		for _, warning := range d.Warnings {
			accent(yellow, "  ! %s", warning)
		}
		if d.Healthy() && len(d.Warnings) == 0 {
			success("everything looks good")
		}
		if len(d.Hints) > 0 {
			blank()
			accent(cyan, "  Suggested fixes:")
			for _, h := range d.Hints {
				plain("    %s", h)
			}
		}
		blank()

		if !d.Healthy() {
			return fmt.Errorf("%d problem(s) found", len(d.Problems))
		}
		return nil
	},
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print the shell line that puts gvm on your PATH",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		shims, err := internal.ShimDir()
		if err != nil {
			return err
		}
		if pathOnly, _ := cmd.Flags().GetBool("path"); pathOnly {
			out("%s", shims)
			return nil
		}
		out("%s", internal.ShellExportLine(shims))
		return nil
	},
}

func init() {
	envCmd.Flags().Bool("path", false, "print only the directory to add to PATH")
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(envCmd)
}
