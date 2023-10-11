package network

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodes on the current network configuration",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := context.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		for _, nodeName := range conf.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", nodeName)
		}

		return nil
	},
}
