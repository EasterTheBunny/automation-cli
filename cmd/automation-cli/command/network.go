package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/internal/node"
)

var networkManagementCmd = &cobra.Command{
	Use:   "network [ACTION] [TYPE]",
	Short: "Manage network components such as a bootstrap node and/or automation nodes",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
}

var networkAddCmd = &cobra.Command{
	Use:       "add [TYPE] [IMAGE]",
	Short:     "Create and add network components such as a bootstrap node and/or automation nodes",
	Long:      ``,
	ValidArgs: []string{"bootstrap", "participant"},
	Args:      cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		config, err := GetConfig(configPath)
		if err != nil {
			return err
		}

		switch args[0] {
		case "bootstrap":
			str, err := node.CreateBootstrapNode(cmd.Context(), node.NodeConfig{
				ChainID:     config.ChainID,
				NodeWSSURL:  config.RPCWSSURL,
				NodeHttpURL: config.RPCHTTPURL,
			}, "groupname", args[1], config.ServiceContract.RegistryAddress, 5688, 8000, false)
			if err != nil {
				return err
			}

			viper.Set("bootstrap_address", str)
		case "participant":
			count, err := cmd.Flags().GetInt8("count")
			if err != nil {
				return err
			}

			existing := len(config.Nodes)

			for idx := 0; idx < int(count); idx++ {
				// TODO: create participant node

				viper.Set(fmt.Sprintf("nodes.%d.chainlink_image", existing+idx), args[1])
			}
		default:
			return fmt.Errorf("unrecognized argument: %s", args[0])
		}

		return SaveConfig(configPath)
	},
}
