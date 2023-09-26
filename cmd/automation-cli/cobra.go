package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/node"
)

var rootCmd = &cobra.Command{
	Use:   "automation-cli",
	Short: "ChainLink Automation CLI tool to manage product assets",
	Long:  `automation-cli is a CLI for running the product management commands.`,
}

var configCmd = &cobra.Command{
	Use:   "set-config-var [NAME] [VALUE]",
	Short: "Shortcut to quickly update config var",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}

		_, err = GetConfig(configPath)
		if err != nil {
			return err
		}

		viper.Set(args[0], args[1])

		return SaveConfig(configPath)
	},
}

var attachCmd = &cobra.Command{
	Use:   "connect [ASSET] [ADDRESS]",
	Short: "Connect to existing asset at address",
	Long: `Connect to existing asset at address.
	
  Available Asset Options:
    registry - base set of contracts for an automation service`,
	ValidArgs: []string{"registrar", "registry", "verifiable-load-log-trigger", "verifiable-load-conditional"},
	Args:      cobra.ExactArgs(2),
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

var contractsManagementCmd = &cobra.Command{
	Use:   "deploy [ASSET]",
	Short: "Deploy all automation contract assets",
	Long: `Deploy automation contract assets.

  Available Options:
	registrar - contract to control upkeep registry
    registry - base set of contracts for an automation service
	verifiable-load-log-trigger - log trigger specific verifiable load contract
	verifiable-load-conditional - conditional trigger verifiable load contract`,
	ValidArgs: []string{"registrar", "registry", "verifiable-load-log-trigger", "verifiable-load-conditional"},
	Args:      cobra.ExactArgs(1),
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

func InitializeCommands() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(contractsManagementCmd)
	rootCmd.AddCommand(networkManagementCmd)

	networkManagementCmd.AddCommand(networkAddCmd)

	_ = rootCmd.PersistentFlags().String("config", "./config.json", "config file to store cli configuration and state")
	_ = contractsManagementCmd.Flags().
		String("mode", "DEFAULT", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")

	_ = networkAddCmd.Flags().Int8("count", 1, "total number of nodes to create with this configuration")
}
