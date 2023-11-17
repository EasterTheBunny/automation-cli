package load

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
)

var (
	readStatsCmd = &cobra.Command{
		Use:   "get-stats",
		Short: "Get delay statistics for load contract",
		Long:  `Get delay statistics for load contract.`,
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

			return runGetStats(cmd.Context(), upkeepType, &env, deployer, vlic)
		},
	}
)

type statsReader interface {
	ReadStats(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

func runGetStats(
	ctx context.Context,
	contractType string,
	env *config.Environment,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var (
		reader statsReader
		err    error
	)

	switch contractType {
	case "conditional":
		if reader, err = asset.NewVerifiableLoadConditionalDeployable(*env.Registrar, env.ConditionalLoad); err != nil {
			return err
		}
	case "log-trigger":
		if reader, err = asset.NewVerifiableLoadLogTriggerDeployable(*env.Registrar, env.LogLoad); err != nil {
			return err
		}
	}

	if err := reader.ReadStats(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}
