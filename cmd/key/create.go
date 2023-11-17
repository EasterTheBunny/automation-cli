package key

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	cliio "github.com/easterthebunny/automation-cli/internal/io"
)

var (
	createCmd = &cobra.Command{
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

			env := cliio.EnvironmentFromContext(cmd.Context())
			if env == nil {
				return fmt.Errorf("environment not found")
			}

			conf, err := config.ReadPrivateKeysFrom(env.Root.MustRead(config.PrivateKeyConfigFilename))
			if err != nil {
				return err
			}

			conf.Keys = append(conf.Keys, config.Key{
				Alias:   args[0],
				Value:   hexutil.Encode(keyBytes)[2:],
				Address: crypto.PubkeyToAddress(privateKey.PublicKey).Hex(),
			})

			return config.WritePrivateKeys(env.Root.MustWrite(config.PrivateKeyConfigFilename), conf)
		},
	}
)
