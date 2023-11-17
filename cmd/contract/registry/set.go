package registry

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var (
	setCmd = &cobra.Command{
		Use:   "set-address [ADDRESS]",
		Short: "Set the address and configuration of an existing registry contract",
		Long:  `Set the address and configuration of an existing registry contract and add the address and configuration parameters to the environment.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := io.EnvironmentFromContext(cmd.Context())
			if path == nil {
				return fmt.Errorf("environment not found")
			}

			env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
			if err != nil {
				return err
			}

			if env.LinkToken == nil || env.LinkETH == nil || env.FastGas == nil {
				return fmt.Errorf("ensure link token, link ETH feed, and fast gas feed have been deployed or set first")
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("provided address must be hex encoded")
			}

			if env.Registry == nil {
				env.Registry = &config.AutomationRegistryV21Contract{
					Type:    config.AutomationRegistryContractType,
					Version: "v2.1",
				}
			}

			env.Registry.Address = args[0]

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
