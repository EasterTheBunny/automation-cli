package node

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func create(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), fs.ModePerm); err != nil {
		return nil, err
	}

	return os.Create(path)
}

func writeCredentials(path string) error {
	apiFile, err := create(fmt.Sprintf("%s/chainlink-node-api", path))
	if err != nil {
		return err
	}

	defer apiFile.Close()

	_, _ = apiFile.WriteString(DefaultChainlinkNodeLogin)
	_, _ = apiFile.WriteString("\n")
	_, _ = apiFile.WriteString(DefaultChainlinkNodePassword)

	passwordFile, err := create(fmt.Sprintf("%s/chainlink-node-password", path))
	if err != nil {
		return err
	}

	defer passwordFile.Close()

	_, _ = passwordFile.WriteString(DefaultChainlinkNodePassword)

	return nil
}

func writeFile(path, data string) error {
	file, err := create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
