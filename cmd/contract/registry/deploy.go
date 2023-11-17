package registry

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
		Use:       "deploy",
		Short:     "Deploy a new registry contract",
		Long:      `Deploy a new registry contract and add the address and configuration parameters to the environment.`,
		ValidArgs: domain.ContractNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			if env.LinkToken == nil || env.LinkETH == nil || env.FastGas == nil {
				return fmt.Errorf("ensure link token, link ETH feed, and fast gas feed have been deployed or set first")
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return err
			}

			env.Registry = &config.AutomationRegistryV21Contract{
				Type:    config.AutomationRegistryContractType,
				Version: "v2.1",
				Mode:    config.GetRegistryMode(mode),
				Offchain: config.AutomationV21OffchainConfig{
					PerformLockoutWindow: asset.DefaultPerformLockoutWindow,
					MinConfirmations:     asset.DefaultMinConfirmations,
					TargetProbability:    asset.DefaultTargetProbability,
					TargetInRounds:       asset.DefaultTargetInRounds,
					GasLimitPerReport:    asset.DefaultGasLimitPerReport,
					GasOverheadPerUpkeep: asset.DefaultGasOverheadPerUpkeep,
					MaxUpkeepBatchSize:   asset.DefaultMaxUpkeepBatchSize,
				},
				Onchain: config.AutomationV21OnchainConfig{
					PaymentPremiumPPB:      asset.DefaultPaymentPremiumPPB,
					FlatFeeMicroLink:       asset.DefaultFlatFeeMicroLink,
					CheckGasLimit:          asset.DefaultCheckGasLimit,
					StalenessSeconds:       asset.DefaultStalenessSeconds,
					GasCeilingMultiplier:   asset.DefaultGasCeilingMultiplier,
					MinUpkeepSpend:         asset.DefaultMinUpkeepSpend,
					MaxPerformGas:          asset.DefaultMaxPerformGas,
					MaxCheckDataSize:       asset.DefaultMaxCheckDataSize,
					MaxPerformDataSize:     asset.DefaultMaxPerformDataSize,
					MaxRevertDataSize:      asset.DefaultMaxRevertDataSize,
					FallbackGasPrice:       asset.DefaultFallbackGasPrice,
					FallbackLinkPrice:      asset.DefaultFallbackLinkPrice,
					Transcoder:             "0x", // not supported
					UpkeepPrivilegeManager: "0x", // not supported
				},
				OCRNetwork: config.OCR3NetworkConfig{
					Version:                                 "v3",
					DeltaProgress:                           asset.DefaultOCR3DeltaProgress,
					DeltaResend:                             asset.DefaultOCR3DeltaResend,
					DeltaInitial:                            asset.DefaultOCR3DeltaInitial,
					DeltaRound:                              asset.DefaultOCR3DeltaRound,
					DeltaGrace:                              asset.DefaultOCR3DeltaGrace,
					DeltaCertifiedCommitRequest:             asset.DefaultOCR3DeltaCertifiedCommitRequest,
					DeltaStage:                              asset.DefaultOCR3DeltaStage,
					MaxRounds:                               asset.DefaultOCR3MaxRounds,
					MaxDurationQuery:                        asset.DefaultOCR3MaxDurationQuery,
					MaxDurationObservation:                  asset.DefaultOCR3MaxDurationObservation,
					MaxDurationShouldAcceptFinalizedReport:  asset.DefaultOCR3MaxDurationShouldAcceptFinalizedReport,
					MaxDurationShouldTransmitAcceptedReport: asset.DefaultOCR3MaxDurationShouldAcceptFinalizedReport,
				},
			}

			if env.Registrar != nil {
				env.Registry.Onchain.Registrars = []string{env.Registrar.Address}
			}

			deployable := asset.NewRegistryV21Deployable(*env.LinkToken, *env.LinkETH, *env.FastGas, env.Registry)

			if _, err := deployable.Deploy(cmd.Context(), deployer); err != nil {
				return err
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
