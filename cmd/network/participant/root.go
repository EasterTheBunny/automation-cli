package participant

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(resetCmd)
	RootCmd.AddCommand(removeCmd)

	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "set the log level for the node")

	removeCmd.Flags().BoolVar(&removeAll, "all", false, "remove all participants")
}

var (
	logLevel  string
	removeAll bool

	RootCmd = &cobra.Command{
		Use:   "participant [ACTION]",
		Short: "Manage network participant nodes.",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
	}
)
