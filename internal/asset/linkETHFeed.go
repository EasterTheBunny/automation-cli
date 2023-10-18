package asset

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/mock_ethlink_aggregator_wrapper"
)

type LinkETHFeedConfig struct {
	Answer uint64
}

type LinkETHFeedDeployable struct {
	contract *mock_ethlink_aggregator_wrapper.MockETHLINKAggregator
	config   *LinkETHFeedConfig
}

func NewLinkETHFeedDeployable(config *LinkETHFeedConfig) *LinkETHFeedDeployable {
	return &LinkETHFeedDeployable{
		config: config,
	}
}

func (d *LinkETHFeedDeployable) Connect(_ context.Context, addr string, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(common.HexToAddress(addr), deployer)
}

func (d *LinkETHFeedDeployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	contractAddr, trx, _, err := mock_ethlink_aggregator_wrapper.DeployMockETHLINKAggregator(
		opts,
		deployer.Client,
		new(big.Int).SetUint64(d.config.Answer),
	)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: LINK token creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	return contractAddr, nil
}

func (d *LinkETHFeedDeployable) connectToInterface(
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := mock_ethlink_aggregator_wrapper.NewMockETHLINKAggregator(
		addr,
		deployer.Client,
	)

	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.contract = contract

	return addr, nil
}
