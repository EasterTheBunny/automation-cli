package interact

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/domain"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/node"
)

func init() {
	RootCmd.AddCommand(contractInteractRegistryCmd)
	RootCmd.AddCommand(contractInteractVerifiableLogCmd)
	RootCmd.AddCommand(contractInteractVerifiableCondCmd)

	_ = contractInteractVerifiableCondCmd.Flags().Bool("cancel-upkeeps", false, "cancel upkeeps before creating new ones")
}

var RootCmd = &cobra.Command{
	Use:       "interact [NAME] [ACTION]",
	Short:     "Run pre-defined actions for contract",
	Long:      `Interact with contracts and run pre-packaged actions. This is not inclusive of all commands possible to run`,
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: domain.ContractNames,
}

var contractInteractRegistryCmd = &cobra.Command{
	Use:       "registry [ACTION]",
	Short:     "Run pre-defined actions for contract",
	Long:      `Interact with the registry and run pre-packaged actions. This is not inclusive of all commands possible to run`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"set-config"},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := context.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		paths := context.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

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
		case "set-config":
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
				nodeConfigPath := fmt.Sprintf("%s/%s", paths.Environment, nodeName)

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
		}

		return nil
	},
}

var contractInteractVerifiableLogCmd = &cobra.Command{
	Use:       "verifiable-load-log-trigger [ACTION]",
	Short:     "Run pre-defined actions for contract",
	Long:      `Interact with the registry and run pre-packaged actions. This is not inclusive of all commands possible to run`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"get-stats", "register-upkeeps"},
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

		switch args[0] {
		case "get-stats":
			interactable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
			})

			if err := interactable.ReadStats(cmd.Context(), deployer, asset.VerifiableLoadInteractionConfig{
				ContractAddr: conf.LogTriggerLoadContract.ContractAddress,
			}); err != nil {
				return err
			}
		case "register-upkeeps":
			interactable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
			})

			if err := interactable.RegisterUpkeeps(cmd.Context(), deployer, asset.VerifiableLoadInteractionConfig{
				ContractAddr:             conf.LogTriggerLoadContract.ContractAddress,
				RegisterUpkeepCount:      5,
				RegisteredUpkeepInterval: 15,
				CancelBeforeRegister:     true,
			}); err != nil {
				return err
			}
		}

		return nil
	},
}

var contractInteractVerifiableCondCmd = &cobra.Command{
	Use:       "verifiable-load-conditional [ACTION]",
	Short:     "Run pre-defined actions for contract",
	Long:      `Interact with the registry and run pre-packaged actions. This is not inclusive of all commands possible to run`,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"get-stats", "register-upkeeps"},
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

		switch args[0] {
		case "get-stats":
			interactable := asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
			})

			if err := interactable.ReadStats(cmd.Context(), deployer, asset.VerifiableLoadInteractionConfig{
				ContractAddr: conf.ConditionalLoadContract.ContractAddress,
			}); err != nil {
				return err
			}
		case "register-upkeeps":
			interactable := asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
				RegistrarAddr: conf.ServiceContract.RegistrarAddress,
				UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
			})

			if err := interactable.RegisterUpkeeps(cmd.Context(), deployer, asset.VerifiableLoadInteractionConfig{
				ContractAddr:             conf.ConditionalLoadContract.ContractAddress,
				RegisterUpkeepCount:      5,
				RegisteredUpkeepInterval: 15,
				CancelBeforeRegister:     false,
			}); err != nil {
				return err
			}
		}

		return nil
	},
}
