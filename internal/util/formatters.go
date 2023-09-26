package util

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// ExplorerLink creates a block explorer link for the given transaction hash. If the chain ID is
// unrecognized, the hash is returned as-is.
func ExplorerLink(chainID int64, txHash common.Hash) string {
	prefix := explorerLinkPrefix(chainID)
	if prefix != "" {
		return fmt.Sprintf("%s/tx/%s", prefix, txHash.String())
	}

	return txHash.String()
}

//nolint:funlen,cyclop
func explorerLinkPrefix(chainID int64) string {
	var prefix string

	switch chainID {
	case 1: // ETH mainnet
		prefix = "https://etherscan.io"
	case 4: // Rinkeby
		prefix = "https://rinkeby.etherscan.io"
	case 5: // Goerli
		prefix = "https://goerli.etherscan.io"
	case 42: // Kovan
		prefix = "https://kovan.etherscan.io"
	case 11155111: // Sepolia
		prefix = "https://sepolia.etherscan.io"

	case 420: // Optimism Goerli
		prefix = "https://goerli-optimism.etherscan.io"

	case ArbitrumGoerliChainID: // Arbitrum Goerli
		prefix = "https://goerli.arbiscan.io"
	case ArbitrumOneChainID: // Arbitrum mainnet
		prefix = "https://arbiscan.io"

	case 56: // BSC mainnet
		prefix = "https://bscscan.com"
	case 97: // BSC testnet
		prefix = "https://testnet.bscscan.com"

	case 137: // Polygon mainnet
		prefix = "https://polygonscan.com"
	case 80001: // Polygon Mumbai testnet
		prefix = "https://mumbai.polygonscan.com"

	case 250: // Fantom mainnet
		prefix = "https://ftmscan.com"
	case 4002: // Fantom testnet
		prefix = "https://testnet.ftmscan.com"

	case 43114: // Avalanche mainnet
		prefix = "https://snowtrace.io"
	case 43113: // Avalanche testnet
		prefix = "https://testnet.snowtrace.io"
	case 335: // Defi Kingdoms testnet
		prefix = "https://subnets-test.avax.network/defi-kingdoms"
	case 53935: // Defi Kingdoms mainnet
		prefix = "https://subnets.avax.network/defi-kingdoms"

	case 1666600000, 1666600001, 1666600002, 1666600003: // Harmony mainnet
		prefix = "https://explorer.harmony.one"
	case 1666700000, 1666700001, 1666700002, 1666700003: // Harmony testnet
		prefix = "https://explorer.testnet.harmony.one"

	case 84531:
		prefix = "https://goerli.basescan.org"
	case 8453:
		prefix = "https://basescan.org"

	default: // Unknown chain, return prefix as-is
		prefix = ""
	}

	return prefix
}
