package command

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/internal/asset"
	"github.com/easterthebunny/automation-cli/internal/node"
)

var networkManagementCmd = &cobra.Command{
	Use:   "network [ACTION]",
	Short: "Manage network components such as a bootstrap node and/or automation nodes",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
}

var networkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all nodes on the current network configuration",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		for _, nodeName := range conf.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", nodeName)
		}

		return nil
	},
}

var networkAddCmd = &cobra.Command{
	Use:       "add [TYPE] [IMAGE]",
	Short:     "Create and add network components such as a bootstrap node and/or automation nodes",
	Long:      ``,
	ValidArgs: []string{"bootstrap", "participant"},
	Args:      cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		paths := GetPathsFromContext(cmd.Context())
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
				keyConf := GetKeyConfigFromContext(cmd.Context())
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

var networkFundCmd = &cobra.Command{
	Use:   "fund [NODE] [AMOUNT]",
	Short: "Transfer funds to node address.",
	Long:  `Transfer funds from the default account to configured node address.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := GetConfigFromContext(cmd.Context())
		if conf == nil {
			return fmt.Errorf("missing config path in context")
		}

		paths := GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		dConfig := config.GetDeployerConfig(conf)

		keyConf := GetKeyConfigFromContext(cmd.Context())
		if keyConf == nil {
			return fmt.Errorf("missing private key config")
		}

		for _, key := range keyConf.Keys {
			if key.Alias == dConfig.PrivateKey {
				dConfig.PrivateKey = key.Value

				break
			}
		}

		deployer, err := asset.NewDeployer(&dConfig)
		if err != nil {
			return err
		}

		nodeName := ""

		for _, n := range conf.Nodes {
			if n == args[0] {
				nodeName = n
			}
		}

		if nodeName == "" {
			return fmt.Errorf("node not available")
		}

		nConf, _, err := config.GetNodeConfig(fmt.Sprintf("%s/%s", paths.Environment, nodeName))
		if err != nil {
			return err
		}

		if nConf.Address == "" {
			return fmt.Errorf("node address not available")
		}

		addr := nConf.Address

		var amount uint64

		allDigits := regexp.MustCompile(`^[0-9]+$`)
		usesExp := regexp.MustCompile(`^[0-9]+\^[0-9]+$`)
		strAmount := strings.TrimSpace(args[1])

		if allDigits.MatchString(strAmount) {
			amount, err = strconv.ParseUint(strAmount, 10, 64)
			if err != nil {
				return err
			}
		} else if usesExp.MatchString(strAmount) {
			parts := strings.Split(strAmount, "^")

			zeros, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return err
			}

			var strVal strings.Builder

			strVal.WriteString(parts[0])

			for x := 0; x < int(zeros); x++ {
				strVal.WriteString("0")
			}

			amount, err = strconv.ParseUint(strVal.String(), 10, 64)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("not a valid amount")
		}

		return deployer.Send(cmd.Context(), addr, amount)
	},
}
