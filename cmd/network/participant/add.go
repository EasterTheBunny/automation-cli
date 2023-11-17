package participant

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/io"
	"github.com/easterthebunny/automation-cli/internal/node"
)

func init() {
	addCmd.Flags().Uint8Var(&count, "count", 1, "total number of nodes to create with this configuration")
	addCmd.Flags().StringVar(&mercuryLegacyURL, "mercury-legacy-url", "https://chain2.old.link", "legacy url to the mercury server")
	addCmd.Flags().StringVar(&mercuryURL, "mercury-url", "https://chain2.link", "url to the mercury server")
	addCmd.Flags().StringVar(&mercuryID, "mercury-id", "username2", "mercury user id")
	addCmd.Flags().StringVar(&mercuryKey, "mercury-key", "password2", "mercury user key")
}

var (
	count            uint8
	mercuryLegacyURL string
	mercuryURL       string
	mercuryID        string
	mercuryKey       string

	addCmd = &cobra.Command{
		Use:   "add [IMAGE]",
		Short: "Create and add participant nodes with provided Docker image",
		Long:  `Create and add participant nodes with provided Docker image.`,
		Example: `To add nodes, a bootstrap node will need to first exist. Then run the following to create 5 nodes in
the same environment as above:

$ automation-cli network participant add chainlink:latest --count=5 --environment="non.default"

A log level can be specified to reduce or increase the log output of individual nodes in the case that only one node is
being evaluated and the others only exist to create the network. Creating this type of network can be done with the
following where all nodes are added to the default network.

$ automation-cli network participant add chainlink:latest --count=4 --log-level="error"
$ automation-cli network participant add chainlink:test --count=1 --log-level="info"

Participant nodes can be assigned different private keys in the case some already exist and are funded. To set up a
network with specific private keys, do the following. This will save the configuration to the default environment.

$ automation-cli network participant add chainlink:test --count=1 --key="mumbai-test-one"
$ automation-cli network participant add chainlink:test --count=1 --key="mumbai-test-two"
$ automation-cli network participant add chainlink:test --count=1 --key="mumbai-test-three"
$ automation-cli network participant add chainlink:test --count=1 --key="mumbai-test-four"

This assumes the four provided keys are aliases to existing saved keys.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			var (
				privateKey *string
			)

			alias := "default"

			if key != nil {
				privateKey = &key.Value
				alias = key.Alias
			}

			basePath, err := path.Path()
			if err != nil {
				return err
			}

			existing := len(env.Participants)

			for idx := 0; idx < int(count); idx++ {
				nodeID := idx + existing

				nodeConf := config.NodeConfig{
					HostType:      config.Docker,
					Name:          fmt.Sprintf("participant-%d", nodeID),
					Image:         args[0],
					LogLevel:      logLevel,
					ListenPort:    uint16(6688 + nodeID),
					LoginName:     config.DefaultChainlinkNodeLogin,
					LoginPassword: config.DefaultChainlinkNodePassword,

					PrivateKeyAlias: alias,
					ChainID:         env.ChainID,
					WSURL:           env.WSURL,
					HTTPURL:         env.HTTPURL,

					MercuryLegacyURL: config.DefaultMercuryLegacyURL,
					MercuryURL:       config.DefaultMercuryURL,
					MercuryID:        config.DefaultMercuryID,
					MercuryKey:       config.DefaultMercuryKey,
				}

				nodeConfigPath := fmt.Sprintf("%s/%s", basePath, nodeConf.Name)

				if err := node.CreateParticipantNode(
					cmd.Context(),
					env.Groupname,
					env.Registry.Address,
					*env.Bootstrap,
					&nodeConf,
					nodeConfigPath,
					privateKey,
					false, // don't attempt to reset a node
				); err != nil {
					return err
				}

				env.Participants = append(env.Participants, nodeConf)
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)

func prepare(cmd *cobra.Command) (io.Environment, config.Environment, *config.Key, error) {
	var (
		env config.Environment
		key *config.Key
		err error
	)

	path := io.EnvironmentFromContext(cmd.Context())
	if path == nil {
		return io.Environment{}, env, nil, fmt.Errorf("environment not found")
	}

	env, err = config.ReadFrom(path.MustRead(config.EnvironmentConfigFilename))
	if err != nil {
		return io.Environment{}, env, nil, err
	}

	keys, err := config.ReadPrivateKeysFrom(path.Root.MustRead(config.PrivateKeyConfigFilename))
	if err != nil {
		return io.Environment{}, env, nil, err
	}

	pkOverride, err := cmd.Flags().GetString("key")
	if err != nil {
		return io.Environment{}, env, nil, err
	}

	if pkOverride != "" {
		keyVal, err := keys.KeyForAlias(pkOverride)
		if err != nil {
			return io.Environment{}, env, nil, err
		}

		key = &keyVal
	}

	return *path, env, key, nil
}
