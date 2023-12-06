package load

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
)

var (
	cancelCmd = &cobra.Command{
		Use:   "cancel-upkeeps",
		Short: "Cancel all upkeeps on the load contract",
		Long:  `Cancel all upkeeps on the load contract.`,
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

			return runCancelUpkeeps(cmd.Context(), upkeepType, &env, deployer, vlic)
		},
	}
)

type upkeepCanceller interface {
	CancelUpkeeps(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

func runCancelUpkeeps(
	ctx context.Context,
	contractType string,
	env *config.Environment,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var (
		register upkeepCanceller
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

	if err := register.CancelUpkeeps(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}
