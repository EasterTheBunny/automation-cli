package configure

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configSetVarCmd = &cobra.Command{
	Use:   "set [NAME] [VALUE]",
	Short: "Shortcut to quickly update config variables",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set(args[0], args[1])

		return nil
	},
}

var configGetVarCmd = &cobra.Command{
	Use:   "get [NAME]",
	Short: "Read config variables",
	Long:  `Read config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val := viper.Get(args[0])

		switch val.(type) {
		case string:
			fmt.Fprintf(cmd.OutOrStdout(), "%s", val)
		case uint, uint8, uint16, uint32, uint64, int, int32, int64:
			fmt.Fprintf(cmd.OutOrStdout(), "%d", val)
		case bool:
			fmt.Fprintf(cmd.OutOrStdout(), "%t", val)
		}

		return nil
	},
}
