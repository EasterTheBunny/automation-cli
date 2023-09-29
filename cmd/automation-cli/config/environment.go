package config

import (
	"time"

	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/spf13/viper"
)

func GetConfig(path string) (*Config, error) {
	configPath, err := ensureExists(path, "config.json")
	if err != nil {
		return nil, err
	}

	vpr := viper.GetViper()

	vpr.SetConfigFile(configPath)
	vpr.SetConfigType("json")

	setEnvironmentDefaults(vpr)
	setOCRConfigDefaults(vpr)

	conf, err := readConfig[Config](vpr, path)
	if err != nil {
		return nil, err
	}

	if len(conf.Nodes) == 0 {
		conf.Nodes = []string{}
	}

	return conf, nil
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
	Nodes                   []string                `mapstructure:"nodes"`
	Verifier                Verifier                `mapstructure:"verifier"`
	OCR                     AutomationNetworkConfig `mapstructure:"ocr"`
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

type Verifier struct {
	ContractsDir   string `mapstructure:"contracts_directory"`
	ExplorerAPIKey string `mapstructure:"explorer_api_key"`
	NetworkName    string `mapstructure:"network_name"`
}

func setEnvironmentDefaults(vpr *viper.Viper) {
	vpr.SetDefault("rpc_wss_url", "")
	vpr.SetDefault("rpc_http_url", "")
	vpr.SetDefault("chain_id", int64(1337))
	vpr.SetDefault("private_key", "")
	vpr.SetDefault("link_contract_address", "0x")
	vpr.SetDefault("link_eth_feed", "0x")
	vpr.SetDefault("fast_gas_feed", "0x")
	vpr.SetDefault("bootstrap_address", "")
	vpr.SetDefault("groupname", "")

	vpr.SetDefault("service_contract.registry_version", "v2.1")
	vpr.SetDefault("service_contract.registrar_address", "0x")
	vpr.SetDefault("service_contract.registry_address", "0x")
	vpr.SetDefault("service_contract.registry_mode", "DEFAULT")

	vpr.SetDefault("log_trigger_load_contract.contract_address", "0x")
	vpr.SetDefault("log_trigger_load_contract.use_mercury", false)
	vpr.SetDefault("log_trigger_load_contract.use_arbitrum", false)
	vpr.SetDefault("log_trigger_load_contract.auto_log", false)

	vpr.SetDefault("conditional_load_contract.contract_address", "0x")
	vpr.SetDefault("conditional_load_contract.use_arbitrum", false)

	vpr.SetDefault("nodes", []string{})

	vpr.SetDefault("verifier.contracts_directory", "")
	vpr.SetDefault("verifier.explorer_api_key", "")
	vpr.SetDefault("verifier.network_name", "")
}

func setOCRConfigDefaults(vpr *viper.Viper) {
	vpr.SetDefault("ocr.network.delta_progress", 5*time.Second)
	vpr.SetDefault("ocr.network.delta_resend", 10*time.Second)
	vpr.SetDefault("ocr.network.delta_initial", 400*time.Millisecond)
	vpr.SetDefault("ocr.network.delta_round", 2500*time.Millisecond)
	vpr.SetDefault("ocr.network.delta_grace", 40*time.Millisecond)
	vpr.SetDefault("ocr.network.delta_certified_commit_request", 300*time.Millisecond)
	vpr.SetDefault("ocr.network.delta_stage", 30*time.Millisecond)
	vpr.SetDefault("ocr.network.max_rounds", uint64(50))
	vpr.SetDefault("ocr.network.max_duration_query", 20*time.Millisecond)
	vpr.SetDefault("ocr.network.max_duration_observation", 1600*time.Millisecond)
	vpr.SetDefault("ocr.network.max_duration_should_accept_finalized_report", 20*time.Millisecond)
	vpr.SetDefault("ocr.network.max_duration_should_transmit_accepted_report", 20*time.Millisecond)
	vpr.SetDefault("ocr.network.max_faulty_nodes", 1)

	vpr.SetDefault("ocr.onchain.payment_premium_ppb", uint32(700_000_000))
	vpr.SetDefault("ocr.onchain.flat_fee_micro_link", uint32(10_000))
	vpr.SetDefault("ocr.onchain.check_gas_limit", uint32(6_500_000))
	vpr.SetDefault("ocr.onchain.staleness_seconds", int64(90_000))
	vpr.SetDefault("ocr.onchain.gas_ceiling_multiplier", uint16(3))
	vpr.SetDefault("ocr.onchain.min_upkeep_spend", int64(0))
	vpr.SetDefault("ocr.onchain.max_perform_gas", uint32(5_000_000))
	vpr.SetDefault("ocr.onchain.max_check_data_size", uint32(5_000))
	vpr.SetDefault("ocr.onchain.max_perform_data_size", uint32(5_000))
	vpr.SetDefault("ocr.onchain.max_revert_data_size", uint32(5_000))
	vpr.SetDefault("ocr.onchain.fallback_gas_price", int64(1_000))
	vpr.SetDefault("ocr.onchain.fallback_link_price", uint64(5_000_000_000_000_000_000))
	vpr.SetDefault("ocr.onchain.transcoder_address", "0x")
	vpr.SetDefault("ocr.onchain.registrar_address", "0x")
	vpr.SetDefault("ocr.onchain.upkeep_privilege_manager_address", "0x")

	vpr.SetDefault("ocr.offchain.perform_lockout_window", int64(75_000))
	vpr.SetDefault("ocr.offchain.min_confirmations", 0)
	vpr.SetDefault("ocr.offchain.mercury_lookup", false)
	vpr.SetDefault("ocr.offchain.gas_limit_per_report", 5_300_000)
	vpr.SetDefault("ocr.offchain.gas_overhead_per_upkeep", 300_000)
	vpr.SetDefault("ocr.offchain.max_upkeep_batch_size", 1)
	vpr.SetDefault("ocr.offchain.report_block_lag", 0)
	vpr.SetDefault("ocr.offchain.sampling_job_duration", 3_000)
	vpr.SetDefault("ocr.offchain.target_in_rounds", 1)
	vpr.SetDefault("ocr.offchain.target_probability", "0.999")
}
