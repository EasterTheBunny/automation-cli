package key

import (
	"fmt"
	"io"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/easterthebunny/automation-cli/internal/config"
	cliio "github.com/easterthebunny/automation-cli/internal/io"
)

var (
	storeCmd = &cobra.Command{
		Use:   "store [NAME]",
		Short: "Store a private key with the reference name",
		Long:  `Securely store private keys under alias names for reference in configurations.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				pkBytes []byte
				err     error
			)

			if readStdIn {
				pkBytes, err = io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
			} else {
				fmt.Fprint(cmd.OutOrStdout(), "Enter private key: ")

				pkBytes, err = term.ReadPassword(syscall.Stdin)
				if err != nil {
					return err
				}
			}

			env := cliio.EnvironmentFromContext(cmd.Context())
			if env == nil {
				return fmt.Errorf("environment not found")
			}

			conf, err := config.ReadPrivateKeysFrom(env.Root.MustRead(config.PrivateKeyConfigFilename))
			if err != nil {
				return err
			}

			for idx, key := range conf.Keys {
				if key.Alias == args[0] {
					conf.Keys[idx].Value = string(pkBytes)

					return config.WritePrivateKeys(env.Root.MustWrite(config.PrivateKeyConfigFilename), conf)
				}
			}

			conf.Keys = append(conf.Keys, config.Key{
				Alias: args[0],
				Value: string(pkBytes),
			})

			return config.WritePrivateKeys(env.Root.MustWrite(config.PrivateKeyConfigFilename), conf)
		},
	}
)
