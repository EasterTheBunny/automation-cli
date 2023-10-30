package key

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(storeCmd)
	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(deleteCmd)
	RootCmd.AddCommand(importGanacheCmd)
}

var RootCmd = &cobra.Command{
	Use:   "key [ACTION]",
	Short: "Shortcut to quickly update config var",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.MinimumNArgs(1),
}
