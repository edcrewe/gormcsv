package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is set at link time to `git describe --tags --always --dirty`, e.g. v0.3.1 on a
// tag or v0.3.1-2-ga6866f6 two commits after it. See the Makefile build target.
var Version = ""

// resolveVersion prefers the linked version. Otherwise it uses the module version Go embeds:
// the tag for `go install ...@v0.3.1`, or a pseudo-version containing the commit hash for a
// plain `go build` in a checkout.
func resolveVersion(linked string, info *debug.BuildInfo, ok bool) string {
	if linked != "" {
		return linked
	}
	if !ok || info == nil {
		return "devel"
	}
	version := info.Main.Version
	if version == "" || version == "(devel)" {
		version = "devel"
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && len(setting.Value) >= 7 {
				version += "-" + setting.Value[:7]
			}
		}
	}
	return version
}

func currentVersion() string {
	info, ok := debug.ReadBuildInfo()
	return resolveVersion(Version, info, ok)
}

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print the gormcsv version",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "gormcsv %s\n", currentVersion()); err != nil {
			return fmt.Errorf("write version: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCommand)
	rootCmd.Version = currentVersion()
	rootCmd.SetVersionTemplate("gormcsv {{.Version}}\n")
}
