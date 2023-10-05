package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
)

var rootCmd = &cobra.Command{
	Use:   "automation-cli",
	Short: "ChainLink Automation CLI tool to manage product assets",
	Long:  `automation-cli is a CLI for running the product management commands.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("state-directory")
		if err != nil {
			return err
		}

		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}

		paths, err := CreateStatePaths(configPath, env)
		if err != nil {
			return err
		}

		ctx := AttachPaths(cmd.Context(), *paths)

		conf, err := config.GetConfig(paths.Environment)
		if err != nil {
			return err
		}

		ctx = AttachConfig(ctx, *conf)

		keyConf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		ctx = AttachKeyConfig(ctx, *keyConf)

		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		paths := GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		return config.SaveConfig(paths.Environment)
	},
}

func InitializeCommands() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(contractManagementCmd)
	rootCmd.AddCommand(networkManagementCmd)

	contractManagementCmd.AddCommand(contractConnectCmd)
	contractManagementCmd.AddCommand(contractDeployCmd)
	contractManagementCmd.AddCommand(contractInteractCmd)

	contractInteractCmd.AddCommand(contractInteractRegistryCmd)
	contractInteractCmd.AddCommand(contractInteractVerifiableLogCmd)
	contractInteractCmd.AddCommand(contractInteractVerifiableCondCmd)

	networkManagementCmd.AddCommand(networkAddCmd)
	networkManagementCmd.AddCommand(networkListCmd)
	networkManagementCmd.AddCommand(networkFundCmd)

	configCmd.AddCommand(configSetVarCmd)
	configCmd.AddCommand(configGetVarCmd)
	configCmd.AddCommand(configSetupCmd)
	configCmd.AddCommand(configStorePKCmd)
	configCmd.AddCommand(configCreatePKCmd)
	configCmd.AddCommand(configListPKCmd)

	_ = rootCmd.PersistentFlags().String(
		"state-directory",
		"~/.automation-cli",
		"directory to store cli configuration and persisted state",
	)

	_ = rootCmd.PersistentFlags().String(
		"environment",
		"default",
		"scope for cli configuration and persisted state",
	)

	_ = contractDeployCmd.Flags().
		String("mode", "DEFAULT", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")

	_ = networkAddCmd.Flags().Int8("count", 1, "total number of nodes to create with this configuration")
	_ = networkAddCmd.Flags().String(
		"private-key",
		"default",
		"use a specific private key. use only an alias for a previously saved private key.",
	)
	_ = networkAddCmd.Flags().String(
		"log-level",
		"error",
		"set the log level for the node",
	)
}

func Run() error {
	return rootCmd.Execute()
}
