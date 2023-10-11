package key

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print list of private key aliases",
	Long:  `List existing private key aliases`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := context.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		conf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		for _, key := range conf.Keys {
			fmt.Println(key.Alias)
		}

		return nil
	},
}
