package asset

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/shared/generated/link_token"

	"github.com/easterthebunny/automation-cli/internal/config"
)

type LinkTokenDeployable struct {
	token  *link_token.LinkToken
	config *config.LinkTokenContract
}

func NewLinkTokenDeployable(conf *config.LinkTokenContract) *LinkTokenDeployable {
	return &LinkTokenDeployable{
		config: conf,
	}
}

func (d *LinkTokenDeployable) Connect(ctx context.Context, deployer *Deployer) (common.Address, error) {
	return d.connectToInterface(ctx, common.HexToAddress(d.config.Address), deployer)
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

	contractAddr, trx, _, err := link_token.DeployLinkToken(opts, deployer.Client)
	if err != nil {
		return contractAddr, fmt.Errorf("%w: LINK token creation failed: %s", ErrContractCreate, err.Error())
	}

	if err := deployer.waitDeployment(ctx, trx); err != nil {
		return contractAddr, err
	}

	d.config.Address = contractAddr.Hex()

	return contractAddr, nil
}

func (d *LinkTokenDeployable) Mint(
	ctx context.Context,
	deployer *Deployer,
	amount *big.Int,
) error {
	if _, err := d.connectToInterface(ctx, common.HexToAddress(d.config.Address), deployer); err != nil {
		return err
	}

	if err := ensureMintingRole(ctx, d.token, deployer); err != nil {
		return err
	}

	if err := mintAmountTo(ctx, d.token, deployer, amount, deployer.Address); err != nil {
		return err
	}

	return nil
}

func (d *LinkTokenDeployable) connectToInterface(
	_ context.Context,
	addr common.Address,
	deployer *Deployer,
) (common.Address, error) {
	contract, err := link_token.NewLinkToken(
		addr,
		deployer.Client,
	)

	if err != nil {
		return addr, fmt.Errorf("%w: failed to connect to contract at (%s): %s", ErrContractConnection, addr, err.Error())
	}

	d.token = contract

	return addr, nil
}

func ensureMintingRole(ctx context.Context, contract *link_token.LinkToken, deployer *Deployer) error {
	minters, err := contract.GetMinters(&bind.CallOpts{Context: ctx})
	if err != nil {
		return err
	}

	var ownerIsMinter bool

	for _, minter := range minters {
		if minter.Hex() == deployer.Address.Hex() {
			ownerIsMinter = true

			break
		}
	}

	if !ownerIsMinter {
		opts, err := deployer.BuildTxOpts(ctx)
		if err != nil {
			return err
		}

		trx, err := contract.GrantMintRole(opts, deployer.Address)
		if err != nil {
			return err
		}

		if err := deployer.wait(ctx, trx); err != nil {
			return err
		}
	}

	return nil
}

func mintAmountTo(
	ctx context.Context,
	contract *link_token.LinkToken,
	deployer *Deployer,
	amt *big.Int,
	to common.Address,
) error {
	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	trx, err := contract.Mint(opts, to, amt)
	if err != nil {
		return err
	}

	if err := deployer.wait(ctx, trx); err != nil {
		return err
	}

	return nil
}
