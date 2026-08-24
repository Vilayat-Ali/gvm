package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vilayat-ali/gvm/internal"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Show installed and available Go versions",
	Args:    cobra.NoArgs,
	Example: "  gvm list\n  gvm list --remote\n  gvm list --current",
	RunE: func(cmd *cobra.Command, args []string) error {
		showRemote, _ := cmd.Flags().GetBool("remote")
		showCurrent, _ := cmd.Flags().GetBool("current")

		switch {
		case showCurrent:
			return printCurrent()
		case showRemote:
			return printRemote()
		default:
			return printInstalled()
		}
	},
}

func printCurrent() error {
	current, err := internal.CurrentVersion()
	if err != nil {
		active := internal.ActiveGoVersion()
		if active == "" {
			warn("no Go version is active")
			hint("run `gvm use latest`")
			return nil
		}
		warn("gvm has no active version; `go` resolves to %s", internal.DisplayVersion(active))
		hint("run `gvm doctor` for details")
		return nil
	}
	out("%s", internal.DisplayVersion(current))
	return nil
}

func printInstalled() error {
	installed, err := internal.InstalledVersions()
	if err != nil {
		return err
	}
	if len(installed) == 0 {
		warn("no Go versions installed")
		hint("run `gvm use latest` to install the newest stable release")
		return nil
	}

	current, _ := internal.CurrentVersion()

	heading("Installed")
	for _, version := range installed {
		if version.Version == current {
			outColor(green, "  * %s", internal.DisplayVersion(version.Version))
			continue
		}
		out("    %s", internal.DisplayVersion(version.Version))
	}
	blank()
	hint("* = active   |   `gvm list --remote` shows downloadable versions")
	return nil
}

func printRemote() error {
	ctx, cancel := signalContext()
	defer cancel()

	config, err := internal.LoadOrInitConfig()
	if err != nil {
		return err
	}
	if config.IsStale() {
		if err := config.Refresh(ctx); err != nil && len(config.AvailableVersions) == 0 {
			return err
		}
	}
	if len(config.AvailableVersions) == 0 {
		warn("no versions in the catalog")
		hint("run `gvm list update`")
		return nil
	}

	current, _ := internal.CurrentVersion()
	latest, _ := config.LatestStable()

	heading("Available")
	shown := 0
	for _, remote := range config.AvailableVersions {
		parsed, err := internal.ParseVersion(remote.Version)
		if err != nil {
			continue
		}
		display := internal.DisplayVersion(remote.Version)

		switch {
		case remote.Version == current:
			outColor(green, "  * %-12s active", display)
		case latest != nil && remote.Version == latest.Version:
			outColor(magenta, "    %-12s latest stable", display)
		case parsed.IsPrerelease():
			outColor(yellow, "    %-12s pre-release", display)
		case internal.IsInstalled(remote.Version):
			out("    %-12s installed", display)
		default:
			out("    %s", display)
		}

		shown++
		if shown >= 20 {
			break
		}
	}

	if remaining := len(config.AvailableVersions) - shown; remaining > 0 {
		blank()
		hint("... and %d older releases", remaining)
	}
	fmt.Println()
	if latest != nil {
		hint("install with `gvm use %s`", internal.DisplayVersion(latest.Version))
	}
	return nil
}

var updateListCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh the catalog of downloadable Go versions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signalContext()
		defer cancel()

		config, err := internal.LoadOrInitConfig()
		if err != nil {
			return err
		}
		if err := config.Refresh(ctx); err != nil {
			return err
		}

		success("catalog updated — %d releases available", len(config.AvailableVersions))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(updateListCmd)
	listCmd.AddCommand(newRemoveCmd())

	listCmd.Flags().BoolP("remote", "r", false, "show versions available to download")
	listCmd.Flags().BoolP("current", "c", false, "print only the active version")
	listCmd.Flags().BoolP("downloaded", "d", false, "show installed versions (default)")
	_ = listCmd.Flags().MarkDeprecated("downloaded", "installed versions are shown by default")
}
