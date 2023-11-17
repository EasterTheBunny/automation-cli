package configure

import (
	"github.com/easterthebunny/automation-cli/internal/io"
	"github.com/spf13/cobra"
)

var (
	configDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete environment",
		Long:  `Delete environment configurations`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			env := io.EnvironmentFromContext(cmd.Context())
			if env != nil {
				return env.Delete()
			}

			return nil
		},
	}
)
