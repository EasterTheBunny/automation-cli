package io

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrFSOpFailure = fmt.Errorf("fs access failure")
)

const (
	FileCreateMode      = 0640
	DirectoryCreateMode = 0760
)

type Root string

func (r Root) Path() (string, error) {
	path := string(r)

	// check if starts with ~/ and replace with home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf(
				"%w: failed to get user home directory using prefix '~/': %s",
				ErrFSOpFailure,
				err.Error())
		}

		path = strings.Replace(path, "~", home, 1)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, DirectoryCreateMode); err != nil {
			return "", fmt.Errorf("%w: failed to make directories: %s", ErrFSOpFailure, err.Error())
		}
	}

	return path, nil
}

// MustRead opens the provided filename on the root path as read only. If the file does not exist, it is created. Panics
// on error.
func (r Root) MustRead(filename string) io.ReadCloser {
	rootPath, err := r.Path()
	if err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("%s/%s", rootPath, filename)

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, FileCreateMode)
	if err != nil {
		panic(err)
	}

	return file
}

// MustWrite opens the provided filename on the root path as writable. If the file does not exist, it is created. Panics
// on error. Overwrites the file if it exists
func (r Root) MustWrite(filename string) io.WriteCloser {
	rootPath, err := r.Path()
	if err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("%s/%s", rootPath, filename)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, FileCreateMode)
	if err != nil {
		panic(err)
	}

	return file
}

type Environment struct {
	Root
	Name string
}

// Path returns the absolute filesystem path to the environment directory. If the directory does not exist, it is
// created.
func (e Environment) Path() (string, error) {
	rootPath, err := e.Root.Path()
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("%s/%s", rootPath, e.Name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, DirectoryCreateMode); err != nil {
			return "", fmt.Errorf("%w: failed to make directories: %s", ErrFSOpFailure, err.Error())
		}
	}

	return path, nil
}

// MustRead opens the provided filename on the root path as read only. If the file does not exist, it is created. Panics
// on error.
func (e Environment) MustRead(filename string) io.ReadCloser {
	rootPath, err := e.Path()
	if err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("%s/%s", rootPath, filename)

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, FileCreateMode)
	if err != nil {
		panic(err)
	}

	return file
}

// MustWrite opens the provided filename on the root path as writable. If the file does not exist, it is created. Panics
// on error. Overwrites the file if it exists
func (e Environment) MustWrite(filename string) io.WriteCloser {
	rootPath, err := e.Path()
	if err != nil {
		panic(err)
	}

	filePath := fmt.Sprintf("%s/%s", rootPath, filename)

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, FileCreateMode)
	if err != nil {
		panic(err)
	}

	return file
}

func (e Environment) Delete() error {
	path, err := e.Path()
	if err != nil {
		return err
	}

	return os.RemoveAll(path)
}
