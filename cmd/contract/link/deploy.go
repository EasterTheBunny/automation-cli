package link

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
	"github.com/easterthebunny/automation-cli/internal/util"
)

func init() {
	deployFeedCmd.Flags().StringVar(&answer, "answer", "2e18", "value returned by the mock LINK-ETH or fast gas feed")
}

var (
	answer   string
	feedType string

	deployTokenCmd = &cobra.Command{
		Use:   "deploy-token",
		Short: "Create new LINK token contract and add to environment",
		Long:  `Create new LINK token contract and add to environment.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			env.LinkToken = &config.LinkTokenContract{
				Mocked: true,
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return err
			}

			deployable := asset.NewLinkTokenDeployable(env.LinkToken)

			if _, err = deployable.Deploy(cmd.Context(), deployer); err != nil {
				return err
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}

	deployFeedCmd = &cobra.Command{
		Use:   "deploy-feed [TYPE]",
		Short: "Create new mock LINK-ETH or fast gas feed contracts",
		Long:  `Create new mock LINK-ETH or fast gas feed contract. The resulting contract always returns the configured amount.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			amount, err := util.ParseExp(answer)
			if err != nil {
				return err
			}

			feed := config.FeedContract{
				Mocked:        true,
				DefaultAnswer: amount.Uint64(),
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return err
			}

			switch args[0] {
			case "link-eth":
				env.LinkETH = &feed

				deployable := asset.NewLinkETHFeedDeployable(env.LinkETH)

				_, err := deployable.Deploy(cmd.Context(), deployer)
				if err != nil {
					return err
				}
			case "fast-gas":
				env.FastGas = &feed

				deployable := asset.NewFastGasFeedDeployable(env.FastGas)

				_, err := deployable.Deploy(cmd.Context(), deployer)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown feed type")
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)

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
