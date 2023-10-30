package key

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
	"github.com/easterthebunny/automation-cli/cmd/automation-cli/context"
)

type GanacheAddresses struct {
	Addresses   map[string]string `json:"addresses"`
	PrivateKeys map[string]string `json:"private_keys"`
}

var (
	importGanacheCmd = &cobra.Command{
		Use:   "import [FILE]",
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

			paths := context.GetPathsFromContext(cmd.Context())
			if paths == nil {
				return fmt.Errorf("missing config path in context")
			}

			conf, err := config.GetPrivateKeyConfig(paths.Base)
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

			return config.SavePrivateKeyConfig(paths.Base, conf)
		},
	}
)
