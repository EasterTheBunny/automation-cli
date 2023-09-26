package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/internal/asset"
)

var contractManagementCmd = &cobra.Command{
	Use:   "contract [ACTION]",
	Short: "Manage contracts associated with the current network.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
}

var contractConnectCmd = &cobra.Command{
	Use:   "connect [ASSET] [ADDRESS]",
	Short: "Connect to existing asset at address",
	Long: `Connect to existing asset at address.
	
  Available Assets:
    registry - base set of contracts for an automation service`,
	ValidArgs: []string{
		"registrar",
		"registry",
		"verifiable-load-log-trigger",
		"verifiable-load-conditional",
	},
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		config, err := GetConfig(configPath)
		if err != nil {
			return err
		}

		dConfig := GetDeployerConfig(config)

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		switch args[0] {
		case "registrar":
			deployable := asset.NewV21RegistrarDeployable(&dConfig, &asset.RegistrarV21Config{
				RegistryAddr:  config.ServiceContract.RegistryAddress,
				LinkTokenAddr: config.LinkContract,
				MinLink:       0,
			})

			addr, err := deployer.Connect(cmd.Context(), args[1], deployable)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registrar_address", addr)
		case "registry":
			deployable := asset.NewV21RegistryDeployable(&dConfig, &asset.RegistryV21Config{
				Mode:            GetRegistryMode(config),
				LinkTokenAddr:   config.LinkContract,
				LinkETHFeedAddr: config.LinkETHFeed,
				FastGasFeedAddr: config.FastGasFeed,
			})

			addr, err := deployer.Connect(cmd.Context(), args[1], deployable)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registry_address", addr)
			viper.Set("service_contract.registry_version", dConfig.Version)
		case "verifiable-load-log-trigger":
			if config.ServiceContract.RegistrarAddress == "" || config.ServiceContract.RegistrarAddress == "0x" {
				return fmt.Errorf("no registrar deployed")
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&dConfig, &asset.VerifiableLoadConfig{
				RegistrarAddr: config.ServiceContract.RegistrarAddress,
				UseMercury:    config.LogTriggerLoadContract.UseMercury,
				UseArbitrum:   config.LogTriggerLoadContract.UseArbitrum,
				AutoLog:       config.LogTriggerLoadContract.AutoLog,
			})

			addr, err := deployer.Connect(cmd.Context(), args[1], deployable)
			if err != nil {
				return err
			}

			viper.Set("log_trigger_load_contract.contract_address", addr)
			viper.Set("log_trigger_load_contract.use_mercury", config.LogTriggerLoadContract.UseMercury)
			viper.Set("log_trigger_load_contract.use_arbitrum", config.LogTriggerLoadContract.UseArbitrum)
			viper.Set("log_trigger_load_contract.auto_log", config.LogTriggerLoadContract.AutoLog)
		case "verifiable-load-conditional":
			if config.ServiceContract.RegistrarAddress == "" || config.ServiceContract.RegistrarAddress == "0x" {
				return fmt.Errorf("no registrar deployed")
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&dConfig, &asset.VerifiableLoadConfig{
				RegistrarAddr: config.ServiceContract.RegistrarAddress,
				UseArbitrum:   config.ConditionalLoadContract.UseArbitrum,
			})

			addr, err := deployer.Connect(cmd.Context(), args[1], deployable)
			if err != nil {
				return err
			}

			viper.Set("conditional_load_contract.contract_address", addr)
			viper.Set("conditional_load_contract.use_arbitrum", config.ConditionalLoadContract.UseArbitrum)
		}

		return SaveConfig(configPath)
	},
}

var contractDeployCmd = &cobra.Command{
	Use:   "deploy [ASSET]",
	Short: "Deploy automation contract assets",
	Long: `Use this command to deploy automation related contracts. 

  Available Assets:
	registrar - contract to control upkeep registry
    registry - base set of contracts for an automation service
	verifiable-load-log-trigger - log trigger specific verifiable load contract
	verifiable-load-conditional - conditional trigger verifiable load contract`,
	ValidArgs: []string{
		"registrar",
		"registry",
		"verifiable-load-log-trigger",
		"verifiable-load-conditional",
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		config, err := GetConfig(configPath)
		if err != nil {
			return err
		}

		dConfig := GetDeployerConfig(config)

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		verifyConf := asset.VerifyContractConfig{
			ContractsDir:   config.Verifier.ContractsDir,
			NodeHTTPURL:    config.RPCHTTPURL,
			ExplorerAPIKey: config.Verifier.ExplorerAPIKey,
			NetworkName:    config.Verifier.NetworkName,
		}

		switch args[0] {
		case "registrar":
			deployable := asset.NewV21RegistrarDeployable(&dConfig, &asset.RegistrarV21Config{
				RegistryAddr:  config.ServiceContract.RegistryAddress,
				LinkTokenAddr: config.LinkContract,
				MinLink:       0,
			})

			addr, err := deployer.Deploy(cmd.Context(), deployable, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registrar_address", addr)
		case "registry":
			deployable := asset.NewV21RegistryDeployable(&dConfig, &asset.RegistryV21Config{
				Mode:            GetRegistryMode(config),
				LinkTokenAddr:   config.LinkContract,
				LinkETHFeedAddr: config.LinkETHFeed,
				FastGasFeedAddr: config.FastGasFeed,
			})

			addr, err := deployer.Deploy(cmd.Context(), deployable, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registry_address", addr)
			viper.Set("service_contract.registry_version", dConfig.Version)
		case "verifiable-load-log-trigger":
			if config.ServiceContract.RegistrarAddress == "" || config.ServiceContract.RegistrarAddress == "0x" {
				return fmt.Errorf("no registrar deployed")
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&dConfig, &asset.VerifiableLoadConfig{
				RegistrarAddr: config.ServiceContract.RegistrarAddress,
				UseMercury:    config.LogTriggerLoadContract.UseMercury,
				UseArbitrum:   config.LogTriggerLoadContract.UseArbitrum,
				AutoLog:       config.LogTriggerLoadContract.AutoLog,
			})

			addr, err := deployer.Deploy(cmd.Context(), deployable, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("log_trigger_load_contract.contract_address", addr)
			viper.Set("log_trigger_load_contract.use_mercury", config.LogTriggerLoadContract.UseMercury)
			viper.Set("log_trigger_load_contract.use_arbitrum", config.LogTriggerLoadContract.UseArbitrum)
			viper.Set("log_trigger_load_contract.auto_log", config.LogTriggerLoadContract.AutoLog)
		case "verifiable-load-conditional":
			if config.ServiceContract.RegistrarAddress == "" || config.ServiceContract.RegistrarAddress == "0x" {
				return fmt.Errorf("no registrar deployed")
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&dConfig, &asset.VerifiableLoadConfig{
				RegistrarAddr: config.ServiceContract.RegistrarAddress,
				UseArbitrum:   config.ConditionalLoadContract.UseArbitrum,
			})

			addr, err := deployer.Deploy(cmd.Context(), deployable, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("conditional_load_contract.contract_address", addr)
			viper.Set("conditional_load_contract.use_arbitrum", config.ConditionalLoadContract.UseArbitrum)
		}

		return SaveConfig(configPath)
	},
}
