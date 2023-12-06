package config

import (
	"time"
)

type Environment struct {
	Groupname       string `toml:"group-name"`
	WSURL           string `toml:"ws-url"`
	HTTPURL         string `toml:"http-url"`
	ChainID         int64  `toml:"chain-id"`
	PrivateKeyAlias string `toml:"private-key-alias"`
	GasLimit        uint64 `toml:"deployer-gas-limit"`
	Verifier        *Verifier

	LinkToken       *LinkTokenContract
	LinkETH         *FeedContract
	FastGas         *FeedContract
	Registry        *AutomationRegistryV21Contract
	Registrar       *AutomationRegistrarV21Contract
	LogLoad         *VerifiableLoadContract
	ConditionalLoad *VerifiableLoadContract

	Bootstrap    *NodeConfig
	Participants []NodeConfig
}

type ContractType string

const (
	AutomationRegistrarContractType      ContractType = "automation-registrar"
	AutomationRegistryContractType       ContractType = "automation-registry"
	AutomationVerifiableLoadContractType ContractType = "automation-verifiable-load"
)

type FeedContract struct {
	Address       string
	Mocked        bool
	DefaultAnswer uint64
}

type LinkTokenContract struct {
	Address string
	Mocked  bool
}

type AutomationRegistrarV21Contract struct {
	Type          ContractType
	Version       string
	Address       string
	MinLink       int64
	AutoApprovals []AutomationRegistrarV21AutoApprovalConfig
}

type AutomationRegistrarV21AutoApprovalConfig struct {
	TriggerType           uint8
	AutoApproveType       uint8
	AutoApproveMaxAllowed uint32
}

func GetRegistryMode(mode string) uint8 {
	switch mode {
	case "ARBITRUM":
		return 1
	case "OPTIMISM":
		return 2
	default:
		return 0
	}
}

type AutomationRegistryV21Contract struct {
	Type       ContractType
	Version    string
	Address    string
	Mode       uint8
	Offchain   AutomationV21OffchainConfig
	Onchain    AutomationV21OnchainConfig
	OCRNetwork OCR3NetworkConfig
}

type AutomationV21OffchainConfig struct {
	PerformLockoutWindow int64
	MinConfirmations     int
	TargetProbability    string
	TargetInRounds       int
	GasLimitPerReport    uint32
	GasOverheadPerUpkeep uint32
	MaxUpkeepBatchSize   int
}

type AutomationV21OnchainConfig struct {
	PaymentPremiumPPB      uint32
	FlatFeeMicroLink       uint32
	CheckGasLimit          uint32
	StalenessSeconds       int64
	GasCeilingMultiplier   uint16
	MinUpkeepSpend         int64
	MaxPerformGas          uint32
	MaxCheckDataSize       uint32
	MaxPerformDataSize     uint32
	MaxRevertDataSize      uint32
	FallbackGasPrice       int64
	FallbackLinkPrice      int64
	Transcoder             string
	Registrars             []string
	UpkeepPrivilegeManager string
}

type OCR3NetworkConfig struct {
	Version                                 string        `json:"-"`
	DeltaProgress                           time.Duration `json:"deltaProgress,omitempty"`
	DeltaResend                             time.Duration `json:"deltaResend,omitempty"`
	DeltaInitial                            time.Duration `json:"deltaInitial,omitempty"`
	DeltaRound                              time.Duration `json:"deltaRound,omitempty"`
	DeltaGrace                              time.Duration `json:"deltaGrace,omitempty"`
	DeltaCertifiedCommitRequest             time.Duration `json:"deltaCertifiedCommitRequest,omitempty"`
	DeltaStage                              time.Duration `json:"deltaStage,omitempty"`
	MaxRounds                               uint64        `json:"maxRounds,omitempty"`
	MaxDurationQuery                        time.Duration `json:"maxDurationQuery,omitempty"`
	MaxDurationObservation                  time.Duration `json:"maxDurationObservation,omitempty"`
	MaxDurationShouldAcceptFinalizedReport  time.Duration `json:"maxDurationShouldAcceptFinalizedReport,omitempty"`
	MaxDurationShouldTransmitAcceptedReport time.Duration `json:"maxDurationShouldTransmitAcceptedReport,omitempty"`
	MaxFaultyNodes                          int           `json:"maxFaultyNodes,omitempty"`
}

type VerifiableLoadType string

const (
	ConditionalLoad VerifiableLoadType = "conditional"
	LogTriggerLoad  VerifiableLoadType = "log-trigger"
)

type VerifiableLoadContract struct {
	Type        ContractType
	LoadType    VerifiableLoadType
	Address     string
	UseMercury  bool
	UseArbitrum bool
}

type Verifier struct {
	ContractsDirectory string
	ExplorerAPIKey     string
	NetworkName        string
}

type NodeHostType string

const (
	Docker NodeHostType = "docker"
)

type NodeConfig struct {
	// Node instance configurations
	HostType      NodeHostType
	Name          string
	Image         string
	ManagementURL string
	LogLevel      string
	ListenPort    uint16
	LoginName     string
	LoginPassword string

	IsBootstrap         bool
	BootstrapAddress    string
	BootstrapListenPort uint16

	// Chain configurations
	PrivateKeyAlias string
	Address         string
	ChainID         int64
	WSURL           string
	HTTPURL         string

	// Mercury connection configurations
	MercuryLegacyURL string
	MercuryURL       string
	MercuryID        string
	MercuryKey       string

	// OCR network participant configurations
	OffChainPublicKey string
	ConfigPublicKey   string
	OnchainPublicKey  string
	P2PKeyID          string
}
