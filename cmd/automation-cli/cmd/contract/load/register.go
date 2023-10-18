package load

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	cmdContext "github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

func init() {
	registerCmd.Flags().BoolVar(&cancelUpkeeps, "cancel-upkeeps", false, "cancel upkeeps before creating new ones")
	registerCmd.Flags().BoolVar(&sendLINK, "send-link", false, "send LINK to the contract before creating upkeeps")
	registerCmd.Flags().Uint8Var(&upkeepCount, "count", 5, "number of upkeeps to register")
	registerCmd.Flags().Uint32Var(&upkeepInterval, "interval", 15, "eligibility interval for conditional upkeeps")
}

var (
	cancelUpkeeps  bool
	sendLINK       bool
	upkeepCount    uint8
	upkeepInterval uint32

	registerCmd = &cobra.Command{
		Use:   "register-upkeeps",
		Short: "Register new upkeeps to measure load statistics",
		Long:  `Register new upkeeps to measure load statistics.`,
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

			return runRegisterUpkeeps(cmd.Context(), upkeepType, conf, deployer, vlic)
		},
	}
)

type upkeepRegister interface {
	RegisterUpkeeps(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

func runRegisterUpkeeps(
	ctx context.Context,
	contractType string,
	conf *config.Config,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var register upkeepRegister

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

	if err := register.RegisterUpkeeps(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}
