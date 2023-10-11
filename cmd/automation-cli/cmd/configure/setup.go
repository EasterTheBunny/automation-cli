package configure

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
