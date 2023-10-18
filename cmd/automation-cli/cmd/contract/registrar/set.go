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
	setCmd = &cobra.Command{
		Use:   "set-address [ADDRESS]",
		Short: "Set the address and configuration of an existing registrar contract",
		Long:  `Set the address and configuration of an existing registrar contract and add the address and configuration parameters to the environment.`,
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

			registrar := asset.NewRegistrarV21Deployable(&asset.RegistrarV21Config{
				RegistryAddr:  conf.ServiceContract.RegistryAddress,
				LinkTokenAddr: conf.LinkContract,
				MinLink:       0,
			})

			addr, err := registrar.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registrar_address", addr)

			return nil
		},
	}
)
