package asset

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

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
	Config *DeployerConfig
	RPC    *rpc.Client
	Client *ethclient.Client

	privateKey *ecdsa.PrivateKey
	linkToken  *link.LinkToken
	addr       common.Address
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
	privateKey := parsePrivateKey(cfg.PrivateKey)

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
		privateKey: privateKey,
		linkToken:  linkToken,
		addr:       address,
	}, nil
}

func (d *Deployer) Connect(ctx context.Context, addr string, deployable Deployable) (string, error) {
	registryAddr, err := deployable.Connect(ctx, addr, d)

	return registryAddr.Hex(), err
}

func (d *Deployer) Deploy(ctx context.Context, deployable Deployable, config VerifyContractConfig) (string, error) {
	registryAddr, err := deployable.Deploy(ctx, d, config)

	return registryAddr.Hex(), err
}

func (d *Deployer) BuildTxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	nonce, err := d.Client.PendingNonceAt(ctx, d.addr)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: PendingNonceAt failure for address (%s): %s",
			ErrClientInteraction, d.addr.Hex(), err.Error())
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
	auth.Value = big.NewInt(0)        // in wei
	auth.GasLimit = d.Config.GasLimit // in units
	auth.GasPrice = gasPrice

	return auth, nil
}

func (d *Deployer) wait(ctx context.Context, trx *types.Transaction) error {
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
