package load

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/domain"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var (
	setCmd = &cobra.Command{
		Use:   "set-address [ADDRESS]",
		Short: "Set address for existing verifiable-load contract",
		Long:  `Set address for existing verifiable-load contract.`,
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

			if env.Registrar == nil {
				return domain.ErrRegistrarNotAvailable
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("provided address must be hex encoded")
			}

			switch upkeepType {
			case domain.VerifiableLoadLogTrigger:
				if env.LogLoad == nil {
					env.LogLoad = &config.VerifiableLoadContract{
						Type:     config.AutomationVerifiableLoadContractType,
						LoadType: config.LogTriggerLoad,
					}
				}

				env.LogLoad.Address = args[0]
			case domain.VerifiableLoadConditional:
				if env.ConditionalLoad == nil {
					env.ConditionalLoad = &config.VerifiableLoadContract{
						Type:     config.AutomationVerifiableLoadContractType,
						LoadType: config.ConditionalLoad,
					}
				}

				env.ConditionalLoad.Address = args[0]
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
