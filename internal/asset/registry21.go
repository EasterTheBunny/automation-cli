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
	offchain21config "github.com/smartcontractkit/ocr2keepers/pkg/v3/config"

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
	StalenessSeconds       *big.Int
	GasCeilingMultiplier   uint16
	MinUpkeepSpend         *big.Int
	MaxPerformGas          uint32
	MaxCheckDataSize       uint32
	MaxPerformDataSize     uint32
	MaxRevertDataSize      uint32
	FallbackGasPrice       *big.Int
	FallbackLinkPrice      *big.Int
	Transcoder             string
	Registrar              string
	UpkeepPrivilegeManager string
}

type OCR3NetworkConfig struct {
	DeltaProgress                           time.Duration
	DeltaResend                             time.Duration
	DeltaInitial                            time.Duration
	DeltaRound                              time.Duration
	DeltaGrace                              time.Duration
	DeltaCertifiedCommitRequest             time.Duration
	DeltaStage                              time.Duration
	MaxRounds                               uint64
	MaxDurationQuery                        time.Duration
	MaxDurationObservation                  time.Duration
	MaxDurationShouldAcceptFinalizedReport  time.Duration
	MaxDurationShouldTransmitAcceptedReport time.Duration
	MaxFaultyNodes                          int
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

func (d *RegistryV21Deployable) Deploy(ctx context.Context, deployer *Deployer) (common.Address, error) {
	var registryAddr common.Address

	automationForwarderLogicAddr, err := d.deployForwarder(ctx, deployer)
	if err != nil {
		return registryAddr, err
	}

	registryLogicBAddr, err := d.deployLogicB(ctx, automationForwarderLogicAddr, deployer)
	if err != nil {
		return registryAddr, err
	}

	registryLogicAAddr, err := d.deployLogicA(ctx, registryLogicBAddr, deployer)
	if err != nil {
		return registryAddr, err
	}

	registryAddr, err = d.deployRegistry(ctx, registryLogicAAddr, deployer)
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
	onchain AutomationV21OnchainConfig,
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

	offC, err := json.Marshal(offchain21config.OffchainConfig{
		PerformLockoutWindow: offchain.PerformLockoutWindow, // 100 * 3 * 1000, // ~100 block lockout (on mumbai)
		MinConfirmations:     offchain.MinConfirmations,
		TargetProbability:    offchain.TargetProbability,
		TargetInRounds:       offchain.TargetInRounds,
		GasLimitPerReport:    offchain.GasLimitPerReport,
		GasOverheadPerUpkeep: offchain.GasOverheadPerUpkeep,
		MaxUpkeepBatchSize:   offchain.MaxUpkeepBatchSize,
	})
	if err != nil {
		return err
	}

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
		PaymentPremiumPPB:      onchain.PaymentPremiumPPB,
		FlatFeeMicroLink:       onchain.FlatFeeMicroLink,
		CheckGasLimit:          onchain.CheckGasLimit,
		StalenessSeconds:       onchain.StalenessSeconds,
		GasCeilingMultiplier:   onchain.GasCeilingMultiplier,
		MinUpkeepSpend:         onchain.MinUpkeepSpend,
		MaxPerformGas:          onchain.MaxPerformGas,
		MaxCheckDataSize:       onchain.MaxCheckDataSize,
		MaxPerformDataSize:     onchain.MaxPerformDataSize,
		MaxRevertDataSize:      onchain.MaxRevertDataSize,
		FallbackGasPrice:       onchain.FallbackGasPrice,
		FallbackLinkPrice:      onchain.FallbackLinkPrice,
		Transcoder:             common.HexToAddress(onchain.Transcoder),
		Registrars:             []common.Address{common.HexToAddress(onchain.Registrar)},
		UpkeepPrivilegeManager: common.HexToAddress(onchain.UpkeepPrivilegeManager),
	}

	trx, err := d.registry.SetConfigTypeSafe(opts, signers, transmitters, f, onchainConfig, offchainConfigVersion, offchainConfig)
	if err != nil {
		return err
	}

	if err := deployer.wait(ctx, trx); err != nil {
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

func (d *RegistryV21Deployable) deployForwarder(ctx context.Context, deployer *Deployer) (common.Address, error) {
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
