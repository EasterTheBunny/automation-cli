package config

import (
	"fmt"
)

type PrivateKeys struct {
	Keys []Key `json:"keys"`
}

type Key struct {
	Alias   string `json:"alias"`
	Value   string `json:"value"`
	Address string `json:"address"`
}

func (k PrivateKeys) KeyForAlias(alias string) (Key, error) {
	for _, key := range k.Keys {
		if key.Alias == alias {
			return key, nil
		}
	}

	return Key{}, fmt.Errorf("private key not found")
}
