package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		// check if starts with ~/ and replace with home directory
		if strings.HasPrefix(configPath, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			configPath = strings.Replace(configPath, "~", home, 1)
		}

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

		config, err := GetConfig(configPath)
		if err != nil {
			return err
		}

		ctx = AttachConfig(ctx, *config)

		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		path := GetConfigPathFromContext(cmd.Context())
		if path == nil {
			return fmt.Errorf("missing config path in context")
		}

		return SaveConfig(*path)
	},
}

var configCmd = &cobra.Command{
	Use:   "set-config-var [NAME] [VALUE]",
	Short: "Shortcut to quickly update config var",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set(args[0], args[1])

		return nil
	},
}

func InitializeCommands() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(contractManagementCmd)
	rootCmd.AddCommand(networkManagementCmd)

	contractManagementCmd.AddCommand(contractConnectCmd)
	contractManagementCmd.AddCommand(contractDeployCmd)

	networkManagementCmd.AddCommand(networkAddCmd)

	_ = rootCmd.PersistentFlags().String(
		"state-directory",
		"~/.automation-cli",
		"directory to store cli configuration and persisted state",
	)

	_ = contractDeployCmd.Flags().
		String("mode", "DEFAULT", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")

	_ = networkAddCmd.Flags().Int8("count", 1, "total number of nodes to create with this configuration")
}

func Run() error {
	return rootCmd.Execute()
}
