package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
)

type contextKey int

const (
	ctxStateKey contextKey = iota
	ctxConfigPathKey
	ctxConfigKey
	ctxKeyConfigKey
)

var (
	ErrFSOpFailure = fmt.Errorf("failed to perform operation on file system")
)

// StatePaths is a collection of file paths to configurations.
type StatePaths struct {
	// Base is the file path to the base configuration.
	Base string
	// Environment is the file path to the environment configuration and overrides.
	Environment string
}

// CreateStatePaths ensures that the provided base and environment paths exist and will create them if they do not.
// Returns file permissions related errors.
func CreateStatePaths(base, environment string) (StatePaths, error) {
	// check if starts with ~/ and replace with home directory
	if strings.HasPrefix(base, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return StatePaths{}, fmt.Errorf(
				"%w: failed to get user home directory using prefix '~/': %s",
				ErrFSOpFailure,
				err.Error())
		}

		base = strings.Replace(base, "~", home, 1)
	}

	environment = fmt.Sprintf("%s/%s", base, environment)

	if _, err := os.Stat(environment); os.IsNotExist(err) {
		abs, err := filepath.Abs(environment)
		if err != nil {
			return StatePaths{}, fmt.Errorf(
				"%w: failed to get absolute path for environment: %s",
				ErrFSOpFailure,
				err.Error())
		}

		if err := os.MkdirAll(abs, 0760); err != nil {
			return StatePaths{}, fmt.Errorf("%w: failed to make directories: %s", ErrFSOpFailure, err.Error())
		}
	}

	return StatePaths{
		Base:        base,
		Environment: environment,
	}, nil
}

func AttachPaths(ctx context.Context, paths StatePaths) context.Context {
	return context.WithValue(ctx, ctxStateKey, paths)
}

func GetPathsFromContext(ctx context.Context) *StatePaths {
	val := ctx.Value(ctxStateKey)
	if val == nil {
		return nil
	}

	path, ok := val.(StatePaths)
	if !ok {
		return nil
	}

	return &path
}

func AttachConfig(ctx context.Context, conf config.Config) context.Context {
	return context.WithValue(ctx, ctxConfigKey, conf)
}

func GetConfigFromContext(ctx context.Context) *config.Config {
	val := ctx.Value(ctxConfigKey)
	if val == nil {
		return nil
	}

	config, ok := val.(config.Config)
	if !ok {
		return nil
	}

	return &config
}

func AttachKeyConfig(ctx context.Context, conf config.PrivateKeyConfig) context.Context {
	return context.WithValue(ctx, ctxKeyConfigKey, conf)
}

func GetKeyConfigFromContext(ctx context.Context) *config.PrivateKeyConfig {
	val := ctx.Value(ctxKeyConfigKey)
	if val == nil {
		return nil
	}

	config, ok := val.(config.PrivateKeyConfig)
	if !ok {
		return nil
	}

	return &config
}
