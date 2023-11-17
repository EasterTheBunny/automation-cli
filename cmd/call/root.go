package call

import (
	"context"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/v2/core/gethwrappers/shared/generated/link_token"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/config"
	cliio "github.com/easterthebunny/automation-cli/internal/io"
)

var RootCmd = &cobra.Command{
	Use:   "call",
	Short: "test",
	Long:  `test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := cliio.EnvironmentFromContext(cmd.Context())
		if path == nil {
			return fmt.Errorf("environment not found")
		}

		env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
		if err != nil {
			return err
		}

		keys, err := config.ReadPrivateKeysFrom(path.Root.MustRead(config.PrivateKeyConfigFilename))
		if err != nil {
			return err
		}

		key, err := keys.KeyForAlias(env.PrivateKeyAlias)
		if err != nil {
			return err
		}

		deployer, err := asset.NewDeployer(&env, key)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "-----> test transfer from")
		to, err := keys.KeyForAlias("ganache-1")
		if err != nil {
			return err
		}

		if err := approveLink(cmd.Context(), cmd.OutOrStdout(), to.Address, deployer, &env, 10); err != nil {
			return err
		}

		if err := receive(cmd.Context(), cmd.OutOrStdout(), key.Address, &keys, &env, 6); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "-----> END")
		fmt.Fprintln(cmd.OutOrStdout(), "")

		return nil
	},
}

/*
func accounts(ownerKeyAlias string, keys *config.PrivateKeys) (config.Key, config.Key, error) {
	var (
		owner   config.Key
		account config.Key
		err     error
	)

	owner, err = keys.KeyForAlias(ownerKeyAlias)
	if err != nil {
		return owner, account, err
	}

	account, err = keys.KeyForAlias("ganache-1")
	if err != nil {
		return owner, account, err
	}

	return owner, account, nil
}
*/

func approveLink(
	ctx context.Context,
	writer io.Writer,
	to string,
	deployer *asset.Deployer,
	env *config.Environment,
	amount int64,
) error {
	fmt.Fprintf(writer, "approving %d juels; from: %s; to: %s;\n", amount, deployer.Address.Hex(), to)

	// approve link
	linkContract, err := link_token.NewLinkToken(
		common.HexToAddress(env.LinkToken.Address),
		deployer.Client,
	)
	if err != nil {
		return err
	}

	/*
		opts, err := deployer.BuildTxOpts(ctx)
		if err != nil {
			return err
		}

		trx, err := linkContract.GrantMintRole(opts, deployer.Address)
		if err != nil {
			return err
		}

		if err = deployer.Wait(ctx, trx); err != nil {
			return err
		}

		opts, err = deployer.BuildTxOpts(ctx)
		if err != nil {
			return err
		}

		fmt.Fprintln(writer, "minting")
		trx, err = linkContract.Mint(opts, deployer.Address, big.NewInt(100_000_000_000_000))
		if err != nil {
			return err
		}

		if err = deployer.Wait(ctx, trx); err != nil {
			return err
		}
	*/

	opts, err := deployer.BuildTxOpts(ctx)
	if err != nil {
		return err
	}

	balance, err := linkContract.BalanceOf(&bind.CallOpts{Context: ctx}, deployer.Address)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "balance: %s juels\n", balance.String())

	trx, err := linkContract.Approve(opts, common.HexToAddress(to), big.NewInt(amount))
	if err != nil {
		return err
	}

	if err = deployer.Wait(ctx, trx); err != nil {
		return err
	}

	allowed, err := linkContract.Allowance(&bind.CallOpts{Context: ctx}, deployer.Address, common.HexToAddress(to))
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "allowed: %s juels\n", allowed.String())

	return nil
}

func receive(ctx context.Context, writer io.Writer, from string, keys *config.PrivateKeys, env *config.Environment, amount int64) error {
	key, err := keys.KeyForAlias("ganache-1")
	if err != nil {
		return err
	}

	deployer, err := asset.NewDeployer(env, key)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "receiving %d juels; from: %s; to: %s;\n", amount, from, deployer.Address.Hex())

	linkContract, err := link_token.NewLinkToken(
		common.HexToAddress(env.LinkToken.Address),
		deployer.Client,
	)
	if err != nil {
		return err
	}

	allowed, err := linkContract.Allowance(&bind.CallOpts{Context: ctx}, common.HexToAddress(from), deployer.Address)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "allowance: %s juels\n", allowed)

	/*
		opts, err := deployer.BuildTxOpts(ctx)
		if err != nil {
			return err
		}

		trx, err := linkContract.TransferFrom(opts, common.HexToAddress(from), deployer.Address, big.NewInt(amount))
		if err != nil {
			return err
		}

		if err = deployer.Wait(ctx, trx); err != nil {
			panic(err)
			//return err
		}
	*/

	balance, err := linkContract.BalanceOf(&bind.CallOpts{Context: ctx}, deployer.Address)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "balance: %s juels\n", balance.String())

	allowed, err = linkContract.Allowance(&bind.CallOpts{Context: ctx}, common.HexToAddress(from), deployer.Address)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "allowance: %s juels\n", allowed)

	return nil
}
