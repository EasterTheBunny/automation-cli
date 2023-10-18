package registry

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/domain"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

func init() {
	_ = deployCmd.Flags().
		String("mode", "DEFAULT", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")
}

var (
	deployCmd = &cobra.Command{
		Use:       "deploy",
		Short:     "Deploy a new registry contract",
		Long:      `Deploy a new registry contract and add the address and configuration parameters to the environment.`,
		ValidArgs: domain.ContractNames,
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

			modeVal := config.GetRegistryMode(conf.ServiceContract.Mode)

			if mode != "" {
				modeVal = config.GetRegistryMode(mode)
			}

			deployable := asset.NewRegistryV21Deployable(&asset.RegistryV21Config{
				Mode:            modeVal,
				LinkTokenAddr:   conf.LinkContract,
				LinkETHFeedAddr: conf.LinkETHFeed,
				FastGasFeedAddr: conf.FastGasFeed,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registry_address", addr)
			viper.Set("service_contract.registry_version", "v2.1")

			return nil
		},
	}
)
