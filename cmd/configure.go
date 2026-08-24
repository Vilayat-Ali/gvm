package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Prepare the gvm directory and fetch the version catalog",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		if internal.ConfigExists() && !force {
			root, err := internal.Root()
			if err != nil {
				return err
			}
			success("gvm is already configured")
			hint("root: %s", root)
			hint("run `gvm doctor` to verify your shell setup")
			return nil
		}

		ctx, cancel := signalContext()
		defer cancel()

		if err := internal.EnsureLayout(); err != nil {
			return err
		}

		config, err := internal.NewConfig()
		if err != nil {
			return err
		}
		if err := config.Save(); err != nil {
			return err
		}

		heading("Configuring gvm")
		hint("fetching the Go release catalog...")
		if err := config.Refresh(ctx); err != nil {
			return err
		}

		root, err := internal.Root()
		if err != nil {
			return err
		}
		shims, err := internal.ShimDir()
		if err != nil {
			return err
		}

		fmt.Println()
		success("ready — %d releases available for your platform", len(config.AvailableVersions))
		hint("toolchains live in %s", root)
		fmt.Println()
		accent(cyan, "  Add this to your shell profile:")
		accent(bold, "    %s", internal.ShellExportLine(shims))
		fmt.Println()
		hint("then run `gvm use latest`")
		return nil
	},
}

func init() {
	configureCmd.Flags().Bool("force", false, "rebuild the configuration even if it already exists")
	rootCmd.AddCommand(configureCmd)
}
