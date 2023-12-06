package participant

import (
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/node"
)

func init() {

}

var (
	removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove a node from a network",
		Long:  `Remove a node from a network`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, env, _, err := prepare(cmd)
			if err != nil {
				return err
			}

			if removeAll {
				for _, nodeConf := range env.Participants {
					if err := node.RemoveParticipantNode(
						cmd.Context(),
						env.Groupname,
						nodeConf,
					); err != nil {
						return err
					}
				}

				if env.Bootstrap != nil {
					if err := node.RemoveParticipantNode(
						cmd.Context(),
						env.Groupname,
						*env.Bootstrap,
					); err != nil {
						return err
					}
				}

				env.Bootstrap = nil
				env.Participants = []config.NodeConfig{}
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
