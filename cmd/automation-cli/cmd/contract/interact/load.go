package interact

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	cmdContext "github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/asset"
)

var (
	cancelUpkeeps bool
)

func init() {
	loadCmd.Flags().BoolVar(&cancelUpkeeps, "cancel-upkeeps", false, "cancel upkeeps before creating new ones")
}

var loadCmd = &cobra.Command{
	Use:   "verifiable-load [TYPE] [ACTION]",
	Short: "Run pre-defined actions on a verifiable load contract.",
	Long:  `Run pre-defined actions on a verifiable load contract.`,
	Example: `With an existing running environment and key (non-default) to do the following:

- cancel all existing upkeeps
- register new upkeeps
- collect and print delay statistics to console

$ automation-cli contract interact verifiable-load log-trigger register-upkeeps --environment="non.default" --key="mumbai-dev" --cancel-upkeeps
$ automation-cli contract interact verifiable-load log-trigger get-stats --environment="non.default" --key="mumbai-dev"

The above example assumes an environment set up with the name "non.default" and a verifiable load log trigger contract
deployed to that environment. Also, a key with the name "mumbai-dev" is used. The first command cancels all existing
upkeeps on the contract and deploys 5 (default count) new ones. The second command reads the captured delay data on the
contract and displays it to the console.

With the same setup above, to cancel all existing upkeeps only do the following:

$ automation-cli contract interact verifiable-load log-trigger cancel-upkeeps --environment="non.default" --key="mumbai-dev"

With the default environment and the same private key, leave off the "environment" and "key" parameters:

$ automation-cli contract interact verifiable-load log-trigger register-upkeeps --cancel-upkeeps
$ automation-cli contract interact verifiable-load log-trigger get-stats`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{"conditional", "log-trigger"}, cobra.ShellCompDirectiveNoFileComp
		}

		if len(args) == 1 {
			return []string{"get-stats", "register-upkeeps", "cancel-upkeeps"}, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := cmdContext.GetConfigFromContext(cmd.Context())
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

		keyConf := cmdContext.GetKeyConfigFromContext(cmd.Context())
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

		vlic := asset.VerifiableLoadInteractionConfig{
			ContractAddr:             conf.LogTriggerLoadContract.ContractAddress,
			RegisterUpkeepCount:      5,
			RegisteredUpkeepInterval: 15,
			CancelBeforeRegister:     cancelUpkeeps,
		}

		switch args[1] {
		case "get-stats":
			if err := runGetStats(cmd.Context(), args[0], conf, deployer, vlic); err != nil {
				return err
			}
		case "register-upkeeps":
			if err := runRegisterUpkeeps(cmd.Context(), args[0], conf, deployer, vlic); err != nil {
				return err
			}
		case "cancel-upkeeps":
			if err := runCancelUpkeeps(cmd.Context(), args[0], conf, deployer, vlic); err != nil {
				return err
			}
		}

		return nil
	},
}

type statsReader interface {
	ReadStats(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

type upkeepRegister interface {
	RegisterUpkeeps(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

type upkeepCanceller interface {
	CancelUpkeeps(context.Context, *asset.Deployer, asset.VerifiableLoadInteractionConfig) error
}

func runGetStats(
	ctx context.Context,
	contractType string,
	conf *config.Config,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var reader statsReader

	switch contractType {
	case "conditional":
		reader = asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.ConditionalLoadContract.ContractAddress
	case "log-trigger":
		reader = asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.LogTriggerLoadContract.ContractAddress
	}

	if err := reader.ReadStats(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}

func runRegisterUpkeeps(
	ctx context.Context,
	contractType string,
	conf *config.Config,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var register upkeepRegister

	switch contractType {
	case "conditional":
		register = asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.ConditionalLoadContract.ContractAddress
	case "log-trigger":
		register = asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.LogTriggerLoadContract.ContractAddress
	}

	if err := register.RegisterUpkeeps(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}

func runCancelUpkeeps(
	ctx context.Context,
	contractType string,
	conf *config.Config,
	deployer *asset.Deployer,
	vlic asset.VerifiableLoadInteractionConfig,
) error {
	var register upkeepCanceller

	switch contractType {
	case "conditional":
		register = asset.NewVerifiableLoadConditionalDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.ConditionalLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.ConditionalLoadContract.ContractAddress
	case "log-trigger":
		register = asset.NewVerifiableLoadLogTriggerDeployable(&asset.VerifiableLoadConfig{
			RegistrarAddr: conf.ServiceContract.RegistrarAddress,
			UseArbitrum:   conf.LogTriggerLoadContract.UseArbitrum,
		})

		vlic.ContractAddr = conf.LogTriggerLoadContract.ContractAddress
	}

	if err := register.CancelUpkeeps(ctx, deployer, vlic); err != nil {
		return err
	}

	return nil
}
