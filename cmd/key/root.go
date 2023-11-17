package key

import "github.com/spf13/cobra"

func init() {
	RootCmd.AddCommand(storeCmd)
	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(deleteCmd)
	RootCmd.AddCommand(importGanacheCmd)
	RootCmd.AddCommand(importGethCmd)

	storeCmd.Flags().BoolVar(&readStdIn, "stdin", false, "read value from standard input instead of prompting")
	importGethCmd.Flags().StringVar(&passwordPath, "password", "", "password file path to unlock private keys")
}

var (
	readStdIn    bool
	passwordPath string

	RootCmd = &cobra.Command{
		Use:   "key [ACTION]",
		Short: "Shortcut to quickly update config var",
		Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
		Args:  cobra.MinimumNArgs(1),
	}
)
