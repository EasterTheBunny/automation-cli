package main

import (
	"fmt"
	"os"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/command"
)

func main() {
	command.InitializeCommands()

	if err := command.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
