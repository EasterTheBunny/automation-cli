package asset

import "time"

const (
	// default OCR configuration vars
	DefaultOCR3DeltaProgress               = 5 * time.Second
	DefaultOCR3DeltaResend                 = 10 * time.Second
	DefaultOCR3DeltaInitial                = 400 * time.Millisecond
	DefaultOCR3DeltaRound                  = 2500 * time.Millisecond
	DefaultOCR3DeltaGrace                  = 40 * time.Millisecond
	DefaultOCR3DeltaCertifiedCommitRequest = 300 * time.Millisecond
	DefaultOCR3DeltaStage                  = 30 * time.Millisecond

	DefaultOCR3MaxRounds                               = uint64(50)
	DefaultOCR3MaxDurationQuery                        = 20 * time.Millisecond
	DefaultOCR3MaxDurationObservation                  = 1600 * time.Millisecond
	DefaultOCR3MaxDurationShouldAcceptFinalizedReport  = 20 * time.Millisecond
	DefaultOCR3MaxDurationShouldTransmitAcceptedReport = 20 * time.Millisecond

	// default registry on-chain configuration vars
	DefaultPaymentPremiumPPB    = uint32(200_000_000)
	DefaultFlatFeeMicroLink     = uint32(1)
	DefaultCheckGasLimit        = uint32(6_500_000)
	DefaultStalenessSeconds     = int64(90_000)
	DefaultGasCeilingMultiplier = uint16(1)
	DefaultMinUpkeepSpend       = int64(0)
	DefaultMaxPerformGas        = uint32(5_000_000)
	DefaultMaxCheckDataSize     = uint32(5_000)
	DefaultMaxPerformDataSize   = uint32(5_000)
	DefaultMaxRevertDataSize    = uint32(5_000)
	DefaultFallbackGasPrice     = int64(200_000_000)
	DefaultFallbackLinkPrice    = int64(5_000_000_000_000_000_000)

	// default v2.1 plugin (offchain) configuration vars
	DefaultPerformLockoutWindow = int64(75_000)
	DefaultMinConfirmations     = 0
	DefaultGasLimitPerReport    = 5_300_000
	DefaultGasOverheadPerUpkeep = 300_000
	DefaultMaxUpkeepBatchSize   = 1
	DefaultReportBlockLag       = 0
	DefaultSamplingJobDuration  = 3_000
	DefaultTargetInRounds       = 1
	DefaultTargetProbability    = "0.999"
)
