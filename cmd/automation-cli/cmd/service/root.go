package service

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "service [ACTION]",
	Short: "Run mocked services",
	Long:  `Run mocked services`,
	Args:  cobra.MinimumNArgs(1),
}
