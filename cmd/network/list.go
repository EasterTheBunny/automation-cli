package network

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodes on the current network configuration",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := io.EnvironmentFromContext(cmd.Context())
		if path == nil {
			return fmt.Errorf("environment not found")
		}

		env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
		if err != nil {
			return err
		}

		for _, node := range env.Participants {
			fmt.Fprintf(cmd.OutOrStdout(), "%s-%s\n", env.Groupname, node.Name)
		}

		return nil
	},
}
