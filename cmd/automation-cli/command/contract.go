package command

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/node"
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

		keyConf := GetKeyConfigFromContext(cmd.Context())
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

		keyConf := GetKeyConfigFromContext(cmd.Context())
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

		path := GetConfigPathFromContext(cmd.Context())
		if path == nil {
			return fmt.Errorf("missing config path in context")
		}

		keyConf := GetKeyConfigFromContext(cmd.Context())
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
		case Registry:
			if args[1] != "set-config" {
				return fmt.Errorf("invalid action")
			}

			interactable := asset.NewRegistryV21Deployable(&asset.RegistryV21Config{
				Mode:            config.GetRegistryMode(conf),
				LinkTokenAddr:   conf.LinkContract,
				LinkETHFeedAddr: conf.LinkETHFeed,
				FastGasFeedAddr: conf.FastGasFeed,
			})

			if _, err := interactable.Connect(cmd.Context(), conf.ServiceContract.RegistryAddress, deployer); err != nil {
				return err
			}

			nodeConfs := make([]asset.OCR2NodeConfig, len(conf.Nodes))

			for idx, nodeName := range conf.Nodes {
				nodeConfigPath := fmt.Sprintf("%s/%s", *path, nodeName)

				nodeConf, _, err := config.GetNodeConfig(nodeConfigPath)
				if err != nil {
					return err
				}

				participantConf, err := node.GetParticipantInfo(cmd.Context(), nodeConf.ManagementURL)
				if err != nil {
					return fmt.Errorf("failed to get participant info from %s: %s", nodeName, err.Error())
				}

				nodeConfs[idx] = asset.OCR2NodeConfig{
					Address:           nodeConf.Address,
					OffChainPublicKey: participantConf.OffChainPublicKey,
					ConfigPublicKey:   participantConf.ConfigPublicKey,
					OnchainPublicKey:  participantConf.OnchainPublicKey,
					P2PKeyID:          participantConf.P2PKeyID,
				}
			}

			if err := interactable.SetOffchainConfig(
				cmd.Context(),
				deployer,
				nodeConfs,
				config.GetAssetOCRConfig(conf),
				config.GetOffchainConfig(conf),
				config.GetOnchainConfig(conf),
			); err != nil {
				return err
			}
		case VerifiableLoadLogTrigger:
			if args[1] != "get-stats" {
				return fmt.Errorf("invalid action")
			}

			interactable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
			})

			if err := interactable.ReadStats(cmd.Context(), deployer, asset.VerifiableLoadInteractionConfig{
				ContractAddr: conf.ConditionalLoadContract.ContractAddress,
			}); err != nil {
				return err
			}
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
