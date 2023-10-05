package command

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config [ACTION]",
	Short: "Shortcut to quickly update config var",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.MinimumNArgs(1),
}

var configSetVarCmd = &cobra.Command{
	Use:   "set [NAME] [VALUE]",
	Short: "Shortcut to quickly update config variables",
	Long:  `Update config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set(args[0], args[1])

		return nil
	},
}

var configGetVarCmd = &cobra.Command{
	Use:   "get [NAME]",
	Short: "Read config variables",
	Long:  `Read config variable by name. Only accepts lower case and '.' between nested values.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val := viper.Get(args[0])

		switch val.(type) {
		case string:
			fmt.Fprintf(cmd.OutOrStdout(), "%s", val)
		case uint, uint8, uint16, uint32, uint64, int, int32, int64:
			fmt.Fprintf(cmd.OutOrStdout(), "%d", val)
		case bool:
			fmt.Fprintf(cmd.OutOrStdout(), "%t", val)
		}

		return nil
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup environment",
	Long:  `Setup initial environment configurations`,
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), "Supply configuration options below or press 'enter' to accept defaults")
		fmt.Fprintln(cmd.OutOrStdout(), "")

		reader := bufio.NewReader(cmd.InOrStdin())

		fmt.Fprint(cmd.OutOrStdout(), "Enter the chain ID [1337]: ") // default to test chain

		chainIDStr, _ := reader.ReadString('\n')

		if len(strings.Trim(chainIDStr, "\n")) == 0 {
			viper.Set("chain_id", 1337)
		} else {
			chainID, err := strconv.ParseInt(strings.Trim(chainIDStr, "\n"), 10, 64)
			if err != nil {
				return err
			}

			viper.Set("chain_id", chainID)
		}

		fmt.Fprint(cmd.OutOrStdout(), "Enter Private Key [default]: ")

		privKeyStr, _ := reader.ReadString('\n')

		if len(strings.Trim(privKeyStr, "\n")) == 0 {
			viper.Set("private_key", "default")
		} else {
			viper.Set("private_key", strings.Trim(privKeyStr, "\n"))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Enter the RPC HTTP URL: ")

		httpRPC, _ := reader.ReadString('\n')

		viper.Set("rpc_http_url", strings.Trim(httpRPC, "\n"))

		fmt.Fprintf(cmd.OutOrStdout(), "Enter the RPC WSS URL: ")

		httpWSS, _ := reader.ReadString('\n')

		viper.Set("rpc_wss_url", strings.Trim(httpWSS, "\n"))

		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			return err
		}

		viper.Set("groupname", env)

		return nil
	},
}

var configStorePKCmd = &cobra.Command{
	Use:   "pk-store [NAME]",
	Short: "Store a private key with the reference name",
	Long:  `Securely store private keys under alias names for reference in configurations.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprint(cmd.OutOrStdout(), "Enter private key: ")

		pkBytes, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			return err
		}

		paths := GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		conf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		for idx, key := range conf.Keys {
			if key.Alias == args[0] {
				conf.Keys[idx].Value = string(pkBytes)

				return config.SavePrivateKeyConfig(paths.Base, conf)
			}
		}

		conf.Keys = append(conf.Keys, config.Key{
			Alias: args[0],
			Value: string(pkBytes),
		})

		return config.SavePrivateKeyConfig(paths.Base, conf)
	},
}

var configCreatePKCmd = &cobra.Command{
	Use:   "pk-create [NAME]",
	Short: "Create a private key with the reference name",
	Long:  `Create a new private key and store it under the provided reference name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return err
		}

		keyBytes := crypto.FromECDSA(privateKey)

		paths := GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		conf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		conf.Keys = append(conf.Keys, config.Key{
			Alias: args[0],
			Value: hexutil.Encode(keyBytes),
		})

		return config.SavePrivateKeyConfig(paths.Base, conf)
	},
}

var configListPKCmd = &cobra.Command{
	Use:   "pk-list",
	Short: "Print list of private key aliases",
	Long:  `List existing private key aliases`,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := GetPathsFromContext(cmd.Context())
		if paths == nil {
			return fmt.Errorf("missing config path in context")
		}

		conf, err := config.GetPrivateKeyConfig(paths.Base)
		if err != nil {
			return err
		}

		for _, key := range conf.Keys {
			fmt.Println(key.Alias)
		}

		return nil
	},
}
