package network

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/util"
)

var fundCmd = &cobra.Command{
	Use:   "fund [NODE] [AMOUNT]",
	Short: "Transfer funds to node address.",
	Long:  `Transfer funds from the default account to configured node address.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := context.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		paths := context.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		keyConf := context.GetKeyConfigFromContext(cmd.Context())
		if keyConf == nil {
			return fmt.Errorf("missing private key config")
		}

		for _, key := range keyConf.Keys {
			if key.Alias == dConfig.PrivateKey {
				dConfig.PrivateKey = key.Value

				break
			}
		}

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		nodeName := ""

		for _, n := range conf.Nodes {
			if n == args[0] {
				nodeName = n
			}
		}

		if nodeName == "" {
			return fmt.Errorf("node not available")
		}

		nConf, _, err := config.GetNodeConfig(fmt.Sprintf("%s/%s", paths.Environment, nodeName))
		if err != nil {
			return err
		}

		if nConf.Address == "" {
			return fmt.Errorf("node address not available")
		}

		addr := nConf.Address

		amount, err := util.ParseExp(args[1])
		if err != nil {
			return err
		}

		return deployer.Send(cmd.Context(), addr, amount)
	},
}
