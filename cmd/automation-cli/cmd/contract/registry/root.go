package registry

import "github.com/spf13/cobra"

//nolint:gochecknoinits
func init() {
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(setCmd)
	RootCmd.AddCommand(setConfigCmd)

	RootCmd.PersistentFlags().
		StringVar(&mode, "mode", "", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")
}

var (
	mode string

	RootCmd = &cobra.Command{
		Use:   "registry [ACTION]",
		Short: "Create and interact with a registry contract",
		Long:  `Create a registry contract, connect to an existing registry, or run a set-config.`,
		Args:  cobra.MinimumNArgs(1),
	}
)
