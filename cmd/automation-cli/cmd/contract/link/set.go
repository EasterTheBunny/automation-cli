package link

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	setTokenCmd = &cobra.Command{
		Use:   "set-token-address [ADDRESS]",
		Short: "Create new LINK token contract and add to environment",
		Long:  `Create new LINK token contract and add to environment.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := common.Address([]byte(args[0]))

			viper.Set("link_contract_address", addr)

			return nil
		},
	}

	setFeedCmd = &cobra.Command{
		Use:   "set-feed-address [ADDRESS]",
		Short: "Create new mock LINK-ETH feed contract",
		Long:  `Create new mock LINK-ETH feed contract. The resulting contract always returns the configured amount.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := common.Address([]byte(args[0]))

			viper.Set("link_eth_feed", addr)

			return nil
		},
	}
)
