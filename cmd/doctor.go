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
		fmt.Printf("    root         %s\n", d.Root)
		fmt.Printf("    shims        %s\n", d.ShimDir)
		fmt.Printf("    installed    %d version(s)\n", d.Installed)

		if d.CurrentErr == nil {
			fmt.Printf("    active       %s\n", internal.DisplayVersion(d.Current))
		} else {
			fmt.Printf("    active       none\n")
		}
		if d.ResolvedGo != "" {
			fmt.Printf("    go on PATH   %s (%s)\n", d.ResolvedGo, internal.DisplayVersion(d.ActiveGo))
		} else {
			fmt.Printf("    go on PATH   not found\n")
		}

		fmt.Println()
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
			fmt.Println()
			accent(cyan, "  Suggested fixes:")
			for _, h := range d.Hints {
				plain("    %s", h)
			}
		}
		fmt.Println()

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
			fmt.Println(shims)
			return nil
		}
		fmt.Println(internal.ShellExportLine(shims))
		return nil
	},
}

func init() {
	envCmd.Flags().Bool("path", false, "print only the directory to add to PATH")
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(envCmd)
}
