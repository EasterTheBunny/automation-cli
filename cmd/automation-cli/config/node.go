package config

import "github.com/spf13/viper"

func GetNodeConfig(path string) (*NodeConfig, *viper.Viper, error) {
	configPath, err := ensureExists(path, "config.json")
	if err != nil {
		return nil, nil, err
	}

	vpr := viper.New()

	vpr.SetConfigFile(configPath)
	vpr.SetConfigType("json")

	setNodeConfigDefaults(vpr)

	conf, err := readConfig[NodeConfig](vpr, path)
	if err != nil {
		return nil, nil, err
	}

	return conf, vpr, nil
}

type NodeConfig struct {
	ChainlinkImage string `mapstructure:"chainlink_image"`
	ManagementURL  string `mapstructure:"management_url"`
	Address        string `mapstructure:"address"`
}

func setNodeConfigDefaults(vpr *viper.Viper) {
	vpr.SetDefault("chainlink_image", "chainlink:latest")
	vpr.SetDefault("management_url", "")
	vpr.SetDefault("address", "")
}
