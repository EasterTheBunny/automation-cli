package asset

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	ocr2config "github.com/smartcontractkit/libocr/offchainreporting2plus/confighelper"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3confighelper"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	offchain20config "github.com/smartcontractkit/ocr2keepers/pkg/v2/config"

	forwarder "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/automation_forwarder_logic"
	iregistry "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/i_keeper_registry_master_wrapper_2_1"
	logica "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/keeper_registry_logic_a_wrapper_2_1"
	logicb "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/keeper_registry_logic_b_wrapper_2_1"
	registry "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/keeper_registry_wrapper_2_1"
)

var (
	ErrContractConnection = fmt.Errorf("contract connection")
	ErrContractCreate     = fmt.Errorf("contract creation")
)

type RegistryV21Config struct {
	Mode            uint8
	LinkTokenAddr   string
	LinkETHFeedAddr string
	FastGasFeedAddr string
}

type OCR2NodeConfig struct {
	Address           string
	OffChainPublicKey string
	ConfigPublicKey   string
	OnchainPublicKey  string
	P2PKeyID          string
}

type AutomationV21OffchainConfig struct {
	MercuryLookup          bool     `json:"mercuryLookup"`
	PaymentPremiumPPB      uint32   `json:"paymentPremiumPPB"`
	FlatFeeMicroLink       uint32   `json:"flatFeeMicroLink"`
	CheckGasLimit          uint32   `json:"checkGasLimit"`
	StalenessSeconds       *big.Int `json:"stalenessSeconds"`
	GasCeilingMultiplier   uint16   `json:"gasCeilingMultiplier"`
	MinUpkeepSpend         *big.Int `json:"minUpkeepSpend"`
	MaxPerformGas          uint32   `json:"maxPerformGas"`
	MaxCheckDataSize       uint32   `json:"maxCheckDataSize"`
	MaxPerformDataSize     uint32   `json:"maxPerformDataSize"`
	MaxRevertDataSize      uint32   `json:"maxRevertDataSize"`
	FallbackGasPrice       *big.Int `json:"fallbackGasPrice"`
	FallbackLinkPrice      *big.Int `json:"fallbackLinkPrice"`
	Transcoder             string   `json:"transcoderAddress"`
	Registrar              string   `json:"registrarAddress"`
	UpkeepPrivilegeManager string   `json:"upkeepPrivilegeManagerAddress"`
}

