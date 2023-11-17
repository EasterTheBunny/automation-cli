package key

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print list of private key aliases",
	Long:  `List existing private key aliases`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env := io.EnvironmentFromContext(cmd.Context())
		if env == nil {
			return fmt.Errorf("environment not found")
		}

		conf, err := config.ReadPrivateKeysFrom(env.Root.MustRead(config.PrivateKeyConfigFilename))
		if err != nil {
			return err
		}

		for _, key := range conf.Keys {
			fmt.Println(key.Alias)
		}

		return nil
	},
}
