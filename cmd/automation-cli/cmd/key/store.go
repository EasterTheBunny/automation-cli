package key

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

var storeCmd = &cobra.Command{
	Use:   "store [NAME]",
	Short: "Store a private key with the reference name",
	Long:  `Securely store private keys under alias names for reference in configurations.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(cmd.OutOrStdout(), "Enter private key: ")

		pkBytes, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return err
		}

		paths := context.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		conf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		for idx, key := range conf.Keys {
			if key.Alias == args[0] {
				conf.Keys[idx].Value = string(pkBytes)

				return config.SavePrivateKeyConfig(paths.Base, conf)
			}
		}

		conf.Keys = append(conf.Keys, config.Key{
			Alias: args[0],
			Value: string(pkBytes),
		})

		return config.SavePrivateKeyConfig(paths.Base, conf)
	},
}