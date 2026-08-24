package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var useCmd = &cobra.Command{
	Use:     "use <version>",
	Short:   "Activate a Go version, installing it first if needed",
	Args:    cobra.ExactArgs(1),
	Example: "  gvm use 1.22.0\n  gvm use latest",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signalContext()
		defer cancel()

		config, err := internal.LoadOrInitConfig()
		if err != nil {
			return err
		}

		remote, err := config.ResolveVersion(ctx, args[0])
		if err != nil {
			return err
		}
		display := internal.DisplayVersion(remote.Version)

		if current, err := internal.CurrentVersion(); err == nil && current == remote.Version {
			success("Go %s is already active", display)
			return nil
		}

		if !internal.IsInstalled(remote.Version) {
			heading("Installing Go " + display)
			hint("source: %s", remote.DownloadLink)
			if _, err := internal.EnsureInstalled(ctx, remote, newProgress(remote.Size, "  downloading")); err != nil {
				return err
			}
			fmt.Println()
		}

		if err := internal.Activate(remote.Version); err != nil {
			return err
		}

		success("switched to Go %s", display)

		diagnosis := internal.Diagnose()
		if !diagnosis.OnPath {
			fmt.Println()
			warn("%s is not on your PATH yet", diagnosis.ShimDir)
			accent(bold, "    %s", internal.ShellExportLine(diagnosis.ShimDir))
			return nil
		}
		if diagnosis.Shadowed {
			fmt.Println()
			warn("another Go in %s still takes priority — run `gvm doctor`", diagnosis.ShadowedBy)
			return nil
		}
		hint("run `go version` to confirm")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
