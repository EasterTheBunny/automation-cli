package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

		// check if starts with ~/ and replace with home directory
		if strings.HasPrefix(configPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			configPath = strings.Replace(configPath, "~", home, 1)
		}

		privateKeyPath := configPath
		configPath = fmt.Sprintf("%s/%s", configPath, env)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			abs, err := filepath.Abs(configPath)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "creating absolute path: %s\n", abs)
			if err := os.MkdirAll(abs, 0760); err != nil {
				return err
			}
		}

		ctx := AttachConfigPath(cmd.Context(), configPath)

		conf, err := config.GetConfig(configPath)
		if err != nil {
			return err
		}

		ctx = AttachConfig(ctx, *conf)

		keyConf, err := config.GetPrivateKeyConfig(privateKeyPath)
		if err != nil {
			return err
		}

		ctx = AttachKeyConfig(ctx, *keyConf)

		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		path := GetConfigPathFromContext(cmd.Context())
		if path == nil {
			return fmt.Errorf("missing config path in context")
		}

		return config.SaveConfig(*path)
	},
}

func InitializeCommands() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(contractManagementCmd)
	rootCmd.AddCommand(networkManagementCmd)

	contractManagementCmd.AddCommand(contractConnectCmd)
	contractManagementCmd.AddCommand(contractDeployCmd)
	contractManagementCmd.AddCommand(contractInteractCmd)

	networkManagementCmd.AddCommand(networkAddCmd)

	configCmd.AddCommand(configSetVarCmd)
	configCmd.AddCommand(configGetVarCmd)
	configCmd.AddCommand(configSetupCmd)
	configCmd.AddCommand(configStorePKCmd)

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
		"with-private-key",
		"default",
		"use a specific private key. use only an alias for a previously saved private key.",
	)
}

func Run() error {
	return rootCmd.Execute()
}
