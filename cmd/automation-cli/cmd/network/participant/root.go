package participant

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(resetCmd)

	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "set the log level for the node")
}

var (
	logLevel string

	RootCmd = &cobra.Command{
		Use:   "participant [ACTION]",
		Short: "Manage network participant nodes.",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}
)
