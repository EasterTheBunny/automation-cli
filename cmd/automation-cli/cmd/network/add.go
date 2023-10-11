package network

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
	"github.com/easterthebunny/automation-cli/internal/node"
)

func init() {
	_ = addCmd.Flags().Int8("count", 1, "total number of nodes to create with this configuration")
	_ = addCmd.Flags().String(
		"private-key",
		"default",
		"use a specific private key. use only an alias for a previously saved private key.",
	)
	_ = addCmd.Flags().String(
		"log-level",
		"error",
		"set the log level for the node",
	)
}

var addCmd = &cobra.Command{
	Use:       "add [TYPE] [IMAGE]",
	Short:     "Create and add network components such as a bootstrap node and/or automation nodes",
	Long:      ``,
	ValidArgs: []string{"bootstrap", "participant"},
	Args:      cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := context.GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		paths := context.GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		logLevel, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return err
		}

		switch args[0] {
		case "bootstrap":
			nodeConfigPath := fmt.Sprintf("%s/%s", paths.Environment, "bootstrap")
			str, err := node.CreateBootstrapNode(cmd.Context(), node.NodeConfig{
				ChainID:          conf.ChainID,
				NodeWSSURL:       conf.RPCWSSURL,
				NodeHttpURL:      conf.RPCHTTPURL,
				LogLevel:         logLevel,
				MercuryLegacyURL: "https://chain2.old.link",
				MercuryURL:       "https://chain2.link",
				MercuryID:        "username2",
				MercuryKey:       "password2",
			}, conf.Groupname, args[1], conf.ServiceContract.RegistryAddress, 5688, 8000, nodeConfigPath, false)
			if err != nil {
				return err
			}

			viper.Set("bootstrap_address", str)
		case "participant":
			count, err := cmd.Flags().GetInt8("count")
			if err != nil {
				return err
			}

			withPK, err := cmd.Flags().GetString("private-key")
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

			existing := len(conf.Nodes)

			for idx := 0; idx < int(count); idx++ {
				nodeID := idx + existing
				nodeName := fmt.Sprintf("participant-%d", nodeID)
				nodeConfigPath := fmt.Sprintf("%s/%s", paths.Environment, nodeName)

				_, vpr, err := config.GetNodeConfig(nodeConfigPath)
				if err != nil {
					return err
				}

				clNode, err := node.CreatParticipantNode(
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

				conf.Nodes = append(conf.Nodes, nodeName)
			}

			viper.Set("nodes", conf.Nodes)
		default:
			return fmt.Errorf("unrecognized argument: %s", args[0])
		}

		return nil
	},
}
