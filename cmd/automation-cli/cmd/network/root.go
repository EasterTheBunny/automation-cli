package network

import (
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/network/bootstrap"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/network/participant"
)

func init() {
	RootCmd.AddCommand(participant.RootCmd)
	RootCmd.AddCommand(bootstrap.RootCmd)
	RootCmd.AddCommand(fundCmd)
	RootCmd.AddCommand(listCmd)
}

var RootCmd = &cobra.Command{
	Use:   "network [ACTION]",
	Short: "Manage network components such as a bootstrap node and/or automation nodes",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
}
