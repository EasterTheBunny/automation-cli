package participant

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	"github.com/easterthebunny/automation-cli/internal/node"
)

func init() {
}

var (
	resetCmd = &cobra.Command{
		Use:   "reset [NODE] [IMAGE]",
		Short: "Create and add participant nodes with provided Docker image",
		Long:  `Create and add participant nodes with provided Docker image.`,
		Example: `To add nodes, a bootstrap node will need to first exist. Then run the following to create 5 nodes in
the same environment as above:

$ automation-cli network participant reset 0 chainlink:latest --log-level="debug" --environment="non.default"

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
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, env, key, err := prepare(cmd)
			if err != nil {
				return err
			}

			var conf *config.NodeConfig

			// need to identify the node name
			for idx, nodeConf := range env.Participants {
				if nodeConf.Name == args[0] || strconv.FormatInt(int64(idx), 10) == args[0] {
					conf = &env.Participants[idx]
				}
			}

			if conf == nil {
				return fmt.Errorf("no participant found by the provided name")
			}

			basePath, err := path.Path()
			if err != nil {
				return err
			}

			alias := "default"

			var privateKey *string
			if key != nil {
				privateKey = &key.Value
				alias = key.Alias
			}

			nodeConfigPath := fmt.Sprintf("%s/%s", basePath, conf.Name)

			// set the new image
			conf.Image = args[1]
			conf.PrivateKeyAlias = alias

			if err := node.CreateParticipantNode(
				cmd.Context(),
				env.Groupname, env.Registry.Address,
				*env.Bootstrap,
				conf,
				nodeConfigPath,
				privateKey,
				true,
			); err != nil {
				return err
			}

			return config.Write(path.MustWrite(config.EnvironmentConfigFilename), env)
		},
	}
)
