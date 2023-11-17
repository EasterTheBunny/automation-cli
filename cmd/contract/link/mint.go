package link

import (
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/util"
	"github.com/spf13/cobra"
)

var (
	mintTokenCmd = &cobra.Command{
		Use:   "mint [AMOUNT]",
		Short: "Mint new link token.",
		Long:  "Mint new link token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			amount, err := util.ParseExp(args[0])
			if err != nil {
				return err
			}

			deployer, err := asset.NewDeployer(&env, key)
			if err != nil {
				return err
			}

			deployable := asset.NewLinkTokenDeployable(env.LinkToken)

			if err := deployable.Mint(cmd.Context(), deployer, amount); err != nil {
				return err
			}

			return nil
		},
	}
)
