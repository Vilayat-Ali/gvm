package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var downloadCmd = &cobra.Command{
	Use:     "download <version>",
	Aliases: []string{"install"},
	Short:   "Download and unpack a Go version without activating it",
	Args:    cobra.ExactArgs(1),
	Example: "  gvm download 1.22.0\n  gvm download latest",
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

		if internal.IsInstalled(remote.Version) {
			success("Go %s is already installed", internal.DisplayVersion(remote.Version))
			hint("run `gvm use %s` to activate it", internal.DisplayVersion(remote.Version))
			return nil
		}

		heading("Installing Go " + internal.DisplayVersion(remote.Version))
		hint("source: %s", remote.DownloadLink)

		dir, err := internal.EnsureInstalled(ctx, remote, newProgress(remote.Size, "  downloading"))
		if err != nil {
			return err
		}

		blank()
		success("checksum verified and unpacked")
		hint("location: %s", dir)
		hint("run `gvm use %s` to activate it", internal.DisplayVersion(remote.Version))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
