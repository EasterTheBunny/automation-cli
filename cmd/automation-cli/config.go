package main

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/internal/asset"
)

var (
	ErrReadConfig  = fmt.Errorf("failed to read config")
	ErrWriteConfig = fmt.Errorf("failed to write config")
)

func GetConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("json")

	setViperDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			if err := viper.WriteConfigAs(path); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrWriteConfig, err.Error())
			}
		}
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("%w, %s", ErrReadConfig, err.Error())
	}

	return &config, nil
}

func SaveConfig(path string) error {
	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("%w: %s", ErrWriteConfig, err.Error())
	}

	return nil
}

func GetDeployerConfig(config *Config) asset.DeployerConfig {
	return asset.DeployerConfig{
		RPCURL:       config.RPCHTTPURL,
		ChainID:      config.ChainID,
		PrivateKey:   config.PrivateKey,
		LinkContract: config.LinkContract,
		Version:      config.ServiceContract.Version,
	}
}

func GetRegistryMode(config *Config) uint8 {
	switch config.ServiceContract.Mode {
	case "ARBITRUM":
		return 1
	case "OPTIMISM":
		return 2
	default:
		return 0
	}
}

func setViperDefaults() {
	viper.SetDefault("rpc_wss_url", "")
	viper.SetDefault("rpc_http_url", "")
	viper.SetDefault("chain_id", int64(1337))
	viper.SetDefault("private_key", "")
	viper.SetDefault("link_contract_address", "0x")
	viper.SetDefault("link_eth_feed", "0x")
	viper.SetDefault("fast_gas_feed", "0x")
	viper.SetDefault("bootstrap_address", "")
	viper.SetDefault("groupname", "")

	viper.SetDefault("service_contract.registry_version", "v2.1")
	viper.SetDefault("service_contract.registrar_address", "0x")
	viper.SetDefault("service_contract.registry_address", "0x")
	viper.SetDefault("service_contract.registry_mode", "DEFAULT")

	viper.SetDefault("log_trigger_load_contract.contract_address", "0x")
	viper.SetDefault("log_trigger_load_contract.use_mercury", false)
	viper.SetDefault("log_trigger_load_contract.use_arbitrum", false)
	viper.SetDefault("log_trigger_load_contract.auto_log", false)

	viper.SetDefault("conditional_load_contract.contract_address", "0x")
	viper.SetDefault("conditional_load_contract.use_arbitrum", false)

	viper.SetDefault("verifier.contracts_directory", "")
	viper.SetDefault("verifier.explorer_api_key", "")
	viper.SetDefault("verifier.network_name", "")
}

type Config struct {
	RPCWSSURL               string                  `mapstructure:"rpc_wss_url"`
	RPCHTTPURL              string                  `mapstructure:"rpc_http_url"`
	ChainID                 int64                   `mapstructure:"chain_id"`
	PrivateKey              string                  `mapstructure:"private_key"`
	LinkContract            string                  `mapstructure:"link_contract_address"`
	LinkETHFeed             string                  `mapstructure:"link_eth_feed"`
	FastGasFeed             string                  `mapstructure:"fast_gas_feed"`
	BootstrapAddress        string                  `mapstructure:"bootstrap_address"`
	Groupname               string                  `mapstructure:"groupname"`
	ServiceContract         ServiceContract         `mapstructure:"service_contract"`
	LogTriggerLoadContract  LogTriggerLoadContract  `mapstructure:"log_trigger_load_contract"`
	ConditionalLoadContract ConditionalLoadContract `mapstructure:"conditional_load_contract"`
	Nodes                   []Node                  `mapstructure:"nodes"`
	Verifier                Verifier                `mapstructure:"verifier"`
}

type ServiceContract struct {
	Version          string `mapstructure:"registry_version"`
	RegistrarAddress string `mapstructure:"registrar_address"`
	RegistryAddress  string `mapstructure:"registry_address"`
	Mode             string `mapstructure:"registry_mode"`
}

type LogTriggerLoadContract struct {
	ContractAddress string `mapstructure:"contract_address"`
	UseMercury      bool   `mapstructure:"use_mercury"`
	UseArbitrum     bool   `mapstructure:"use_arbitrum"`
	AutoLog         bool   `mapstructure:"auto_log"`
}

type ConditionalLoadContract struct {
	ContractAddress string `mapstructure:"contract_address"`
	UseArbitrum     bool   `mapstructure:"use_arbitrum"`
}

type Node struct {
	ChainlinkImage string `mapstructure:"chainlink_image"`
}

type Verifier struct {
	ContractsDir   string `mapstructure:"contracts_directory"`
	ExplorerAPIKey string `mapstructure:"explorer_api_key"`
	NetworkName    string `mapstructure:"network_name"`
}
