package asset

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type VerifyContractConfig struct {
	ContractsDir   string
	NodeHTTPURL    string
	ExplorerAPIKey string
	NetworkName    string
}

func PrintVerifyContractCommand(config VerifyContractConfig, params ...string) {
	// Change to the contracts directory where the hardhat.config.ts file is located
	if err := changeToContractsDirectory(config.ContractsDir); err != nil {
		log.Fatalf("failed to change to directory where the hardhat.config.ts file is located: %v", err)
	}

	// Append the address and params to the commandArgs slice
	commandArgs := append([]string{}, params...)

	// Format the command string with the commandArgs
	command := fmt.Sprintf(
		"NODE_HTTP_URL='%s' EXPLORER_API_KEY='%s' NETWORK_NAME='%s' pnpm hardhat verify --network env %s",
		config.NodeHTTPURL,
		config.ExplorerAPIKey,
		config.NetworkName,
		strings.Join(commandArgs, " "),
	)

	fmt.Println("Running command to verify contract: ", command)

	if err := runCommand(command); err != nil {
		log.Println("Contract verification on Explorer failed: ", err)
	}
}

func changeToContractsDirectory(path string) error {
	// Check if hardhat.config.ts exists in the current directory, return if it does
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	// Change directory
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Check if hardhat.config.ts exists in the current directory
	if _, err := os.Stat(filepath.Join(path, "hardhat.config.ts")); err != nil {
		return fmt.Errorf("hardhat.config.ts not found in the current directory")
	}

	log.Printf("Successfully changed to directory %s\n", path)

	return nil
}

func runCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
