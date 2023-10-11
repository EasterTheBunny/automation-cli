package configure

import (
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(configSetVarCmd)
	RootCmd.AddCommand(configGetVarCmd)
	RootCmd.AddCommand(configSetupCmd)
}

var RootCmd = &cobra.Command{
	Use:   "config [ACTION]",
	Short: "Shortcut to quickly update config var",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.MinimumNArgs(1),
}
