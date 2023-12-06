package load

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/domain"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new verifiable-load contract",
		Long:  `Deploy a new verifiable-load contract.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			if env.Registrar == nil {
				return domain.ErrRegistrarNotAvailable
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return fmt.Errorf("failed to create deployer: %s", err.Error())
			}

			switch upkeepType {
			case domain.VerifiableLoadLogTrigger:
				env.LogLoad = &config.VerifiableLoadContract{
					Type:        config.AutomationVerifiableLoadContractType,
					LoadType:    config.LogTriggerLoad,
					UseMercury:  false,
					UseArbitrum: false,
				}

				deployable, err := asset.NewVerifiableLoadLogTriggerDeployable(*env.Registrar, env.LogLoad)
				if err != nil {
					return err
				}

				if _, err := deployable.Deploy(cmd.Context(), deployer); err != nil {
					return fmt.Errorf("deployment failed: %s", err.Error())
				}
			case domain.VerifiableLoadConditional:
				env.ConditionalLoad = &config.VerifiableLoadContract{
					Type:        config.AutomationVerifiableLoadContractType,
					LoadType:    config.ConditionalLoad,
					UseArbitrum: false,
				}

				deployable, err := asset.NewVerifiableLoadConditionalDeployable(*env.Registrar, env.ConditionalLoad)
				if err != nil {
					return err
				}

				if _, err := deployable.Deploy(cmd.Context(), deployer); err != nil {
					return err
				}
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)

func prepare(cmd *cobra.Command) (io.Environment, config.Environment, config.Key, error) {
	var (
		env config.Environment
		key config.Key
		err error
	)

	path := io.EnvironmentFromContext(cmd.Context())
	if path == nil {
		return io.Environment{}, env, key, fmt.Errorf("environment not found")
	}

	env, err = config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
	if err != nil {
		return io.Environment{}, env, key, err
	}

	keys, err := config.ReadPrivateKeysFrom(path.Root.MustRead(config.PrivateKeyConfigFilename))
	if err != nil {
		return io.Environment{}, env, key, err
	}

	pkOverride, err := cmd.Flags().GetString("key")
	if err != nil {
		return io.Environment{}, env, key, err
	}

	if pkOverride == "" {
		pkOverride = env.PrivateKeyAlias
	}

	key, err = keys.KeyForAlias(pkOverride)
	if err != nil {
		return io.Environment{}, env, key, err
	}

	return *path, env, key, nil
}
