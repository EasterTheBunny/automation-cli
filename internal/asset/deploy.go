package asset

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	link "github.com/smartcontractkit/chainlink/v2/core/gethwrappers/generated/link_token_interface"
	bigmath "github.com/smartcontractkit/chainlink/v2/core/utils/big_math"

	"github.com/easterthebunny/automation-cli/internal/config"
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

	gasMultiplier        int64  = 5
	defaultConfirmations uint16 = 10
)

type Deployable interface {
	Connect(context.Context, string, *Deployer) (common.Address, error)
	Deploy(context.Context, *Deployer, VerifyContractConfig) (common.Address, error)
}

type Deployer struct {
	Keys    config.PrivateKeys
	Config  *config.Environment
	RPC     *rpc.Client
	Client  *ethclient.Client
	Address common.Address

	privateKey *ecdsa.PrivateKey
	linkToken  *link.LinkToken
}

// NewDeployer creates a new deployer and sets the primary address to
// the address associated with the configured private key.
func NewDeployer(cfg *config.Environment, key config.Key) (*Deployer, error) {
	// Created a client by the given node address
	rpcClient, err := rpc.Dial(cfg.HTTPURL)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to connect to chain node (%s): %s", ErrNetworkConnection, cfg.HTTPURL, err.Error())
	}

	nodeClient := ethclient.NewClient(rpcClient)
	privateKey := parsePrivateKey(strings.TrimSpace(key.Value))

	address, err := getAddressFromKey(privateKey)
	if err != nil {
		return nil, err
	}

	var token *link.LinkToken

	if cfg.LinkToken != nil {
		// Create link token wrapper
		token, err = link.NewLinkToken(common.HexToAddress(cfg.LinkToken.Address), nodeClient)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: failed to connect to link token contract at (%s): %s",
				ErrContractInitialization, cfg.LinkToken.Address, err.Error())
		}
	}

	return &Deployer{
		Config:     cfg,
		RPC:        rpcClient,
		Client:     nodeClient,
		Address:    address,
		privateKey: privateKey,
		linkToken:  token,
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
	// auth.GasLimit = config.DefaultDeployerGasLimit
	auth.GasPrice = gasPrice

	return auth, nil
}

func (d *Deployer) Send(ctx context.Context, toAddr string, amount *big.Int) error {
	opts, err := d.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	addr := common.HexToAddress(toAddr)
	trx := types.NewTx(&types.LegacyTx{
		Nonce:    opts.Nonce.Uint64(),
		To:       &addr,
		Value:    amount,
		Gas:      50_000,
		GasPrice: opts.GasPrice,
		Data:     nil,
	})

	signedTx, err := types.SignTx(trx, types.NewEIP155Signer(big.NewInt(d.Config.ChainID)), d.privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

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

func (d *Deployer) BalanceLINK(ctx context.Context, addr string) (*big.Int, error) {
	return d.linkToken.BalanceOf(&bind.CallOpts{
		Context: ctx,
	}, common.HexToAddress(addr))
}

func (d *Deployer) BlockDetail(ctx context.Context) (json.RawMessage, error) {
	var message json.RawMessage
	err := d.RPC.CallContext(ctx, &message, "eth_getBlockByNumber", "latest", false)

	return message, err
}

func (d *Deployer) Wait(ctx context.Context, trx *types.Transaction) error {
	return d.wait(ctx, trx)
}

func (d *Deployer) wait(ctx context.Context, trx *types.Transaction) error {
	fmt.Println("waiting for transaction to be mined: ", trx.Hash())

	receipt, err := bind.WaitMined(ctx, d.Client, trx)
	if err != nil {
		return fmt.Errorf("%w: failed to wait for transaction (%s): %s", ErrChainTransaction, trx.Hash(), err.Error())
	}

	if receipt.Status == types.ReceiptStatusFailed {
		var link string

		if trx != nil {
			link = util.ExplorerLink(d.Config.ChainID, trx.Hash())
		}

		var errStr string

		if err != nil {
			errStr = err.Error()
		}

		return fmt.Errorf("%w: %s: %s", ErrChainTransaction, link, errStr)
	}

	if err := waitConfirmations(ctx, d.Client, receipt, defaultConfirmations); err != nil {
		return err
	}

	return nil
}

func (d *Deployer) WaitDeployment(ctx context.Context, trx *types.Transaction) error {
	return d.waitDeployment(ctx, trx)
}

func (d *Deployer) waitDeployment(ctx context.Context, trx *types.Transaction) error {
	if trx.To() != nil {
		return errors.New("tx is not contract creation")
	}

	receipt, err := bind.WaitMined(ctx, d.Client, trx)
	if err != nil {
		return fmt.Errorf("%w: failed to wait for transaction (%s): %s", ErrChainTransaction, trx.Hash(), err.Error())
	}

	if receipt.ContractAddress == (common.Address{}) {
		return errors.New("zero address")
	}

	if err := waitConfirmations(ctx, d.Client, receipt, defaultConfirmations); err != nil {
		return err
	}

	// Check that code has indeed been deployed at the address.
	// This matters on pre-Homestead chains: OOG in the constructor
	// could leave an empty account behind.
	code, err := d.Client.CodeAt(ctx, receipt.ContractAddress, nil)
	if err == nil && len(code) == 0 {
		err = bind.ErrNoCodeAfterDeploy
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

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *jsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func waitConfirmations(ctx context.Context, client *ethclient.Client, receipt *types.Receipt, confs uint16) error {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	for {
		block, err := client.BlockNumber(ctx)
		if err != nil {
			return err
		}

		if block-receipt.BlockNumber.Uint64() >= uint64(confs) {
			return nil
		}

		// Wait for the next round.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-queryTicker.C:
		}
	}
}
