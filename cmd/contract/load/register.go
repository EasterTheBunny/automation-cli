package load

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
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
			_, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return err
			}

			vlic := asset.VerifiableLoadInteractionConfig{
				RegisterUpkeepCount:      upkeepCount,
				RegisteredUpkeepInterval: upkeepInterval,
				CancelBeforeRegister:     cancelUpkeeps,
				SendLINKBeforeRegister:   sendLINK,
			}

			return runRegisterUpkeeps(cmd.Context(), upkeepType, &env, deployer, vlic)
		},
	}
)

type upkeepRegister interface {
	RegisterUpkeeps(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

func runRegisterUpkeeps(
	ctx context.Context,
	contractType string,
	env *config.Environment,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var (
		register upkeepRegister
		err      error
	)

	switch contractType {
	case "conditional":
		if register, err = asset.NewVerifiableLoadConditionalDeployable(*env.Registrar, env.ConditionalLoad); err != nil {
			return err
		}
	case "log-trigger":
		if register, err = asset.NewVerifiableLoadLogTriggerDeployable(*env.Registrar, env.LogLoad); err != nil {
			return err
		}
	}

	if err := register.RegisterUpkeeps(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}
