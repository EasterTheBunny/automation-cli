package contract

import (
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/contract/link"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/contract/load"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/contract/registrar"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/contract/registry"
)

func init() {
	RootCmd.AddCommand(link.RootCmd)
	RootCmd.AddCommand(load.RootCmd)
	RootCmd.AddCommand(registrar.RootCmd)
	RootCmd.AddCommand(registry.RootCmd)
}

var RootCmd = &cobra.Command{
	Use:   "contract [ASSET]",
	Short: "Manage contracts associated with the current network.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
}
