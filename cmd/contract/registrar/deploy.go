package registrar

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new registrar contract",
		Long:  `Deploy a new registrar contract and add the address and configuration parameters to the environment.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return err
			}

			if env.LinkToken == nil || env.Registry == nil {
				return fmt.Errorf("link token and registry required")
			}

			env.Registrar = &config.AutomationRegistrarV21Contract{
				Type:    config.AutomationRegistrarContractType,
				Version: "v2.1",
				MinLink: 0,
				AutoApprovals: []config.AutomationRegistrarV21AutoApprovalConfig{
					{
						TriggerType:           0,
						AutoApproveType:       2,
						AutoApproveMaxAllowed: 1_000,
					},
					{
						TriggerType:           1,
						AutoApproveType:       2,
						AutoApproveMaxAllowed: 1_000,
					},
				},
			}

			deployable := asset.NewRegistrarV21Deployable(*env.LinkToken, *env.Registry, env.Registrar)

			if _, err := deployable.Deploy(cmd.Context(), deployer); err != nil {
				return err
			}

			env.Registry.Onchain.Registrars = []string{env.Registrar.Address}

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
