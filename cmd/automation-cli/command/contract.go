package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

const (
	Registrar                 = "registrar"
	Registry                  = "registry"
	VerifiableLoadLogTrigger  = "verifiable-load-log-trigger"
	VerifiableLoadConditional = "verifiable-load-conditional"
)

var (
	ErrRegistrarNotAvailable = fmt.Errorf("registrar not available")

	validContractNames = []string{
		Registrar,
		Registry,
		VerifiableLoadLogTrigger,
		VerifiableLoadConditional,
	}
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
	registrar - contract to control upkeep registry
    registry - base set of contracts for an automation service
	verifiable-load-log-trigger - log trigger specific verifiable load contract
	verifiable-load-conditional - conditional trigger verifiable load contract`,
	ValidArgs: validContractNames,
	Args:      cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		switch args[0] {
		case Registrar:
			registrar := asset.NewRegistrarV21Deployable(&asset.RegistrarV21Config{
				RegistryAddr:  conf.ServiceContract.RegistryAddress,
				LinkTokenAddr: conf.LinkContract,
				MinLink:       0,
			})

			addr, err := registrar.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registrar_address", addr)
		case Registry:
			registry := asset.NewRegistryV21Deployable(&asset.RegistryV21Config{
				Mode:            config.GetRegistryMode(conf),
				LinkTokenAddr:   conf.LinkContract,
				LinkETHFeedAddr: conf.LinkETHFeed,
				FastGasFeedAddr: conf.FastGasFeed,
			})

			addr, err := registry.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registry_address", addr)
			viper.Set("service_contract.registry_version", "v2.1")
		case VerifiableLoadLogTrigger:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return ErrRegistrarNotAvailable
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseMercury:    conf.LogTriggerLoadContract.UseMercury,
				UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
				AutoLog:       conf.LogTriggerLoadContract.AutoLog,
			})

			addr, err := deployable.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("log_trigger_load_contract.contract_address", addr)
			viper.Set("log_trigger_load_contract.use_mercury", conf.LogTriggerLoadContract.UseMercury)
			viper.Set("log_trigger_load_contract.use_arbitrum", conf.LogTriggerLoadContract.UseArbitrum)
			viper.Set("log_trigger_load_contract.auto_log", conf.LogTriggerLoadContract.AutoLog)
		case VerifiableLoadConditional:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return ErrRegistrarNotAvailable
			}

			deployable := asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
			})

			addr, err := deployable.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("conditional_load_contract.contract_address", addr)
			viper.Set("conditional_load_contract.use_arbitrum", conf.ConditionalLoadContract.UseArbitrum)
		}

		return nil
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
	ValidArgs: validContractNames,
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		verifyConf := asset.VerifyContractConfig{
			ContractsDir:   conf.Verifier.ContractsDir,
			NodeHTTPURL:    conf.RPCHTTPURL,
			ExplorerAPIKey: conf.Verifier.ExplorerAPIKey,
			NetworkName:    conf.Verifier.NetworkName,
		}

		switch args[0] {
		case Registrar:
			deployable := asset.NewRegistrarV21Deployable(&asset.RegistrarV21Config{
				RegistryAddr:  conf.ServiceContract.RegistryAddress,
				LinkTokenAddr: conf.LinkContract,
				MinLink:       0,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registrar_address", addr)
		case Registry:
			deployable := asset.NewRegistryV21Deployable(&asset.RegistryV21Config{
				Mode:            config.GetRegistryMode(conf),
				LinkTokenAddr:   conf.LinkContract,
				LinkETHFeedAddr: conf.LinkETHFeed,
				FastGasFeedAddr: conf.FastGasFeed,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("service_contract.registry_address", addr)
			viper.Set("service_contract.registry_version", "v2.1")
		case VerifiableLoadLogTrigger:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return ErrRegistrarNotAvailable
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseMercury:    conf.LogTriggerLoadContract.UseMercury,
				UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
				AutoLog:       conf.LogTriggerLoadContract.AutoLog,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("log_trigger_load_contract.contract_address", addr)
			viper.Set("log_trigger_load_contract.use_mercury", conf.LogTriggerLoadContract.UseMercury)
			viper.Set("log_trigger_load_contract.use_arbitrum", conf.LogTriggerLoadContract.UseArbitrum)
			viper.Set("log_trigger_load_contract.auto_log", conf.LogTriggerLoadContract.AutoLog)
		case VerifiableLoadConditional:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return ErrRegistrarNotAvailable
			}

			deployable := asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("conditional_load_contract.contract_address", addr)
			viper.Set("conditional_load_contract.use_arbitrum", conf.ConditionalLoadContract.UseArbitrum)
		}

		return nil
	},
}

var contractInteractCmd = &cobra.Command{
	Use:       "interact [NAME] [ACTION]",
	Short:     "Run pre-defined actions for contract",
	Long:      `Interact with contracts and run pre-packaged actions. This is not inclusive of all commands possible to run`,
	Args:      cobra.ExactArgs(2),
	ValidArgs: []string{VerifiableLoadConditional, "get-stats"},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		switch args[0] {
		case VerifiableLoadConditional:
			if args[1] != "get-stats" {
				return fmt.Errorf("invalid action")
			}

			interactable := asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
			})

			if err := interactable.ReadStats(cmd.Context(), deployer, asset.VerifiableLoadInteractionConfig{
				ContractAddr: conf.ConditionalLoadContract.ContractAddress,
			}); err != nil {
				return err
			}
		}

		return nil
	},
}
