package registry

import "github.com/spf13/cobra"

//nolint:gochecknoinits
func init() {
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(setCmd)
	RootCmd.AddCommand(setConfigCmd)

	deployCmd.Flags().
		StringVar(&mode, "mode", "DEFAULT", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")

	setConfigCmd.Flags().
		StringVar(&ocrConfigPath, "with-ocr-config", "", "override stored config with values from a json file")

	setConfigCmd.Flags().
		Uint8Var(&maxFaulty, "max-faulty", 1, "set max faulty nodes (default 1)")
}

var (
	mode          string
	ocrConfigPath string
	maxFaulty     uint8

	RootCmd = &cobra.Command{
		Use:   "registry [ACTION]",
		Short: "Create and interact with a registry contract",
		Long:  `Create a registry contract, connect to an existing registry, or run a set-config.`,
		Args:  cobra.MinimumNArgs(1),
	}
)
