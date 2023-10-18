package asset

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/link_token_interface"
)

type LinkTokenDeployable struct {
	token *link_token_interface.LinkToken
}

func NewLinkTokenDeployable() *LinkTokenDeployable {
	return &LinkTokenDeployable{}
}

func (d *LinkTokenDeployable) Connect(ctx context.Context, addr string, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(addr), deployer)
}

func (d *LinkTokenDeployable) Deploy(
	ctx context.Context,
	deployer *Deployer,
) (common.Address, error) {
	var contractAddr common.Address

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: deploy failed: %s", ErrContractCreate, err.Error())
	}

	contractAddr, trx, _, err := link_token_interface.DeployLinkToken(opts, deployer.Client)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: LINK token creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	return contractAddr, nil
}

func (d *LinkTokenDeployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := link_token_interface.NewLinkToken(
		addr,
		deployer.Client,
	)

	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.token = contract

	return addr, nil
}
