The Automation-CLI is intended to be a tool that helps set up an automation environment piece-by-piece and interact
with that environment. Each stage of a setup stores state values locally so that specific assets can be referenced
in later commands.

# Usage
Use the Automation-CLI like any other command line application. Each command is designed to do small steps of a larger
environment interaction process.

## Build and Help
Use make to build the application. The command line tool is structured using Cobra, so you can ask for command help at
any level with the `-h` flag.

**Make and Run**
```
$ make build && ./bin/automation-cli -h
```

**Install and Run**
```
$ cd cmd/automation-cli
$ go install
$ automation-cli -h
```

## Setup
Configurations are stored by environment such that the same tool can manage multiple environments at once. For example
you are working with a local and a staging network on Polygon Mumbai. One environment name could be `local.mumbai` and
the other could be `staging.mumbai`. This maintains state for both using the same tool. The `--environment` flag is
global and allows setting the environment to be set for any command.

```
$ automation-cli --environment="local.mumbai"
```

### Add Private Keys
Private keys are separate from environment configurations to allow configurations to be shared. Configurations contain
alias references to private keys. To add a private key run the following, but be aware that private keys are not stored
per network, they are globally available:

```
$ automation-cli key store [ALIAS]
```

If you don't have a private key available and wish to create and store a new one, run the following:

```
$ automation-cli key create [ALIAS]
```

Any of the above keys can be funded after they are included as private keys on a participant node. You can fund a
participant node (private key address) by the following:

```
$ automation-cli network fund local.mumbai-participant-2 10^17
```

This will fund the node's address with 10^17 native token from the default private key or add `--key="some.other.key"`
as an alternative source.

### Setup Environment
Environments need a few values to get started such as chain id, private keys, and RPC urls. Run the following to set up
a new environment or update the values in an existing environment and remember that private key input here should be
the alias and not the actual private key:

```
$ automation-cli config setup --environment="some.environment"
```

## Contract Management
Generally you can connect to existing contracts or deploy new ones.

**Available Contracts**
- registrar
- registry
- verifiable-load-log-trigger
- verifiable-load-conditional

### Connect to Existing
If you know the address of an existing contract and just want to connect and store that contract for future use in the
CLI tool, use the following command. Remember to define your environment or it will select the default (default).

```
$ automation-cli contract connect registry [ADDRESS] --environment"some.environment"
```

You can get help on the available contract types that you can connect to:

```
$ automation-cli contract connect -h
```

### Deploying Contracts
Deploying contracts uses a similar method, but will save the resulting contract address to the environment config state.

```
$ automation-cli contract deploy registry --environment="some.environment"
```

This command will use the existing configuration within the environment to deploy a contract. Some contracts require
specific configurations and at the moment you will need to set these configurations manually in the state directory.

### Interactions
Some contracts have interactions you can do through the CLI tool. These interactions are not intended to replace 
interacting with a contract using a wallet and browser, but instead roll up more complex interactions into simple
commands. An example interaction follows where statistics are printed from a verifiable load contract:

```
# connect to the verifiable load contract first if you haven't already
$ automation-cli contract connect verifiable-load-conditional [ADDRESS]

# run an interaction against the contract
$ automation-cli contract interact verifiable-load-conditional get-stats
```