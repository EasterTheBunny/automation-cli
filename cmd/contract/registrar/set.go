package registrar

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
		Short: "Set the address and configuration of an existing registrar contract",
		Long:  `Set the address and configuration of an existing registrar contract and add the address and configuration parameters to the environment.`,
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

			if env.Registry == nil {
				return fmt.Errorf("create or connect to a registry first")
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("provided address must be hex encoded")
			}

			if env.Registrar == nil {
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
			}

			env.Registrar.Address = args[0]

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
