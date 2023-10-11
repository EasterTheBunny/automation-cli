package interact

import (
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/domain"
)

func init() {
	RootCmd.AddCommand(registryCmd)
	RootCmd.AddCommand(loadCmd)
}

var RootCmd = &cobra.Command{
	Use:       "interact [NAME] [ACTION]",
	Short:     "Run pre-defined actions for contract",
	Long:      `Interact with contracts and run pre-packaged actions. This is not inclusive of all commands possible to run`,
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: domain.ContractNames,
}
