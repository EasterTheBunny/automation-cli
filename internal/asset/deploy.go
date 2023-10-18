package asset

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	link "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/link_token_interface"
	bigmath "github.com/smartcontractkit/chainlink/v2/core/utils/big_math"

	"github.com/easterthebunny/automation-cli/internal/util"
)

var (
	ErrNetworkConnection      = fmt.Errorf("network connection failure")
	ErrPublicKeyCasting       = fmt.Errorf("error casting public key to ECDSA")
	ErrContractInitialization = fmt.Errorf("failed to initialize contract")
	ErrClientInteraction      = fmt.Errorf("eth client interaction")
	ErrChainTransaction       = fmt.Errorf("chain transaction failure")
	ErrVerifyContract         = fmt.Errorf("contract verification failure")
)

const (
	RegistryModeDefault  = 0
	RegistryModeArbitrum = 1
	RegistryModeOptimism = 2

	gasMultiplier int64 = 5
)

type Deployable interface {
	Connect(context.Context, string, *Deployer) (common.Address, error)
	Deploy(context.Context, *Deployer, VerifyContractConfig) (common.Address, error)
}

type DeployerConfig struct {
	Version      string
	RPCURL       string
	PrivateKey   string
	LinkContract string
	ChainID      int64
	GasLimit     uint64
}

type Deployer struct {
	Config  *DeployerConfig
	RPC     *rpc.Client
	Client  *ethclient.Client
	Address common.Address

	privateKey *ecdsa.PrivateKey
	linkToken  *link.LinkToken
}

// NewDeployer creates a new deployer and sets the primary address to
// the address associated with the configured private key.
func NewDeployer(cfg *DeployerConfig) (*Deployer, error) {
	// Created a client by the given node address
	rpcClient, err := rpc.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to connect to chain node (%s): %s", ErrNetworkConnection, cfg.RPCURL, err.Error())
	}

	nodeClient := ethclient.NewClient(rpcClient)
	privateKey := parsePrivateKey(strings.TrimSpace(cfg.PrivateKey))

	address, err := getAddressFromKey(privateKey)
	if err != nil {
		return nil, err
	}

	// Create link token wrapper
	linkToken, err := link.NewLinkToken(common.HexToAddress(cfg.LinkContract), nodeClient)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to connect to link token contract at (%s): %s",
			ErrContractInitialization, cfg.LinkContract, err.Error())
	}

	return &Deployer{
		Config:     cfg,
		RPC:        rpcClient,
		Client:     nodeClient,
		Address:    address,
		privateKey: privateKey,
		linkToken:  linkToken,
	}, nil
}

func (d *Deployer) BuildTxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	nonce, err := d.Client.PendingNonceAt(ctx, d.Address)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: PendingNonceAt failure for address (%s): %s",
			ErrClientInteraction, d.Address.Hex(), err.Error())
	}

	gasPrice, err := d.Client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: SuggestGasPrice failure: %s", ErrClientInteraction, err.Error())
	}

	gasPrice = bigmath.Add(gasPrice, bigmath.Div(gasPrice, big.NewInt(gasMultiplier))) // add 20%

	auth, err := bind.NewKeyedTransactorWithChainID(d.privateKey, big.NewInt(d.Config.ChainID))
	if err != nil {
		return nil, fmt.Errorf(
			"%w: NewKeyedTransactorWithChainID failed for chain id (%d): %s",
			ErrClientInteraction, d.Config.ChainID, err.Error())
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // in wei
	// auth.GasLimit = d.Config.GasLimit // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func (d *Deployer) Send(ctx context.Context, toAddr string, amount uint64) error {
	opts, err := d.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	addr := common.HexToAddress(toAddr)
	trx := types.NewTx(&types.LegacyTx{
		Nonce:    opts.Nonce.Uint64(),
		To:       &addr,
		Value:    new(big.Int).SetUint64(amount),
		Gas:      opts.GasLimit,
		GasPrice: opts.GasPrice,
		Data:     nil,
	})

	signedTx, err := types.SignTx(trx, types.NewEIP155Signer(big.NewInt(d.Config.ChainID)), d.privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	fmt.Println("attempting to send transaction: ", signedTx.Hash())
	if err = d.Client.SendTransaction(ctx, signedTx); err != nil {
		return fmt.Errorf("failed to send tx: %w", err)
	}

	return d.wait(ctx, signedTx)
}

func (d *Deployer) SendLINK(ctx context.Context, toAddr string, amount uint64) error {
	if amount == 0 {
		return nil
	}

	opts, err := d.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	trx, err := d.linkToken.Transfer(opts, common.HexToAddress(toAddr), new(big.Int).SetUint64(amount))
	if err != nil {
		return err
	}

	if err := d.wait(ctx, trx); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) ApproveLINK(ctx context.Context, toAddr string, amount uint64) error {
	if amount == 0 {
		return nil
	}

	opts, err := d.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	trx, err := d.linkToken.Approve(opts, common.HexToAddress(toAddr), new(big.Int).SetUint64(amount))
	if err != nil {
		return err
	}

	if err := d.wait(ctx, trx); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) wait(ctx context.Context, trx *types.Transaction) error {
	fmt.Println("waiting for transaction to be mined: ", trx.Hash())

	receipt, err := bind.WaitMined(ctx, d.Client, trx)
	if err != nil {
		return fmt.Errorf("%w: failed to wait for transaction (%s): %s", ErrChainTransaction, trx.Hash(), err.Error())
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("%w: %s: %s", ErrChainTransaction, util.ExplorerLink(d.Config.ChainID, trx.Hash()), err.Error())
	}

	return nil
}

func (d *Deployer) waitDeployment(ctx context.Context, trx *types.Transaction) error {
	if _, err := bind.WaitDeployed(ctx, d.Client, trx); err != nil {
		return fmt.Errorf(
			"%w: WaitDeployed failed %s: %s",
			ErrChainTransaction, util.ExplorerLink(d.Config.ChainID, trx.Hash()), err.Error())
	}

	return nil
}

func parsePrivateKey(encodedKey string) *ecdsa.PrivateKey {
	pkBase := new(big.Int).SetBytes(common.FromHex(encodedKey))
	pkX, pkY := crypto.S256().ScalarBaseMult(pkBase.Bytes())

	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: crypto.S256(),
			X:     pkX,
			Y:     pkY,
		},
		D: pkBase,
	}
}

func getAddressFromKey(key *ecdsa.PrivateKey) (common.Address, error) {
	publicKey := key.Public()

	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, ErrPublicKeyCasting
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}
