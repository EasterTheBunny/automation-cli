package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/node"
)

var (
	setCmd = &cobra.Command{
		Use:   "set [IMAGE]",
		Short: "Set a bootstrap node for a network",
		Long:  `Set a bootstrap node for a network`,
		Example: `The following will create a bootstrap node and save the interaction details to the provided
environment named 'non.default'.

$ automation-cli network bootstrap set chainlink:latest --environment="non.default"`,
		ValidArgs: []string{"bootstrap", "participant"},
		Args:      cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := context.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			paths := context.GetPathsFromContext(cmd.Context())
			if paths == nil {
				return fmt.Errorf("missing config path in context")
			}

			nodeConfigPath := fmt.Sprintf("%s/%s", paths.Environment, "bootstrap")
			str, err := node.CreateBootstrapNode(cmd.Context(), node.NodeConfig{
				ChainID:          conf.ChainID,
				NodeWSSURL:       conf.RPCWSSURL,
				NodeHttpURL:      conf.RPCHTTPURL,
				LogLevel:         logLevel,
				MercuryLegacyURL: "https://chain2.old.link",
				MercuryURL:       "https://chain2.link",
				MercuryID:        "username2",
				MercuryKey:       "password2",
			}, conf.Groupname, args[1], conf.ServiceContract.RegistryAddress, 5688, 8000, nodeConfigPath, true)
			if err != nil {
				return err
			}

			viper.Set("bootstrap_address", str)

			return nil
		},
	}
)
