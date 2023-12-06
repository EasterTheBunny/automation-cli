package configure

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	cliio "github.com/easterthebunny/automation-cli/internal/io"
)

const (
	defaultChainID         int64 = 1337
	defaultPrivateKeyAlias       = "default"
)

func init() {
	configSetupCmd.Flags().BoolVar(&jsonInput, "json", false, "read configuration from stdin as JSON")
}

type envConfig struct {
	ChainID    int64  `json:"chain_id"`
	PrivateKey string `json:"private_key_alias"`
	HTTPRPC    string `json:"http_rpc"`
	WSRPC      string `json:"ws_rpc"`
}

var (
	jsonInput bool

	configSetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Setup environment",
		Long:  `Setup initial environment configurations`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var conf envConfig

			switch {
			case jsonInput:
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}

				if err := json.Unmarshal(data, &conf); err != nil {
					return err
				}
			default:
				fmt.Fprintln(cmd.OutOrStdout(), "Supply configuration options below or press 'enter' to accept defaults")
				fmt.Fprintln(cmd.OutOrStdout(), "")

				reader := bufio.NewReader(cmd.InOrStdin())

				if err := promptChainID(reader, cmd.OutOrStdout(), &conf); err != nil {
					return err
				}

				if err := promptPrivateKey(reader, cmd.OutOrStdout(), &conf); err != nil {
					return err
				}

				if err := promptHTTPRPC(reader, cmd.OutOrStdout(), &conf); err != nil {
					return err
				}

				if err := promptWSRPC(reader, cmd.OutOrStdout(), &conf); err != nil {
					return err
				}
			}

			env := cliio.EnvironmentFromContext(cmd.Context())
			if env == nil {
				return fmt.Errorf("environment not available")
			}

			envConf := config.Environment{
				Groupname:       env.Name,
				ChainID:         conf.ChainID,
				PrivateKeyAlias: defaultPrivateKeyAlias,
				HTTPURL:         conf.HTTPRPC,
				WSURL:           conf.WSRPC,
				GasLimit:        config.DefaultDeployerGasLimit,
			}

			if conf.PrivateKey != "" {
				envConf.PrivateKeyAlias = conf.PrivateKey
			}

			return config.Write(env.MustWrite(config.EnvironmentConfigFilename), envConf)
		},
	}
)

func promptChainID(reader *bufio.Reader, writer io.Writer, conf *envConfig) error {
	fmt.Fprintf(writer, "Enter the chain ID [%d]: ", defaultChainID) // default to test chain

	chainIDStr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if len(strings.Trim(chainIDStr, "\n")) > 0 {
		chainID, err := strconv.ParseInt(strings.Trim(chainIDStr, "\n"), 10, 64)
		if err != nil {
			return err
		}

		conf.ChainID = chainID
	}

	return nil
}

func promptPrivateKey(reader *bufio.Reader, writer io.Writer, conf *envConfig) error {
	fmt.Fprintf(writer, "Enter Private Key [%s]: ", defaultPrivateKeyAlias)

	privKeyStr, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if len(strings.Trim(privKeyStr, "\n")) > 0 {
		conf.PrivateKey = strings.Trim(privKeyStr, "\n")
	}

	return nil
}

func promptHTTPRPC(reader *bufio.Reader, writer io.Writer, conf *envConfig) error {
	fmt.Fprintf(writer, "Enter the RPC HTTP URL: ")

	httpRPC, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	conf.HTTPRPC = strings.Trim(httpRPC, "\n")

	return nil
}

func promptWSRPC(reader *bufio.Reader, writer io.Writer, conf *envConfig) error {
	fmt.Fprintf(writer, "Enter the RPC WSS URL: ")

	httpWSS, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	conf.WSRPC = strings.Trim(httpWSS, "\n")

	return nil
}
