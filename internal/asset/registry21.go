package asset

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

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
