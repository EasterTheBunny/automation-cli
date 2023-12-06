package key

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	cliio "github.com/easterthebunny/automation-cli/internal/io"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [NAME]",
	Short: "Delete a private key with the reference name",
	Long:  `Delete private keys under alias names. Usage of '*' at the end will do a prefix match.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		env := cliio.EnvironmentFromContext(cmd.Context())
		if env == nil {
			return fmt.Errorf("environment not found")
		}

		conf, err := config.ReadPrivateKeysFrom(env.Root.MustRead(config.PrivateKeyConfigFilename))
		if err != nil {
			return err
		}

		matchType := 0
		match := args[0]
		if strings.HasSuffix(match, "*") {
			matchType = 1
			match = strings.ReplaceAll(match, "*", "")
		}

		for idx, key := range conf.Keys {
			if (matchType == 0 && key.Alias == match) || (matchType == 1 && strings.HasPrefix(key.Alias, match)) {
				fmt.Fprintf(cmd.OutOrStdout(), "key with alias '%s' was removed", args[0])

				conf.Keys = append(conf.Keys[:idx], conf.Keys[idx+1:]...)

				return config.WritePrivateKeys(env.Root.MustWrite(config.PrivateKeyConfigFilename), conf)
			}
		}

		return nil
	},
}
