package contract

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/cmd/contract/interact"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/domain"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

func init() {
	RootCmd.AddCommand(contractConnectCmd)
	RootCmd.AddCommand(contractDeployCmd)
	RootCmd.AddCommand(interact.RootCmd)

	_ = contractDeployCmd.Flags().
		String("mode", "DEFAULT", "registry mode (applies to v2.x; valid options are DEFAULT, ARBITRUM, OPTIMISM)")
}

var RootCmd = &cobra.Command{
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
	verifiable-load-conditional - conditional trigger verifiable load contract
	link-token - LINK token contract
	link-eth-feed - LINK-ETH price feed`,
	ValidArgs: domain.ContractNames,
	Args:      cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := context.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		keyConf := context.GetKeyConfigFromContext(cmd.Context())
		if keyConf == nil {
			return fmt.Errorf("missing private key config")
		}

		for _, key := range keyConf.Keys {
			if key.Alias == dConfig.PrivateKey {
				dConfig.PrivateKey = key.Value

				break
			}
		}

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		switch args[0] {
		case domain.Registrar:
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
		case domain.Registry:
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
		case domain.VerifiableLoadLogTrigger:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return domain.ErrRegistrarNotAvailable
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseMercury:    conf.LogTriggerLoadContract.UseMercury,
				UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
			})

			addr, err := deployable.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("log_trigger_load_contract.contract_address", addr)
			viper.Set("log_trigger_load_contract.use_mercury", conf.LogTriggerLoadContract.UseMercury)
			viper.Set("log_trigger_load_contract.use_arbitrum", conf.LogTriggerLoadContract.UseArbitrum)
		case domain.VerifiableLoadConditional:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return domain.ErrRegistrarNotAvailable
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
		case domain.LinkToken:
			deployable := asset.NewLinkTokenDeployable()

			addr, err := deployable.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("link_contract_address", addr)
		case domain.LinkEthFeed:
			deployable := asset.NewLinkETHFeedDeployable(&asset.LinkETHFeedConfig{
				Answer: 2e18,
			})

			addr, err := deployable.Connect(cmd.Context(), args[1], deployer)
			if err != nil {
				return err
			}

			viper.Set("link_eth_feed", addr)
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
	verifiable-load-conditional - conditional trigger verifiable load contract
	link-token - LINK token contract`,
	ValidArgs: domain.ContractNames,
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := context.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)
		selectedPK := dConfig.PrivateKey

		keyOverride, err := cmd.Flags().GetString("key")
		if err != nil {
			return err
		}

		if keyOverride != "" {
			selectedPK = keyOverride
		}

		keyConf := context.GetKeyConfigFromContext(cmd.Context())
		if keyConf == nil {
			return fmt.Errorf("missing private key config")
		}

		for _, key := range keyConf.Keys {
			if key.Alias == selectedPK {
				dConfig.PrivateKey = key.Value

				break
			}
		}

		if dConfig.PrivateKey == "" {
			return fmt.Errorf("private key alias not found")
		}

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
		case domain.Registrar:
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
		case domain.Registry:
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
		case domain.VerifiableLoadLogTrigger:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return domain.ErrRegistrarNotAvailable
			}

			deployable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseMercury:    conf.LogTriggerLoadContract.UseMercury,
				UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer, verifyConf)
			if err != nil {
				return err
			}

			viper.Set("log_trigger_load_contract.contract_address", addr)
			viper.Set("log_trigger_load_contract.use_mercury", conf.LogTriggerLoadContract.UseMercury)
			viper.Set("log_trigger_load_contract.use_arbitrum", conf.LogTriggerLoadContract.UseArbitrum)
		case domain.VerifiableLoadConditional:
			if conf.ServiceContract.RegistrarAddress == "" || conf.ServiceContract.RegistrarAddress == "0x" {
				return domain.ErrRegistrarNotAvailable
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
		case domain.LinkToken:
			deployable := asset.NewLinkTokenDeployable()

			addr, err := deployable.Deploy(cmd.Context(), deployer)
			if err != nil {
				return err
			}

			viper.Set("link_contract_address", addr)
		case domain.LinkEthFeed:
			deployable := asset.NewLinkETHFeedDeployable(&asset.LinkETHFeedConfig{
				Answer: 2e18,
			})

			addr, err := deployable.Deploy(cmd.Context(), deployer)
			if err != nil {
				return err
			}

			viper.Set("link_eth_feed", addr)
		}

		return nil
	},
}
