The Automation-CLI is intended to be a tool that helps set up an automation environment piece-by-piece and interact
with that environment. Each stage of a setup stores state values locally so that specific assets can be referenced
in later commands.

# Usage
Use the Automation-CLI like any other command line application. Each command is designed to do small steps of a larger
environment interaction process.

## Build and Help
Use make to build the application. The command line tool is structured using Cobra, so you can ask for command help at
any level with the `-h` flag.

```
make build && ./bin/automation-cli -h
```

## Setup
Configurations are stored by environment such that the same tool can manage multiple environments at once. For example
you are working with a local and a staging network on Polygon Mumbai. One environment name could be `local.mumbai` and
the other could be `staging.mumbai`. This maintains state for both using the same tool. The `--environment` flag is
global and allows setting the environment to be set for any command.

```
automation-cli --environment="local.mumbai"
```

### Add Private Keys
Private keys are separate from environment configurations to allow configurations to be shared. Configurations contain
alias references to private keys. To add a private key run the following, but be aware that private keys are not stored
per network, they are globally available:

```
automation-cli config pk-store [ALIAS]
```

### Setup Environment
Environments need a few values to get started such as chain id, private keys, and RPC urls. Run the following to set up
a new environment or update the values in an existing environment and remember that private key input here should be
the alias and not the actual private key:

```
automation-cli config setup --environment="some.environment"
```