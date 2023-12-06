package config

import (
	"encoding/json"
	"io"

	toml "github.com/pelletier/go-toml/v2"
)

// Write encodes the Environment, writes to the writer, and closes the writer.
func Write(writer io.WriteCloser, env Environment) error {
	defer writer.Close()

	enc := toml.NewEncoder(writer)
	enc.SetIndentTables(true)

	return enc.Encode(env)
}

func ReadFrom(reader io.ReadCloser) (Environment, error) {
	defer reader.Close()

	var env Environment

	if err := toml.NewDecoder(reader).Decode(&env); err != nil {
		return env, err
	}

	return env, nil
}

// WritePrivateKeys encodes the PrivateKeys, writes to the writer, and closes the writer.
func WritePrivateKeys(writer io.WriteCloser, keys PrivateKeys) error {
	defer writer.Close()

	return json.NewEncoder(writer).Encode(keys)
}

func ReadPrivateKeysFrom(reader io.ReadCloser) (PrivateKeys, error) {
	defer reader.Close()

	var keys PrivateKeys

	_ = json.NewDecoder(reader).Decode(&keys)

	return keys, nil
}
