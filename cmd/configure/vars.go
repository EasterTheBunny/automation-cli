package configure

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configSetVarCmd = &cobra.Command{
	Use:   "set [NAME] [VALUE]",
	Short: "Shortcut to quickly update config variables",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("unimplemented")
	},
}

var configGetVarCmd = &cobra.Command{
	Use:   "get [NAME]",
	Short: "Read config variables",
	Long:  `Read config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("unimplemented")
	},
}
