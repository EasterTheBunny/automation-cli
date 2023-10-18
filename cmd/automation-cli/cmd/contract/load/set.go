package load

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/domain"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

var (
	setCmd = &cobra.Command{
		Use:   "set-address [ADDRESS]",
		Short: "Set address for existing verifiable-load contract",
		Long:  `Set address for existing verifiable-load contract.`,
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

			switch upkeepType {
			case domain.VerifiableLoadLogTrigger:
				if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
					return domain.ErrRegistrarNotAvailable
				}

				deployable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
					RegistrarAddr: conf.ServiceContract.RegistrarAddress,
					UseMercury:    conf.LogTriggerLoadContract.UseMercury,
					UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
				})

				addr, err := deployable.Connect(cmd.Context(), args[0], deployer)
				if err != nil {
					return err
				}

				viper.Set("log_trigger_load_contract.contract_address", addr)
				viper.Set("log_trigger_load_contract.use_mercury", conf.LogTriggerLoadContract.UseMercury)
				viper.Set("log_trigger_load_contract.use_arbitrum", conf.LogTriggerLoadContract.UseArbitrum)
			case domain.VerifiableLoadConditional:
				if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
					return domain.ErrRegistrarNotAvailable
				}

				deployable := asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
					RegistrarAddr: conf.ServiceContract.RegistrarAddress,
					UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
				})

				addr, err := deployable.Connect(cmd.Context(), args[0], deployer)
				if err != nil {
					return err
				}

				viper.Set("conditional_load_contract.contract_address", addr)
				viper.Set("conditional_load_contract.use_arbitrum", conf.ConditionalLoadContract.UseArbitrum)
			}

			return nil
		},
	}
)
