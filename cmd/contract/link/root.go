package link

import "github.com/spf13/cobra"

//nolint:gochecknoinits
func init() {
	RootCmd.AddCommand(deployTokenCmd)
	RootCmd.AddCommand(setTokenCmd)
	RootCmd.AddCommand(mintTokenCmd)

	RootCmd.AddCommand(deployFeedCmd)
	RootCmd.AddCommand(setLinkFeedCmd)
	RootCmd.AddCommand(setGasFeedCmd)
}

var RootCmd = &cobra.Command{
	Use:   "link [ACTION]",
	Short: "Manage contracts related to the LINK token.",
	Long:  `Create token contract, LINK-ETH feed. Connect to existing contracts. Send LINK token to addresses.`,
	Args:  cobra.MinimumNArgs(1),
}
