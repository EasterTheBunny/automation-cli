package load

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	cmdContext "github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

var (
	cancelCmd = &cobra.Command{
		Use:   "cancel-upkeeps",
		Short: "Cancel all upkeeps on the load contract",
		Long:  `Cancel all upkeeps on the load contract.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf := cmdContext.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			dConfig := config.GetDeployerConfig(conf)

			keyConf := cmdContext.GetKeyConfigFromContext(cmd.Context())
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

			vlic := asset.VerifiableLoadInteractionConfig{
				ContractAddr:             conf.LogTriggerLoadContract.ContractAddress,
				RegisterUpkeepCount:      upkeepCount,
				RegisteredUpkeepInterval: upkeepInterval,
				CancelBeforeRegister:     cancelUpkeeps,
				SendLINKBeforeRegister:   sendLINK,
			}

			return runCancelUpkeeps(cmd.Context(), upkeepType, conf, deployer, vlic)
		},
	}
)

type upkeepCanceller interface {
	CancelUpkeeps(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

func runCancelUpkeeps(
	ctx context.Context,
	contractType string,
	conf *config.Config,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var register upkeepCanceller

	switch contractType {
	case "conditional":
		register = asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.ConditionalLoadContract.ContractAddress
	case "log-trigger":
		register = asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.LogTriggerLoadContract.ContractAddress
	}

	if err := register.CancelUpkeeps(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}
