package registrar

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new registrar contract",
		Long:  `Deploy a new registrar contract and add the address and configuration parameters to the environment.`,
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

			deployable := asset.NewRegistrarV21Deployable(&asset.RegistrarV21Config{
				RegistryAddr:  conf.ServiceContract.RegistryAddress,
				LinkTokenAddr: conf.LinkContract,
				MinLink:       0,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registrar_address", addr)

			return nil
		},
	}
)
