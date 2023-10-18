package bootstrap

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(setCmd)

	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "set the log level for the node")
}

var (
	logLevel string

	RootCmd = &cobra.Command{
		Use:   "bootstrap [ACTION]",
		Short: "Manage network bootstrap nodes.",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}
)
