package link

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var (
	setTokenCmd = &cobra.Command{
		Use:   "set-token-address [ADDRESS]",
		Short: "Set address for link token contract",
		Long:  "Set address for link token contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := io.EnvironmentFromContext(cmd.Context())
			if path == nil {
				return fmt.Errorf("environment not found")
			}

			env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
			if err != nil {
				return err
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("provided address must be hex encoded")
			}

			if env.LinkToken == nil {
				env.LinkToken = &config.LinkTokenContract{}
			}

			env.LinkToken.Address = args[0]

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}

	setLinkFeedCmd = &cobra.Command{
		Use:   "set-link-eth-feed-address [ADDRESS]",
		Short: "Set address for Link-ETH feed contract",
		Long:  "Set address for Link-ETH feed contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := io.EnvironmentFromContext(cmd.Context())
			if path == nil {
				return fmt.Errorf("environment not found")
			}

			env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
			if err != nil {
				return err
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("provided address must be hex encoded")
			}

			if env.LinkETH == nil {
				env.LinkETH = &config.FeedContract{}
			}

			env.LinkETH.Address = args[0]

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}

	setGasFeedCmd = &cobra.Command{
		Use:   "set-fast-gas-feed-address [ADDRESS]",
		Short: "Set address for fast gas feed contract",
		Long:  "Set address for fast gas feed contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := io.EnvironmentFromContext(cmd.Context())
			if path == nil {
				return fmt.Errorf("environment not found")
			}

			env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
			if err != nil {
				return err
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("provided address must be hex encoded")
			}

			if env.FastGas == nil {
				env.FastGas = &config.FeedContract{}
			}

			env.FastGas.Address = args[0]

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