type OCR3NetworkConfig struct {
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

type RegistryV21Deployable struct {
	registry *iregistry.IKeeperRegistryMaster
	rCfg     *RegistryV21Config
}

func NewRegistryV21Deployable(rCfg *RegistryV21Config) *RegistryV21Deployable {
	return &RegistryV21Deployable{
		rCfg: rCfg,
	}
}

func (d *RegistryV21Deployable) Connect(ctx context.Context, addr string, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *RegistryV21Deployable) Deploy(ctx context.Context, deployer *Deployer, config VerifyContractConfig) (common.Address, error) {
	var registryAddr common.Address

	automationForwarderLogicAddr, err := d.deployForwarder(ctx, deployer, config)
	if err != nil {
		return registryAddr, err
	}

	registryLogicBAddr, err := d.deployLogicB(ctx, automationForwarderLogicAddr, deployer, config)
	if err != nil {
		return registryAddr, err
	}

	registryLogicAAddr, err := d.deployLogicA(ctx, registryLogicBAddr, deployer, config)
	if err != nil {
		return registryAddr, err
	}

	registryAddr, err = d.deployRegistry(ctx, registryLogicAAddr, deployer, config)
	if err != nil {
		return registryAddr, err
	}

	return d.connectToInterface(ctx, registryAddr, deployer)
}

func (d *RegistryV21Deployable) SetOffchainConfig(
	ctx context.Context,
	deployer *Deployer,
	nodeConfs []OCR2NodeConfig,
	ocrConf OCR3NetworkConfig,
	offchain AutomationV21OffchainConfig,
) error {
	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	S := make([]int, len(nodeConfs))
	oracleIdentities := make([]ocr2config.OracleIdentityExtra, len(nodeConfs))
	sharedSecretEncryptionPublicKeys := make([]ocr2types.ConfigEncryptionPublicKey, len(nodeConfs))

	for idx, nodeConf := range nodeConfs {
		offchainPkBytes, err := hex.DecodeString(strings.TrimPrefix(nodeConf.OffChainPublicKey, "ocr2off_evm_"))
		if err != nil {
			return fmt.Errorf("failed to decode %s: %v", nodeConf.OffChainPublicKey, err)
		}

		offchainPkBytesFixed := [ed25519.PublicKeySize]byte{}
		n := copy(offchainPkBytesFixed[:], offchainPkBytes)
		if n != ed25519.PublicKeySize {
			return fmt.Errorf("wrong num elements copied")
		}

		configPkBytes, err := hex.DecodeString(strings.TrimPrefix(nodeConf.ConfigPublicKey, "ocr2cfg_evm_"))
		if err != nil {
			return fmt.Errorf("failed to decode %s: %v", nodeConf.ConfigPublicKey, err)
		}

		configPkBytesFixed := [ed25519.PublicKeySize]byte{}
		n = copy(configPkBytesFixed[:], configPkBytes)
		if n != ed25519.PublicKeySize {
			return fmt.Errorf("wrong num elements copied")
		}

		onchainPkBytes, err := hex.DecodeString(strings.TrimPrefix(nodeConf.OnchainPublicKey, "ocr2on_evm_"))
		if err != nil {
			return fmt.Errorf("failed to decode %s: %v", nodeConf.OnchainPublicKey, err)
		}

		sharedSecretEncryptionPublicKeys[idx] = configPkBytesFixed
		oracleIdentities[idx] = ocr2config.OracleIdentityExtra{
			OracleIdentity: ocr2config.OracleIdentity{
				OnchainPublicKey:  onchainPkBytes,
				OffchainPublicKey: offchainPkBytesFixed,
				PeerID:            nodeConf.P2PKeyID,
				TransmitAccount:   ocr2types.Account(nodeConf.Address),
			},
			ConfigEncryptionPublicKey: configPkBytesFixed,
		}

		S[idx] = 1
	}

	offC, err := json.Marshal(offchain20config.OffchainConfig{
		PerformLockoutWindow: 100 * 3 * 1000, // ~100 block lockout (on mumbai)
		MinConfirmations:     1,
		MercuryLookup:        offchain.MercuryLookup,
	})
	if err != nil {
		return err
	}

	ocrConf = overrideDefaultsOCR3NetworkConfig(ocrConf)

	signerOnchainPublicKeys, transmitterAccounts, f, _, offchainConfigVersion, offchainConfig, err := ocr3confighelper.ContractSetConfigArgsForTests(
		ocrConf.DeltaProgress,
		ocrConf.DeltaResend,
		ocrConf.DeltaInitial,
		ocrConf.DeltaRound,
		ocrConf.DeltaGrace,
		ocrConf.DeltaCertifiedCommitRequest,
		ocrConf.DeltaStage,
		ocrConf.MaxRounds,
		S,                // s []int,
		oracleIdentities, // oracles []OracleIdentityExtra,
		offC,             // reportingPluginConfig []byte,
		ocrConf.MaxDurationQuery,
		ocrConf.MaxDurationObservation,
		ocrConf.MaxDurationShouldAcceptFinalizedReport,
		ocrConf.MaxDurationShouldTransmitAcceptedReport,
		ocrConf.MaxFaultyNodes,
		nil, // onchainConfig []byte,
	)
	if err != nil {
		return err
	}

	signers := make([]common.Address, 0)

	for _, signer := range signerOnchainPublicKeys {
		if len(signer) != 20 {
			return fmt.Errorf("OnChainPublicKey has wrong length for address")
		}

		signers = append(signers, common.BytesToAddress(signer))
	}

	transmitters := make([]common.Address, 0)

	for _, transmitter := range transmitterAccounts {
		if !common.IsHexAddress(string(transmitter)) {
			return fmt.Errorf("TransmitAccount is not a valid Ethereum address")
		}

		transmitters = append(transmitters, common.HexToAddress(string(transmitter)))
	}

	onchainConfig := iregistry.KeeperRegistryBase21OnchainConfig{
		PaymentPremiumPPB:      offchain.PaymentPremiumPPB,
		FlatFeeMicroLink:       offchain.FlatFeeMicroLink,
		CheckGasLimit:          offchain.CheckGasLimit,
		StalenessSeconds:       offchain.StalenessSeconds,
		GasCeilingMultiplier:   offchain.GasCeilingMultiplier,
		MinUpkeepSpend:         offchain.MinUpkeepSpend,
		MaxPerformGas:          offchain.MaxPerformGas,
		MaxCheckDataSize:       offchain.MaxCheckDataSize,
		MaxPerformDataSize:     offchain.MaxPerformDataSize,
		MaxRevertDataSize:      offchain.MaxRevertDataSize,
		FallbackGasPrice:       offchain.FallbackGasPrice,
		FallbackLinkPrice:      offchain.FallbackLinkPrice,
		Transcoder:             common.HexToAddress(offchain.Transcoder),
		Registrars:             []common.Address{common.HexToAddress(offchain.Registrar)},
		UpkeepPrivilegeManager: common.HexToAddress(offchain.UpkeepPrivilegeManager),
	}

	trx, err := d.registry.SetConfigTypeSafe(opts, signers, transmitters, f, onchainConfig, offchainConfigVersion, offchainConfig)
	if err != nil {
		return err
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return err
	}

	return nil
}

func (d *RegistryV21Deployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := iregistry.NewIKeeperRegistryMaster(
		addr,
		deployer.Client,
	)

	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.registry = contract

	return addr, nil
}

func (d *RegistryV21Deployable) deployForwarder(ctx context.Context, deployer *Deployer, config VerifyContractConfig) (common.Address, error) {
	var logicAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return logicAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	logicAddr, tx, _, err := forwarder.DeployAutomationForwarderLogic(opts, deployer.Client)
	if err != nil {
		return logicAddr, fmt.Errorf("%w: AutomationForwarderLogic creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, tx); err != nil {
		return logicAddr, err
	}

	// PrintVerifyContractCommand(config, logicAddr.String())

	return logicAddr, nil
}

func (d *RegistryV21Deployable) deployLogicB(
	ctx context.Context,
	forwarderAddr common.Address,
	deployer *Deployer,
	config VerifyContractConfig,
) (common.Address, error) {
	var logicAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return logicAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	logicAddr, trx, _, err := logicb.DeployKeeperRegistryLogicB(
		opts,
		deployer.Client,
		d.rCfg.Mode,
		common.HexToAddress(d.rCfg.LinkTokenAddr),
		common.HexToAddress(d.rCfg.LinkETHFeedAddr),
		common.HexToAddress(d.rCfg.FastGasFeedAddr),
		forwarderAddr,
	)

	if err != nil {
		return logicAddr, fmt.Errorf("%w: deploy LogicB ABI failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return logicAddr, err
	}

	/*
		PrintVerifyContractCommand(
			config,
			logicAddr.String(),
			fmt.Sprintf("%d", d.rCfg.Mode),
			d.rCfg.LinkTokenAddr,
			d.rCfg.LinkETHFeedAddr,
			d.rCfg.FastGasFeedAddr,
			forwarderAddr.String(),
		)
	*/

	return logicAddr, nil
}

func (d *RegistryV21Deployable) deployLogicA(
	ctx context.Context,
	logicBAddr common.Address,
	deployer *Deployer,
	config VerifyContractConfig,
) (common.Address, error) {
	var logicAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return logicAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	logicAddr, trx, _, err := logica.DeployKeeperRegistryLogicA(
		opts,
		deployer.Client,
		logicBAddr,
	)

	if err != nil {
		return logicAddr, fmt.Errorf("%w: deploy LogicA ABI failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return logicAddr, err
	}

	// PrintVerifyContractCommand(config, logicAddr.String(), logicBAddr.String())

	return logicAddr, nil
}

func (d *RegistryV21Deployable) deployRegistry(
	ctx context.Context,
	logicAAddr common.Address,
	deployer *Deployer,
	config VerifyContractConfig,
) (common.Address, error) {
	var registryAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return registryAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registryAddr, trx, _, err := registry.DeployKeeperRegistry(
		opts,
		deployer.Client,
		logicAAddr,
	)

	if err != nil {
		return registryAddr, fmt.Errorf("%w: deploy Registry ABI failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return registryAddr, err
	}

	// PrintVerifyContractCommand(config, registryAddr.String(), logicAAddr.String())

	// fmt.Printf("registry deployed to: %s\n", util.ExplorerLink(deployer.Config.ChainID, trx.Hash()))

	return registryAddr, nil
}

func overrideDefaultsOCR3NetworkConfig(conf OCR3NetworkConfig) OCR3NetworkConfig {
	defaultConf := OCR3NetworkConfig{
		DeltaProgress:                           5 * time.Second,
		DeltaResend:                             10 * time.Second,
		DeltaInitial:                            400 * time.Millisecond,
		DeltaRound:                              2500 * time.Millisecond,
		DeltaGrace:                              40 * time.Millisecond,
		DeltaCertifiedCommitRequest:             300 * time.Millisecond,
		DeltaStage:                              30 * time.Second,
		MaxRounds:                               50,
		MaxDurationQuery:                        20 * time.Millisecond,
		MaxDurationObservation:                  1600 * time.Millisecond,
		MaxDurationShouldAcceptFinalizedReport:  20 * time.Millisecond,
		MaxDurationShouldTransmitAcceptedReport: 20 * time.Millisecond,
		MaxFaultyNodes:                          1,
	}

	if conf.DeltaProgress != 0 {
		defaultConf.DeltaProgress = conf.DeltaProgress
	}

	if conf.DeltaResend != 0 {
		defaultConf.DeltaResend = conf.DeltaResend
	}

	if conf.DeltaInitial != 0 {
		defaultConf.DeltaInitial = conf.DeltaInitial
	}

	if conf.DeltaRound != 0 {
		defaultConf.DeltaRound = conf.DeltaRound
	}

	if conf.DeltaGrace != 0 {
		defaultConf.DeltaGrace = conf.DeltaGrace
	}

	if conf.DeltaCertifiedCommitRequest != 0 {
		defaultConf.DeltaCertifiedCommitRequest = conf.DeltaCertifiedCommitRequest
	}

	if conf.DeltaStage != 0 {
		defaultConf.DeltaStage = conf.DeltaStage
	}

	if conf.MaxRounds != 0 {
		defaultConf.MaxRounds = conf.MaxRounds
	}

	if conf.MaxDurationQuery != 0 {
		defaultConf.MaxDurationQuery = conf.MaxDurationQuery
	}

	if conf.MaxDurationObservation != 0 {
		defaultConf.MaxDurationObservation = conf.MaxDurationObservation
	}

	if conf.MaxDurationShouldAcceptFinalizedReport != 0 {
		defaultConf.MaxDurationShouldAcceptFinalizedReport = conf.MaxDurationShouldAcceptFinalizedReport
	}

	if conf.MaxDurationShouldTransmitAcceptedReport != 0 {
		defaultConf.MaxDurationShouldTransmitAcceptedReport = conf.MaxDurationShouldTransmitAcceptedReport
	}

	if conf.MaxFaultyNodes != 0 {
		defaultConf.MaxFaultyNodes = conf.MaxFaultyNodes
	}

	return defaultConf
}
