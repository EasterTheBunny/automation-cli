package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	ErrReadConfig  = fmt.Errorf("failed to read config")
	ErrWriteConfig = fmt.Errorf("failed to write config")
)

func ensureExists(path, filename string) (string, error) {
	configPath := fmt.Sprintf("%s/%s", path, filename)

	_, err := os.Stat(configPath)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(configPath), fs.ModePerm); err != nil {
			return "", err
		}

		file, err := os.OpenFile(configPath, os.O_CREATE, 0640)
		if err != nil {
			return "", err
		}

		file.Close()
	}

	return configPath, nil
}

func readConfig[T any](vpr *viper.Viper, path string) (*T, error) {
	if err := vpr.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			if err := vpr.WriteConfigAs(path); err != nil {
				return nil, fmt.Errorf("%w: %s", ErrWriteConfig, err.Error())
			}
		}
	}

	var conf T

	if err := vpr.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("%w, %s", ErrReadConfig, err.Error())
	}

	return &conf, nil
}

func SaveConfig(path string) error {
	if err := viper.WriteConfigAs(fmt.Sprintf("%s/config.json", path)); err != nil {
		return fmt.Errorf("%w: %s", ErrWriteConfig, err.Error())
	}

	return nil
}

func SaveViperConfig(vpr *viper.Viper, path string) error {
	if err := vpr.WriteConfigAs(fmt.Sprintf("%s/config.json", path)); err != nil {
		return fmt.Errorf("%w: %s", ErrWriteConfig, err.Error())
	}

	return nil
}
