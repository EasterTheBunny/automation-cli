package participant

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

func init() {

}

var (
	removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a node from a network",
		Long:  `Remove a node from a network`,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := context.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			if removeAll {
				/*
					for nodeID, nodeName := range conf.Nodes {
						if err := node.RemoveParticipantNode(
							cmd.Context(),
							uint16(6688+nodeID),
							conf.Groupname,
							nodeName,
							"not-postgres-image",
						); err != nil {
							return err
						}
					}
				*/

				conf.Nodes = []string{}

				viper.Set("nodes", conf.Nodes)
			}

			return nil
		},
	}
)
