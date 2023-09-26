package asset

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	registrar "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/automation_registrar_wrapper2_1"
)

type RegistrarV21Config struct {
	RegistryAddr  string
	LinkTokenAddr string
	MinLink       int64
}

type RegistrarV21Deployable struct {
	contract *registrar.AutomationRegistrar
	cCfg     *RegistrarV21Config
}

func NewRegistrarV21Deployable(cCfg *RegistrarV21Config) *RegistrarV21Deployable {
	return &RegistrarV21Deployable{
		cCfg: cCfg,
	}
}

func (d *RegistrarV21Deployable) Connect(
	ctx context.Context,
	addr string,
	deployer *Deployer,
) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *RegistrarV21Deployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
	config VerifyContractConfig,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registry := common.HexToAddress(d.cCfg.RegistryAddr)
	linkAddr := common.HexToAddress(d.cCfg.LinkTokenAddr)
	minLink := big.NewInt(d.cCfg.MinLink)

	contractAddr, trx, _, err := registrar.DeployAutomationRegistrar(
		opts, deployer.Client,
		linkAddr, registry, minLink, nil,
	)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: AutomationForwarderLogic creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	PrintVerifyContractCommand(
		config,
		contractAddr.String(),
		d.cCfg.LinkTokenAddr,
		d.cCfg.RegistryAddr,
		minLink.String(),
		"[]",
	)

	return contractAddr, nil
}

func (d *RegistrarV21Deployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := registrar.NewAutomationRegistrar(addr, deployer.Client)
	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}
