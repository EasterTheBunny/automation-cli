package cmd

import (
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/call"
	"github.com/easterthebunny/automation-cli/cmd/configure"
	"github.com/easterthebunny/automation-cli/cmd/contract"
	"github.com/easterthebunny/automation-cli/cmd/key"
	"github.com/easterthebunny/automation-cli/cmd/network"
	"github.com/easterthebunny/automation-cli/internal/io"
)

func init() {
	rootCmd.AddCommand(configure.RootCmd)
	rootCmd.AddCommand(key.RootCmd)
	rootCmd.AddCommand(contract.RootCmd)
	rootCmd.AddCommand(network.RootCmd)

	rootCmd.AddCommand(call.RootCmd)

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
		env, err := getEnvironmentStatePath(cmd)
		if err != nil {
			return err
		}

		ctx := io.ContextWithEnvironment(cmd.Context(), env)

		cmd.SetContext(ctx)

		return nil
	},
}

func Run() error {
	return rootCmd.Execute()
}

func getEnvironmentStatePath(cmd *cobra.Command) (io.Environment, error) {
	rootPathOpt, err := cmd.Flags().GetString("state-directory")
	if err != nil {
		return io.Environment{}, err
	}

	selectedEnv, err := cmd.Flags().GetString("environment")
	if err != nil {
		return io.Environment{}, err
	}

	return io.Environment{
		Root: io.Root(rootPathOpt),
		Name: selectedEnv,
	}, nil
}
