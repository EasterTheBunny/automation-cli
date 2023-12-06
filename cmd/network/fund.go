package network

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
	"github.com/easterthebunny/automation-cli/internal/util"
)

var fundCmd = &cobra.Command{
	Use:   "fund [NODE] [AMOUNT]",
	Short: "Transfer funds to node address.",
	Long:  `Transfer funds from the default account to configured node address. Provide either the node name or the index number for the node.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, env, key, err := prepare(cmd)
		if err != nil {
			return err
		}

		deployer, err := asset.NewDeployer(&env, key)
		if err != nil {
			return err
		}

		var nodeConf config.NodeConfig

		for i, node := range env.Participants {
			if node.Name == args[0] || strconv.FormatInt(int64(i), 10) == args[0] {
				nodeConf = env.Participants[i]
			}
		}

		if nodeConf.Name == "" {
			return fmt.Errorf("node not available")
		}

		if nodeConf.Address == "" {
			return fmt.Errorf("node address not available")
		}

		amount, err := util.ParseExp(args[1])
		if err != nil {
			return err
		}

		return deployer.Send(cmd.Context(), nodeConf.Address, amount)
	},
}

func prepare(cmd *cobra.Command) (io.Environment, config.Environment, config.Key, error) {
	var (
		env config.Environment
		key config.Key
		err error
	)

	path := io.EnvironmentFromContext(cmd.Context())
	if path == nil {
		return io.Environment{}, env, key, fmt.Errorf("environment not found")
	}

	env, err = config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
	if err != nil {
		return io.Environment{}, env, key, err
	}

	keys, err := config.ReadPrivateKeysFrom(path.Root.MustRead(config.PrivateKeyConfigFilename))
	if err != nil {
		return io.Environment{}, env, key, err
	}

	pkOverride, err := cmd.Flags().GetString("key")
	if err != nil {
		return io.Environment{}, env, key, err
	}

	if pkOverride == "" {
		pkOverride = env.PrivateKeyAlias
	}

	key, err = keys.KeyForAlias(pkOverride)
	if err != nil {
		return io.Environment{}, env, key, err
	}

	return *path, env, key, nil
}
