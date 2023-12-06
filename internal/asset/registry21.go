package asset

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/easterthebunny/automation-cli/internal/config"
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
	ErrConfiguration      = fmt.Errorf("configuration error")
	ErrContractConnection = fmt.Errorf("contract connection")
	ErrContractCreate     = fmt.Errorf("contract creation")
)

const (
	publicKeyLength = 20
)

type RegistryV21Deployable struct {
	registry *iregistry.IKeeperRegistryMaster
	link     config.LinkTokenContract
	linkFeed config.FeedContract
	gasFeed  config.FeedContract
	rCfg     *config.AutomationRegistryV21Contract
}

func NewRegistryV21Deployable(
	link config.LinkTokenContract,
	linkFeed config.FeedContract,
	gasFeed config.FeedContract,
	rCfg *config.AutomationRegistryV21Contract,
) *RegistryV21Deployable {
	return &RegistryV21Deployable{
		link:     link,
		linkFeed: linkFeed,
		gasFeed:  gasFeed,
		rCfg:     rCfg,
	}
}

func (d *RegistryV21Deployable) Connect(ctx context.Context, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(d.rCfg.Address), deployer)
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

	d.rCfg.Address = registryAddr.Hex()

	return d.connectToInterface(ctx, registryAddr, deployer)
}

