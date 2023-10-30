package configure

import (
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/spf13/cobra"
)

var (
	configDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete environment",
		Long:  `Delete environment configurations`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := context.GetPathsFromContext(cmd.Context())
			if paths == nil {
				return nil
			}

			if err := config.DeleteConfig(paths.Environment); err != nil {
				return err
			}

			ctx := context.RemoveConfig(cmd.Context())

			cmd.SetContext(ctx)

			return nil
		},
	}
)
