package cmd

import (
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/configure"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/contract"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/key"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/network"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

func init() {
	rootCmd.AddCommand(configure.RootCmd)
	rootCmd.AddCommand(key.RootCmd)
	rootCmd.AddCommand(contract.RootCmd)
	rootCmd.AddCommand(network.RootCmd)

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

	_ = rootCmd.PersistentFlags().String(
		"key",
		"",
		"use to override configured state private key for command",
	)
}

var rootCmd = &cobra.Command{
	Use:     "automation-cli",
	Short:   "ChainLink Automation CLI tool to manage product assets",
	Long:    `automation-cli is a CLI for running the product management commands.`,
	Version: "v2.1",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("state-directory")
		if err != nil {
			return err
		}

		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}

		paths, err := context.CreateStatePaths(configPath, env)
		if err != nil {
			return err
		}

		ctx := context.AttachPaths(cmd.Context(), paths)

		conf, err := config.GetConfig(paths.Environment)
		if err != nil {
			return err
		}

		ctx = context.AttachConfig(ctx, *conf)

		keyConf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		ctx = context.AttachKeyConfig(ctx, *keyConf)

		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		paths := context.GetPathsFromContext(cmd.Context())
		if paths != nil {
			if conf := context.GetConfigFromContext(cmd.Context()); conf != nil {
				return config.SaveConfig(paths.Environment)
			}
		}

		return nil
	},
}

func Run() error {
	return rootCmd.Execute()
}
