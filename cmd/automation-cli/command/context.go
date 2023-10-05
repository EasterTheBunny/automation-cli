package command

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

type StatePaths struct {
	Base        string
	Environment string
}

func CreateStatePaths(base, environment string) (*StatePaths, error) {
	// check if starts with ~/ and replace with home directory
	if strings.HasPrefix(base, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		base = strings.Replace(base, "~", home, 1)
	}

	environment = fmt.Sprintf("%s/%s", base, environment)

	if _, err := os.Stat(environment); os.IsNotExist(err) {
		abs, err := filepath.Abs(environment)
		if err != nil {
			return nil, err
		}

		if err := os.MkdirAll(abs, 0760); err != nil {
			return nil, err
		}
	}

	return &StatePaths{
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
