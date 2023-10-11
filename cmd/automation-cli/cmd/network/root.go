package network

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(fundCmd)
}

var RootCmd = &cobra.Command{
	Use:   "network [ACTION]",
	Short: "Manage network components such as a bootstrap node and/or automation nodes",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
}
