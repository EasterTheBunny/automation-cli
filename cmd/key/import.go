package key

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/internal/config"
	cliio "github.com/easterthebunny/automation-cli/internal/io"
)

type GanacheAddresses struct {
	Addresses   map[string]string `json:"addresses"`
	PrivateKeys map[string]string `json:"private_keys"`
}

var (
	importGanacheCmd = &cobra.Command{
		Use:   "import-ganache [FILE]",
		Short: "Import accounts and private keys from a Ganache instance",
		Long:  `Import accounts and private keys from a Ganache instance.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}

			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			var addresses GanacheAddresses

			if err := json.Unmarshal(data, &addresses); err != nil {
				return err
			}

			keys := make([]config.Key, 0, len(addresses.Addresses))
			for _, address := range addresses.Addresses {
				privkey, ok := addresses.PrivateKeys[address]
				if !ok {
					continue
				}

				nameIdx := strconv.FormatInt(int64(len(keys)), 10)
				if nameIdx == "0" {
					nameIdx = "primary"
				}

				keys = append(keys, config.Key{
					Alias:   fmt.Sprintf("ganache-%s", nameIdx),
					Value:   privkey,
					Address: address,
				})
			}

			env := cliio.EnvironmentFromContext(cmd.Context())
			if env == nil {
				return fmt.Errorf("environment not found")
			}

			conf, err := config.ReadPrivateKeysFrom(env.Root.MustRead(config.PrivateKeyConfigFilename))
			if err != nil {
				return err
			}

			toAdd := []config.Key{}

		MainLoop:
			for _, newKey := range keys {
				for idx, key := range conf.Keys {
					if newKey.Alias == key.Alias {
						conf.Keys[idx] = newKey

						continue MainLoop
					}
				}

				toAdd = append(toAdd, newKey)
			}

			conf.Keys = append(conf.Keys, toAdd...)

			return config.WritePrivateKeys(env.Root.MustWrite(config.PrivateKeyConfigFilename), conf)
		},
	}

	importGethCmd = &cobra.Command{
		Use:   "import-geth [NAME] [FILE | DIRECTORY]",
		Short: "Import accounts and private keys from geth",
		Long:  `Import accounts and private keys from geth.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePaths, err := getFilePaths(args[1])
			if err != nil {
				return err
			}

			env := cliio.EnvironmentFromContext(cmd.Context())
			if env == nil {
				return fmt.Errorf("environment not found")
			}

			conf, err := config.ReadPrivateKeysFrom(env.Root.MustRead(config.PrivateKeyConfigFilename))
			if err != nil {
				return err
			}

			for idx, path := range filePaths {
				pkData, err := readAllFrom(path)
				if err != nil {
					return err
				}

				var password string

				if passwordPath != "" {
					pwData, err := readAllFrom(passwordPath)
					if err != nil {
						return err
					}

					password = strings.TrimSpace(string(pwData))
				}

				// This can take up a good bit of RAM and time. When running on the remote-test-runner, this can lead to OOM
				// issues. So we avoid running in parallel; slower, but safer.
				decryptedKey, err := keystore.DecryptKey(pkData, password)
				if err != nil {
					return err
				}

				privateKeyBytes := crypto.FromECDSA(decryptedKey.PrivateKey)

				conf.Keys = append(conf.Keys, config.Key{
					Alias:   fmt.Sprintf("%s-%d", args[0], idx),
					Value:   hexutil.Encode(privateKeyBytes)[2:],
					Address: decryptedKey.Address.Hex(),
				})
			}

			return config.WritePrivateKeys(env.Root.MustWrite(config.PrivateKeyConfigFilename), conf)
		},
	}
)

func getFilePaths(path string) ([]string, error) {
	detail, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if detail.IsDir() {
		// walk the directory
		entries, err := os.ReadDir(abs)
		if err != nil {
			return nil, err
		}

		paths := []string{}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			paths = append(paths, fmt.Sprintf("%s/%s", abs, entry.Name()))
		}

		return paths, nil
	} else {
		return []string{abs}, nil
	}
}

func readAllFrom(path string) ([]byte, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}
