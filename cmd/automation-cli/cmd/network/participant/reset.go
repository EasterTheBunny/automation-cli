package participant

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
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

$ automation-cli network participant reset 0 chainlink:latest --count=5 --environment="non.default"

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
			conf := context.GetConfigFromContext(cmd.Context())
			if conf == nil {
				return fmt.Errorf("missing config path in context")
			}

			paths := context.GetPathsFromContext(cmd.Context())
			if paths == nil {
				return fmt.Errorf("missing config path in context")
			}

			withPK, err := cmd.Flags().GetString("key")
			if err != nil {
				return err
			}

			var privateKey *string
			if withPK != "default" {
				keyConf := context.GetKeyConfigFromContext(cmd.Context())
				if keyConf == nil {
					return fmt.Errorf("missing private key config")
				}

				for _, key := range keyConf.Keys {
					if key.Alias == withPK {
						privateKey = &key.Value

						break
					}
				}
			}

			parsedID, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			nodeID := parsedID
			nodeName := fmt.Sprintf("participant-%d", nodeID)

			nodeConfigPath := fmt.Sprintf("%s/%s", paths.Environment, nodeName)

			_, vpr, err := config.GetNodeConfig(nodeConfigPath)
			if err != nil {
				return err
			}

			clNode, err := node.CreateParticipantNode(
				cmd.Context(),
				node.NodeConfig{
					ChainID:          conf.ChainID,
					NodeWSSURL:       conf.RPCWSSURL,
					NodeHttpURL:      conf.RPCHTTPURL,
					LogLevel:         logLevel,
					MercuryLegacyURL: "https://chain2.old.link",
					MercuryURL:       "https://chain2.link",
					MercuryID:        "username2",
					MercuryKey:       "password2",
				},
				uint16(6688+nodeID),
				conf.Groupname,
				nodeName,
				args[1],
				conf.ServiceContract.RegistryAddress,
				conf.BootstrapAddress,
				nodeConfigPath,
				privateKey,
				true,
			)
			if err != nil {
				return err
			}

			vpr.Set("chainlink_image", args[1])
			vpr.Set("management_url", clNode.URL())
			vpr.Set("address", clNode.Address)

			if err := config.SaveViperConfig(vpr, nodeConfigPath); err != nil {
				return err
			}

			return nil
		},
	}
)