func (d *RegistryV21Deployable) SetOffchainConfig(
	ctx context.Context,
	deployer *Deployer,
	nodeConfs []config.NodeConfig,
) error {
	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	networkS, oracleIdentities, _, err := makeOracles(nodeConfs)
	if err != nil {
		return err
	}

	//nolint:lll
	signerOnchainPublicKeys, transmitterAccounts, maxFault, offchainConfigVersion, offchainConfig, err := makeOCRConfigs(d.rCfg, networkS, oracleIdentities)
	if err != nil {
		return err
	}

	signers := make([]common.Address, 0)

	for _, signer := range signerOnchainPublicKeys {
		if len(signer) != publicKeyLength {
			return fmt.Errorf("%w: OnChainPublicKey has wrong length for address", ErrConfiguration)
		}

		signers = append(signers, common.BytesToAddress(signer))
	}

	transmitters := make([]common.Address, 0)

	for _, transmitter := range transmitterAccounts {
		if !common.IsHexAddress(string(transmitter)) {
			return fmt.Errorf("%w: TransmitAccount is not a valid Ethereum address", ErrConfiguration)
		}

		transmitters = append(transmitters, common.HexToAddress(string(transmitter)))
	}

	onchainConfig := makeOnchainConfig(d.rCfg.Onchain)

	trx, err := d.registry.SetConfigTypeSafe(
		opts, signers, transmitters, maxFault,
		onchainConfig, offchainConfigVersion, offchainConfig,
	)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrContractConnection, err.Error())
	}

	if err := deployer.wait(ctx, trx); err != nil {
		return fmt.Errorf("%w: %s", ErrContractConnection, err.Error())
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
		common.HexToAddress(d.link.Address),
		common.HexToAddress(d.linkFeed.Address),
		common.HexToAddress(d.gasFeed.Address),
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

func makeOracles(
	nodeConfs []config.NodeConfig,
) ([]int, []ocr2config.OracleIdentityExtra, []ocr2types.ConfigEncryptionPublicKey, error) {
	var (
		Svar    = make([]int, len(nodeConfs))
		oracles = make([]ocr2config.OracleIdentityExtra, len(nodeConfs))
		keys    = make([]ocr2types.ConfigEncryptionPublicKey, len(nodeConfs))
	)

	for idx, nodeConf := range nodeConfs {
		offchainPkBytes, err := hex.DecodeString(strings.TrimPrefix(nodeConf.OffChainPublicKey, "ocr2off_evm_"))
		if err != nil {
			return Svar, oracles, keys, fmt.Errorf("failed to decode %s: %v", nodeConf.OffChainPublicKey, err)
		}

		offchainPkBytesFixed := [ed25519.PublicKeySize]byte{}
		nBts := copy(offchainPkBytesFixed[:], offchainPkBytes)

		if nBts != ed25519.PublicKeySize {
			return Svar, oracles, keys, fmt.Errorf("%w: wrong num elements copied; offchain", ErrConfiguration)
		}

		configPkBytes, err := hex.DecodeString(strings.TrimPrefix(nodeConf.ConfigPublicKey, "ocr2cfg_evm_"))
		if err != nil {
			return Svar, oracles, keys, fmt.Errorf("failed to decode %s: %v", nodeConf.ConfigPublicKey, err)
		}

		configPkBytesFixed := [ed25519.PublicKeySize]byte{}
		nBts = copy(configPkBytesFixed[:], configPkBytes)

		if nBts != ed25519.PublicKeySize {
			return Svar, oracles, keys, fmt.Errorf("%w: wrong num elements copied; config", ErrConfiguration)
		}

		onchainPkBytes, err := hex.DecodeString(strings.TrimPrefix(nodeConf.OnchainPublicKey, "ocr2on_evm_"))
		if err != nil {
			return Svar, oracles, keys, fmt.Errorf("failed to decode %s: %v", nodeConf.OnchainPublicKey, err)
		}

		keys[idx] = configPkBytesFixed
		oracles[idx] = ocr2config.OracleIdentityExtra{
			OracleIdentity: ocr2config.OracleIdentity{
				OnchainPublicKey:  onchainPkBytes,
				OffchainPublicKey: offchainPkBytesFixed,
				PeerID:            nodeConf.P2PKeyID,
				TransmitAccount:   ocr2types.Account(nodeConf.Address),
			},
			ConfigEncryptionPublicKey: configPkBytesFixed,
		}

		Svar[idx] = 1
	}

	return Svar, oracles, keys, nil
}

func makeOCRConfigs(
	conf *config.AutomationRegistryV21Contract,
	svar []int,
	oracles []ocr2config.OracleIdentityExtra,
) ([]ocr2types.OnchainPublicKey, []ocr2types.Account, uint8, uint64, []byte, error) {
	var (
		signerOnchainPublicKeys []ocr2types.OnchainPublicKey
		transmitterAccounts     []ocr2types.Account
		offchainConfigVersion   uint64
		offchainConfig          []byte
		faultyNodes             uint8
		err                     error
	)

	offC, err := json.Marshal(offchain21config.OffchainConfig{
		PerformLockoutWindow: conf.Offchain.PerformLockoutWindow, // 100 * 3 * 1000, // ~100 block lockout (on mumbai)
		MinConfirmations:     conf.Offchain.MinConfirmations,
		TargetProbability:    conf.Offchain.TargetProbability,
		TargetInRounds:       conf.Offchain.TargetInRounds,
		GasLimitPerReport:    conf.Offchain.GasLimitPerReport,
		GasOverheadPerUpkeep: conf.Offchain.GasOverheadPerUpkeep,
		MaxUpkeepBatchSize:   conf.Offchain.MaxUpkeepBatchSize,
	})
	if err != nil {
		return nil, nil, 0, 0, nil, fmt.Errorf("%w: %s", ErrConfiguration, err.Error())
	}

	//nolint:lll
	signerOnchainPublicKeys, transmitterAccounts, faultyNodes, _, offchainConfigVersion, offchainConfig, err = ocr3confighelper.ContractSetConfigArgsForTests(
		conf.OCRNetwork.DeltaProgress,
		conf.OCRNetwork.DeltaResend,
		conf.OCRNetwork.DeltaInitial,
		conf.OCRNetwork.DeltaRound,
		conf.OCRNetwork.DeltaGrace,
		conf.OCRNetwork.DeltaCertifiedCommitRequest,
		conf.OCRNetwork.DeltaStage,
		conf.OCRNetwork.MaxRounds,
		svar,    // s []int,
		oracles, // oracles []OracleIdentityExtra,
		offC,    // reportingPluginConfig []byte,
		conf.OCRNetwork.MaxDurationQuery,
		conf.OCRNetwork.MaxDurationObservation,
		conf.OCRNetwork.MaxDurationShouldAcceptFinalizedReport,
		conf.OCRNetwork.MaxDurationShouldTransmitAcceptedReport,
		conf.OCRNetwork.MaxFaultyNodes,
		nil, // onchainConfig []byte,
	)
	if err != nil {
		return nil, nil, 0, 0, nil, fmt.Errorf("%w: %s", ErrConfiguration, err.Error())
	}

	return signerOnchainPublicKeys, transmitterAccounts, faultyNodes, offchainConfigVersion, offchainConfig, nil
}

func makeOnchainConfig(onchain config.AutomationV21OnchainConfig) iregistry.KeeperRegistryBase21OnchainConfig {
	registrars := make([]common.Address, len(onchain.Registrars))
	for idx := range registrars {
		registrars[idx] = common.HexToAddress(onchain.Registrars[idx])
	}

	return iregistry.KeeperRegistryBase21OnchainConfig{
		PaymentPremiumPPB:      onchain.PaymentPremiumPPB,
		FlatFeeMicroLink:       onchain.FlatFeeMicroLink,
		CheckGasLimit:          onchain.CheckGasLimit,
		StalenessSeconds:       big.NewInt(onchain.StalenessSeconds),
		GasCeilingMultiplier:   onchain.GasCeilingMultiplier,
		MinUpkeepSpend:         big.NewInt(onchain.MinUpkeepSpend),
		MaxPerformGas:          onchain.MaxPerformGas,
		MaxCheckDataSize:       onchain.MaxCheckDataSize,
		MaxPerformDataSize:     onchain.MaxPerformDataSize,
		MaxRevertDataSize:      onchain.MaxRevertDataSize,
		FallbackGasPrice:       big.NewInt(onchain.FallbackGasPrice),
		FallbackLinkPrice:      big.NewInt(onchain.FallbackLinkPrice),
		Transcoder:             common.HexToAddress(onchain.Transcoder),
		Registrars:             registrars,
		UpkeepPrivilegeManager: common.HexToAddress(onchain.UpkeepPrivilegeManager),
	}
}
