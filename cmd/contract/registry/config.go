package registry

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
)

// TODO: override onchain, offchain, and ocr configs from json input

var setConfigCmd = &cobra.Command{
	Use:   "set-config",
	Short: "Set the configuration for the registry",
	Long:  `Set the configuration for the registry including on-chain config and off-chain config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, env, key, err := prepare(cmd)
		if err != nil {
			return err
		}

		deployer, err := asset.NewDeployer(&env, key)
		if err != nil {
			return err
		}

		if env.Registry == nil {
			return fmt.Errorf("registry does not exist")
		}

		env.Registry.OCRNetwork.MaxFaultyNodes = int(maxFaulty)

		interactable := asset.NewRegistryV21Deployable(*env.LinkToken, *env.LinkETH, *env.FastGas, env.Registry)

		if _, err := interactable.Connect(cmd.Context(), deployer); err != nil {
			return err
		}

		if err := interactable.SetOffchainConfig(cmd.Context(), deployer, env.Participants); err != nil {
			return err
		}

		return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
	},
}
