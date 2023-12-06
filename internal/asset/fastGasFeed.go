package asset

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/mock_gas_aggregator_wrapper"

	"github.com/easterthebunny/automation-cli/internal/config"
)

type FastGasFeedDeployable struct {
	contract *mock_gas_aggregator_wrapper.MockGASAggregator
	config   *config.FeedContract
}

func NewFastGasFeedDeployable(conf *config.FeedContract) *FastGasFeedDeployable {
	return &FastGasFeedDeployable{
		config: conf,
	}
}

func (d *FastGasFeedDeployable) Connect(_ context.Context, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(common.HexToAddress(d.config.Address), deployer)
}

func (d *FastGasFeedDeployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	contractAddr, trx, _, err := mock_gas_aggregator_wrapper.DeployMockGASAggregator(
		opts,
		deployer.Client,
		new(big.Int).SetUint64(d.config.DefaultAnswer),
	)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: fast gas feed creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	d.config.Address = contractAddr.Hex()

	return contractAddr, nil
}

func (d *FastGasFeedDeployable) connectToInterface(
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := mock_gas_aggregator_wrapper.NewMockGASAggregator(
		addr,
		deployer.Client,
	)

	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}
