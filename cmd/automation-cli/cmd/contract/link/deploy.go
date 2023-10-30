package link

import (
	"fmt"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	deployFeedCmd.Flags().StringVar(&answer, "answer", "2e18", "value returned by the mock LINK-ETH feed")
}

var (
	answer   string
	feedType string

	deployTokenCmd = &cobra.Command{
		Use:   "deploy-token",
		Short: "Create new LINK token contract and add to environment",
		Long:  `Create new LINK token contract and add to environment.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf := context.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			dConfig := config.GetDeployerConfig(conf)

			keyConf := context.GetKeyConfigFromContext(cmd.Context())
			if keyConf == nil {
				return fmt.Errorf("missing private key config")
			}

			pkOverride, err := cmd.Flags().GetString("key")
			if err != nil {
				return err
			}

			dConfig = config.SetPrivateKey(dConfig, keyConf, pkOverride)

			deployer, err := asset.NewDeployer(&dConfig)
			if err != nil {
				return err
			}

			deployable := asset.NewLinkTokenDeployable()

			addr, err := deployable.Deploy(cmd.Context(), deployer)
			if err != nil {
				return err
			}

			viper.Set("link_contract_address", addr)

			return nil
		},
	}

	deployFeedCmd = &cobra.Command{
		Use:   "deploy-feed [TYPE]",
		Short: "Create new mock LINK-ETH or fast gas feed contracts",
		Long:  `Create new mock LINK-ETH or fast gas feed contract. The resulting contract always returns the configured amount.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := context.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			dConfig := config.GetDeployerConfig(conf)

			keyConf := context.GetKeyConfigFromContext(cmd.Context())
			if keyConf == nil {
				return fmt.Errorf("missing private key config")
			}

			pkOverride, err := cmd.Flags().GetString("key")
			if err != nil {
				return err
			}

			dConfig = config.SetPrivateKey(dConfig, keyConf, pkOverride)

			deployer, err := asset.NewDeployer(&dConfig)
			if err != nil {
				return err
			}

			amount, err := util.ParseExp(answer)
			if err != nil {
				return err
			}

			switch args[0] {
			case "link-eth":
				deployable := asset.NewLinkETHFeedDeployable(&asset.LinkETHFeedConfig{
					Answer: amount,
				})

				addr, err := deployable.Deploy(cmd.Context(), deployer)
				if err != nil {
					return err
				}

				viper.Set("link_eth_feed", addr)
			case "fast-gas":
				deployable := asset.NewFastGasFeedDeployable(&asset.FeedConfig{
					Answer: amount,
				})

				addr, err := deployable.Deploy(cmd.Context(), deployer)
				if err != nil {
					return err
				}

				viper.Set("fast_gas_feed", addr)
			default:
				return fmt.Errorf("unknown feed type")
			}

			return nil
		},
	}
)
