package config

import (
	"math/big"
	"time"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/spf13/viper"
)

func GetOCRConfig(path string) (*AutomationNetworkConfig, *viper.Viper, error) {
	configPath, err := ensureExists(path, "config.json")
	if err != nil {
		return nil, nil, err
	}

	vpr := viper.New()

	vpr.SetConfigFile(configPath)
	vpr.SetConfigType("json")

	setOCRConfigDefaults(vpr)

	conf, err := readConfig[AutomationNetworkConfig](vpr, path)
	if err != nil {
		return nil, nil, err
	}

	return conf, vpr, nil
}

type AutomationNetworkConfig struct {
	Network  OCR3NetworkConfig           `mapstructure:"network"`
	Onchain  AutomationOnchainConfigV21  `mapstructure:"onchain"`
	Offchain AutomationOffchainConfigV21 `mapstructure:"offchain"`
}

func GetAssetOCRConfig(conf *Config) asset.OCR3NetworkConfig {
	return asset.OCR3NetworkConfig{
		DeltaProgress:                           conf.OCR.Network.DeltaProgress,
		DeltaResend:                             conf.OCR.Network.DeltaResend,
		DeltaInitial:                            conf.OCR.Network.DeltaInitial,
		DeltaRound:                              conf.OCR.Network.DeltaRound,
		DeltaGrace:                              conf.OCR.Network.DeltaGrace,
		DeltaCertifiedCommitRequest:             conf.OCR.Network.DeltaCertifiedCommitRequest,
		DeltaStage:                              conf.OCR.Network.DeltaStage,
		MaxRounds:                               conf.OCR.Network.MaxRounds,
		MaxDurationQuery:                        conf.OCR.Network.MaxDurationQuery,
		MaxDurationObservation:                  conf.OCR.Network.MaxDurationObservation,
		MaxDurationShouldAcceptFinalizedReport:  conf.OCR.Network.MaxDurationShouldAcceptFinalizedReport,
		MaxDurationShouldTransmitAcceptedReport: conf.OCR.Network.MaxDurationShouldTransmitAcceptedReport,
		MaxFaultyNodes:                          conf.OCR.Network.MaxFaultyNodes,
	}
}

type OCR3NetworkConfig struct {
	DeltaProgress                           time.Duration `mapstructure:"delta_pogress"`
	DeltaResend                             time.Duration `mapstructure:"delta_resend"`
	DeltaInitial                            time.Duration `mapstructure:"delta_initial"`
	DeltaRound                              time.Duration `mapstructure:"delta_round"`
	DeltaGrace                              time.Duration `mapstructure:"delta_grace"`
	DeltaCertifiedCommitRequest             time.Duration `mapstructure:"delta_certified_commit_request"`
	DeltaStage                              time.Duration `mapstructure:"delta_stage"`
	MaxRounds                               uint64        `mapstructure:"max_rounds"`
	MaxDurationQuery                        time.Duration `mapstructure:"max_duration_query"`
	MaxDurationObservation                  time.Duration `mapstructure:"max_duration_observation"`
	MaxDurationShouldAcceptFinalizedReport  time.Duration `mapstructure:"max_duration_should_accept_finalized_report"`
	MaxDurationShouldTransmitAcceptedReport time.Duration `mapstructure:"max_duration_should_transmit_accepted_report"`
	MaxFaultyNodes                          int           `mapstructure:"max_faulty_nodes"`
}

func GetOnchainConfig(conf *Config) asset.AutomationV21OnchainConfig {
	return asset.AutomationV21OnchainConfig{
		PaymentPremiumPPB:      conf.OCR.Onchain.PaymentPremiumPPB,
		FlatFeeMicroLink:       conf.OCR.Onchain.FlatFeeMicroLink,
		CheckGasLimit:          conf.OCR.Onchain.CheckGasLimit,
		StalenessSeconds:       big.NewInt(conf.OCR.Onchain.StalenessSeconds),
		GasCeilingMultiplier:   conf.OCR.Onchain.GasCeilingMultiplier,
		MinUpkeepSpend:         big.NewInt(conf.OCR.Onchain.MinUpkeepSpend),
		MaxPerformGas:          conf.OCR.Onchain.MaxPerformGas,
		MaxCheckDataSize:       conf.OCR.Onchain.MaxCheckDataSize,
		MaxPerformDataSize:     conf.OCR.Onchain.MaxPerformDataSize,
		MaxRevertDataSize:      conf.OCR.Onchain.MaxRevertDataSize,
		FallbackGasPrice:       big.NewInt(conf.OCR.Onchain.FallbackGasPrice),
		FallbackLinkPrice:      new(big.Int).SetUint64(conf.OCR.Onchain.FallbackLinkPrice),
		Transcoder:             conf.OCR.Onchain.Transcoder,
		Registrar:              conf.ServiceContract.RegistrarAddress,
		UpkeepPrivilegeManager: conf.OCR.Onchain.UpkeepPrivilegeManager,
	}
}

type AutomationOnchainConfigV21 struct {
	PaymentPremiumPPB      uint32 `mapstructure:"payment_premium_ppb"`
	FlatFeeMicroLink       uint32 `mapstructure:"flat_fee_micro_link"`
	CheckGasLimit          uint32 `mapstructure:"check_gas_limit"`
	StalenessSeconds       int64  `mapstructure:"staleness_seconds"`
	GasCeilingMultiplier   uint16 `mapstructure:"gas_ceiling_multiplier"`
	MinUpkeepSpend         int64  `mapstructure:"min_upkeep_spend"`
	MaxPerformGas          uint32 `mapstructure:"max_perform_gas"`
	MaxCheckDataSize       uint32 `mapstructure:"max_check_data_size"`
	MaxPerformDataSize     uint32 `mapstructure:"max_perform_data_size"`
	MaxRevertDataSize      uint32 `mapstructure:"max_revert_data_size"`
	FallbackGasPrice       int64  `mapstructure:"fallback_gas_price"`
	FallbackLinkPrice      uint64 `mapstructure:"fallback_link_price"`
	Transcoder             string `mapstructure:"transcoder_address"`
	UpkeepPrivilegeManager string `mapstructure:"upkeep_privilege_manager_address"`
}

func GetOffchainConfig(conf *Config) asset.AutomationV21OffchainConfig {
	return asset.AutomationV21OffchainConfig{
		PerformLockoutWindow: conf.OCR.Offchain.PerformLockoutWindow,
		MinConfirmations:     conf.OCR.Offchain.MinConfirmations,
		TargetProbability:    conf.OCR.Offchain.TargetProbability,
		TargetInRounds:       conf.OCR.Offchain.TargetInRounds,
		GasLimitPerReport:    uint32(conf.OCR.Offchain.GasLimitPerReport),
		GasOverheadPerUpkeep: uint32(conf.OCR.Offchain.GasOverheadPerUpkeep),
		MaxUpkeepBatchSize:   conf.OCR.Offchain.MaxUpkeepBatchSize,
	}
}

type AutomationOffchainConfigV21 struct {
	PerformLockoutWindow int64  `mapstructure:"perform_lockout_window"`
	MinConfirmations     int    `mapstructure:"min_confirmations"`
	TargetProbability    string `mapstructure:"target_probability"`
	TargetInRounds       int    `mapstructure:"target_in_rounds"`
	GasLimitPerReport    int    `mapstructure:"gas_limit_per_report"`
	GasOverheadPerUpkeep int    `mapstructure:"gas_overhead_per_upkeep"`
	MaxUpkeepBatchSize   int    `mapstructure:"max_upkeep_batch_size"`
}
