package asset

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	verifiableLogTrigger "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_log_trigger_upkeep_wrapper"
	verifiableConditional "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/verifiable_load_upkeep_wrapper"
)

type VerifiableLoadConfig struct {
	RegistrarAddr string
	UseMercury    bool
	UseArbitrum   bool
	AutoLog       bool
}

type VerifiableLoadLogTriggerDeployable struct {
	contract *verifiableLogTrigger.VerifiableLoadLogTriggerUpkeep
	cfg      *DeployerConfig
	cCfg     *VerifiableLoadConfig
}

func NewVerifiableLoadLogTriggerDeployable(
	cfg *DeployerConfig,
	cCfg *VerifiableLoadConfig,
) *VerifiableLoadLogTriggerDeployable {
	return &VerifiableLoadLogTriggerDeployable{
		cfg:  cfg,
		cCfg: cCfg,
	}
}

func (d *VerifiableLoadLogTriggerDeployable) Connect(
	ctx context.Context,
	addr string,
	deployer *Deployer,
) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *VerifiableLoadLogTriggerDeployable) Deploy(ctx context.Context, deployer *Deployer, config VerifyContractConfig) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.cCfg.RegistrarAddr)

	contractAddr, trx, _, err := verifiableLogTrigger.DeployVerifiableLoadLogTriggerUpkeep(
		opts, deployer.Client, registrar,
		d.cCfg.UseArbitrum, d.cCfg.AutoLog, d.cCfg.UseMercury)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: Verifiable Load Contract creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	PrintVerifyContractCommand(
		config,
		contractAddr.String(),
		registrar.String(),
		fmt.Sprintf("%t", d.cCfg.UseArbitrum),
		fmt.Sprintf("%t", d.cCfg.AutoLog),
		fmt.Sprintf("%t", d.cCfg.UseMercury),
	)

	return contractAddr, nil
}

func (d *VerifiableLoadLogTriggerDeployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := verifiableLogTrigger.NewVerifiableLoadLogTriggerUpkeep(addr, deployer.Client)
	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}

type VerifiableLoadConditionalDeployable struct {
	contract *verifiableConditional.VerifiableLoadUpkeep
	cfg      *DeployerConfig
	cCfg     *VerifiableLoadConfig
}

func NewVerifiableLoadConditionalDeployable(
	cfg *DeployerConfig,
	cCfg *VerifiableLoadConfig,
) *VerifiableLoadConditionalDeployable {
	return &VerifiableLoadConditionalDeployable{
		cfg:  cfg,
		cCfg: cCfg,
	}
}

func (d *VerifiableLoadConditionalDeployable) Connect(
	ctx context.Context,
	addr string,
	deployer *Deployer,
) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *VerifiableLoadConditionalDeployable) Deploy(ctx context.Context, deployer *Deployer, config VerifyContractConfig) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	registrar := common.HexToAddress(d.cCfg.RegistrarAddr)

	contractAddr, trx, _, err := verifiableConditional.DeployVerifiableLoadUpkeep(
		opts, deployer.Client, registrar,
		d.cCfg.UseArbitrum,
	)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: Verifiable Load Contract creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	PrintVerifyContractCommand(
		config,
		contractAddr.String(),
		registrar.String(),
		fmt.Sprintf("%t", d.cCfg.UseArbitrum),
	)

	return contractAddr, nil
}

func (d *VerifiableLoadConditionalDeployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := verifiableConditional.NewVerifiableLoadUpkeep(addr, deployer.Client)
	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}
