package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
	"github.com/easterthebunny/automation-cli/internal/node"
)

var (
	setCmd = &cobra.Command{
		Use:   "set [IMAGE]",
		Short: "Set a bootstrap node for a network",
		Long:  `Set a bootstrap node for a network`,
		Example: `The following will create a bootstrap node and save the interaction details to the provided
environment named 'non.default'.

$ automation-cli network bootstrap set chainlink:latest --environment="non.default"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := io.EnvironmentFromContext(cmd.Context())
			if path == nil {
				return fmt.Errorf("environment not found")
			}

			env, err := config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
			if err != nil {
				return err
			}

			if env.Registry == nil {
				return fmt.Errorf("registry required to create bootstrap node")
			}

			basePath, err := path.Path()
			if err != nil {
				return err
			}

			env.Bootstrap = &config.NodeConfig{
				HostType:            config.Docker,
				Name:                "bootstrap",
				Image:               args[0],
				LogLevel:            logLevel,
				ListenPort:          5688,
				LoginName:           config.DefaultChainlinkNodeLogin,
				LoginPassword:       config.DefaultChainlinkNodePassword,
				IsBootstrap:         true,
				BootstrapListenPort: 8000,

				ChainID: env.ChainID,
				WSURL:   env.WSURL,
				HTTPURL: env.HTTPURL,

				MercuryLegacyURL: config.DefaultMercuryLegacyURL,
				MercuryURL:       config.DefaultMercuryURL,
				MercuryID:        config.DefaultMercuryID,
				MercuryKey:       config.DefaultMercuryKey,
			}

			nodeConfigPath := fmt.Sprintf("%s/%s", basePath, "bootstrap")

			if err := node.CreateBootstrapNode(
				cmd.Context(),
				env.Groupname,
				env.Registry.Address,
				env.Bootstrap, nodeConfigPath, true); err != nil {
				return err
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
