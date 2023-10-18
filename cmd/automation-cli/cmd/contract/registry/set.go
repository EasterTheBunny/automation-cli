package registry

import (
	"fmt"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	setCmd = &cobra.Command{
		Use:   "set-address [ADDRESS]",
		Short: "Set the address and configuration of an existing registry contract",
		Long:  `Set the address and configuration of an existing registry contract and add the address and configuration parameters to the environment.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := context.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			dConfig := config.GetDeployerConfig(conf)

			deployer, err := asset.NewDeployer(&dConfig)
			if err != nil {
				return err
			}

			modeVal := config.GetRegistryMode(conf.ServiceContract.Mode)

			if mode != "" {
				modeVal = config.GetRegistryMode(mode)
			}

			registry := asset.NewRegistryV21Deployable(&asset.RegistryV21Config{
				Mode:            modeVal,
				LinkTokenAddr:   conf.LinkContract,
				LinkETHFeedAddr: conf.LinkETHFeed,
				FastGasFeedAddr: conf.FastGasFeed,
			})

			addr, err := registry.Connect(cmd.Context(), args[0], deployer)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registry_address", addr)
			viper.Set("service_contract.registry_version", "v2.1")

			return nil
		},
	}
)
