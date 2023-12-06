package asset

import (
	"context"
	"fmt"
	"math/big"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/ethereum/go-ethereum/common"
	registrar "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/automation_registrar_wrapper2_1"
)

type RegistrarV21Deployable struct {
	contract *registrar.AutomationRegistrar
	link     config.LinkTokenContract
	registry config.AutomationRegistryV21Contract
	cCfg     *config.AutomationRegistrarV21Contract
}

func NewRegistrarV21Deployable(
	link config.LinkTokenContract,
	registry config.AutomationRegistryV21Contract,
	cCfg *config.AutomationRegistrarV21Contract,
) *RegistrarV21Deployable {
	return &RegistrarV21Deployable{
		link:     link,
		registry: registry,
		cCfg:     cCfg,
	}
}

func (d *RegistrarV21Deployable) Connect(ctx context.Context, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(d.cCfg.Address), deployer)
}

func (d *RegistrarV21Deployable) Deploy(ctx context.Context, deployer *Deployer) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registry := common.HexToAddress(d.registry.Address)
	linkAddr := common.HexToAddress(d.link.Address)
	minLink := big.NewInt(d.cCfg.MinLink)

	configs := []registrar.AutomationRegistrar21InitialTriggerConfig{}

	for _, conf := range d.cCfg.AutoApprovals {
		configs = append(configs, registrar.AutomationRegistrar21InitialTriggerConfig{
			TriggerType:           conf.TriggerType,
			AutoApproveType:       conf.AutoApproveType,
			AutoApproveMaxAllowed: conf.AutoApproveMaxAllowed,
		})
	}

	contractAddr, trx, _, err := registrar.DeployAutomationRegistrar(
		opts, deployer.Client,
		linkAddr, registry, minLink, configs,
	)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: AutomationForwarderLogic creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	/*
		PrintVerifyContractCommand(
			config,
			contractAddr.String(),
			d.cCfg.LinkTokenAddr,
			d.cCfg.RegistryAddr,
			minLink.String(),
			"[]",
		)
	*/

	d.cCfg.Address = contractAddr.Hex()

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
