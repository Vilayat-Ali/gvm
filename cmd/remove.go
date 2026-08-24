package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <version>",
		Aliases: []string{"uninstall", "rm", "delete"},
		Short:   "Remove an installed Go version",
		Args:    cobra.ExactArgs(1),
		Example: "  gvm remove 1.21.0\n  gvm remove 1.21.0 --purge",
		RunE: func(cmd *cobra.Command, args []string) error {
			purge, _ := cmd.Flags().GetBool("purge")

			if _, err := internal.CanonicalVersion(args[0]); err != nil && args[0] != internal.LatestAlias {
				return err
			}
			canonical, err := internal.ResolveInstalled(args[0])
			if err != nil {
				return err
			}
			if err := internal.RemoveVersion(canonical, purge); err != nil {
				return err
			}

			success("removed Go %s", internal.DisplayVersion(canonical))
			if purge {
				hint("cached archive deleted as well")
			}
			return nil
		},
	}
	cmd.Flags().Bool("purge", false, "also delete the cached download archive")
	return cmd
}

func init() {
	rootCmd.AddCommand(newRemoveCmd())
}
