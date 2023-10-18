package registrar

import "github.com/spf13/cobra"

//nolint:gochecknoinits
func init() {
	RootCmd.AddCommand(setCmd)
	RootCmd.AddCommand(deployCmd)
}

var RootCmd = &cobra.Command{
	Use:   "registrar [ACTION]",
	Short: "Create and interact with a registrar contract",
	Long:  `Create a registrar contract, connect to an existing registrar.`,
	Args:  cobra.MinimumNArgs(1),
}
