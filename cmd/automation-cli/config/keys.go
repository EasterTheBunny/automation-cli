package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type PrivateKeyConfig struct {
	Keys []Key `json:"keys"`
}

type Key struct {
	Alias   string `json:"alias"`
	Value   string `json:"value"`
	Address string `json:"address"`
}

func GetPrivateKeyConfig(path string) (*PrivateKeyConfig, error) {
	filePath := fmt.Sprintf("%s/keys.json", path)

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0640)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var keys PrivateKeyConfig

	if len(fileBytes) == 0 {
		return &keys, nil
	}

	if err := json.Unmarshal(fileBytes, &keys); err != nil {
		return nil, fmt.Errorf("%w, %s", ErrReadConfig, err.Error())
	}

	return &keys, nil
}

func SavePrivateKeyConfig(path string, conf *PrivateKeyConfig) error {
	filePath := fmt.Sprintf("%s/keys.json", path)

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}

	defer file.Close()

	fileBytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	_, err = file.Write(fileBytes)

	return err
}
