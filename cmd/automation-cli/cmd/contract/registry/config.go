package registry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	cmdContext "github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/node"
)

var setConfigCmd = &cobra.Command{
	Use:   "set-config",
	Short: "Set the configuration for the registry",
	Long:  `Set the configuration for the registry including on-chain config and off-chain config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := cmdContext.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		keyConf := cmdContext.GetKeyConfigFromContext(cmd.Context())
		if keyConf == nil {
			return fmt.Errorf("missing private key config")
		}

		pkOverride, err := cmd.Flags().GetString("key")
		if err != nil {
			return err
		}

		dConfig = config.SetPrivateKey(dConfig, keyConf, pkOverride)

		paths := cmdContext.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		if err := setConfig(cmd.Context(), conf, deployer, paths.Environment); err != nil {
			return err
		}

		return nil
	},
}

func setConfig(ctx context.Context, conf *config.Config, deployer *asset.Deployer, env string) error {
	interactable := asset.NewRegistryV21Deployable(&asset.RegistryV21Config{
		Mode:            config.GetRegistryMode(conf.ServiceContract.Mode),
		LinkTokenAddr:   conf.LinkContract,
		LinkETHFeedAddr: conf.LinkETHFeed,
		FastGasFeedAddr: conf.FastGasFeed,
	})

	if _, err := interactable.Connect(ctx, conf.ServiceContract.RegistryAddress, deployer); err != nil {
		return err
	}

	nodeConfs := make([]asset.OCR2NodeConfig, len(conf.Nodes))

	for idx, nodeName := range conf.Nodes {
		nodeConfigPath := fmt.Sprintf("%s/%s", env, nodeName)

		nodeConf, _, err := config.GetNodeConfig(nodeConfigPath)
		if err != nil {
			return err
		}

		participantConf, err := node.GetParticipantInfo(ctx, nodeConf.ManagementURL)
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
		ctx,
		deployer,
		nodeConfs,
		config.GetAssetOCRConfig(conf),
		config.GetOffchainConfig(conf),
		config.GetOnchainConfig(conf),
	); err != nil {
		return err
	}

	return nil
}
