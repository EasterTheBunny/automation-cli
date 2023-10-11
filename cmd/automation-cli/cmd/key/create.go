package key

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

var createCmd = &cobra.Command{
	Use:   "create [NAME]",
	Short: "Create a private key with the reference name",
	Long:  `Create a new private key and store it under the provided reference name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		keyBytes := crypto.FromECDSA(privateKey)

		paths := context.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		conf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		conf.Keys = append(conf.Keys, config.Key{
			Alias: args[0],
			Value: hexutil.Encode(keyBytes),
		})

		return config.SavePrivateKeyConfig(paths.Base, conf)
	},
}
