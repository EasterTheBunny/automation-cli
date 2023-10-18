package load

import "github.com/spf13/cobra"

func init() {
	RootCmd.PersistentFlags().StringVar(&upkeepType, "type", "conditional", "upkeep type (conditional, log-trigger)")

	RootCmd.AddCommand(setCmd)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(registerCmd)
	RootCmd.AddCommand(cancelCmd)
	RootCmd.AddCommand(readStatsCmd)
}

var (
	upkeepType string

	RootCmd = &cobra.Command{
		Use:   "verifiable-load [ACTION]",
		Short: "Manage contracts related to verifiable load.",
		Long:  `Create or connect to existing contracts, register upkeeps, view load results.`,
		Example: `With an existing running environment and key (non-default) to do the following:

- cancel all existing upkeeps
- register new upkeeps
- collect and print delay statistics to console

$ automation-cli contract verifiable-load register-upkeeps --type="log-trigger" --environment="non.default" --key="mumbai-dev" --cancel-upkeeps
$ automation-cli contract verifiable-load get-stats --type="log-trigger" --environment="non.default" --key="mumbai-dev"

The above example assumes an environment set up with the name "non.default" and a verifiable load log trigger contract
deployed to that environment. Also, a key with the name "mumbai-dev" is used. The first command cancels all existing
upkeeps on the contract and deploys 5 (default count) new ones. The second command reads the captured delay data on the
contract and displays it to the console.

With the same setup above, to cancel all existing upkeeps only do the following:

$ automation-cli contract verifiable-load cancel-upkeeps --type="log-trigger" --environment="non.default" --key="mumbai-dev"

With the default environment and the same private key, leave off the "environment" and "key" parameters:

$ automation-cli contract verifiable-load register-upkeeps --type="log-trigger" --cancel-upkeeps
$ automation-cli contract verifiable-load get-stats --type="log-trigger"

In many cases you will want to send LINK to the contract in the process of registering upkeeps. By default, the
upkeep register function will NOT send LINK and will rely on the contract already having LINK available. To send LINK
as part of the registration process, use the --send-link argument as follows:

$ automation-cli contract verifiable-load register-upkeeps --type="log-trigger" --send-link`,
		Args: cobra.MinimumNArgs(1),
	}
)
