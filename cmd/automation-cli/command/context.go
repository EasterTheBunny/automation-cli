package command

import (
	"context"

	"github.com/easterthebunny/automation-cli/cmd/automation-cli/config"
)

type contextKey int

const (
	ctxConfigPathKey contextKey = iota
	ctxConfigKey
	ctxKeyConfigKey
)

func AttachConfigPath(ctx context.Context, path string) context.Context {
	return context.WithValue(ctx, ctxConfigPathKey, path)
}

func GetConfigPathFromContext(ctx context.Context) *string {
	val := ctx.Value(ctxConfigPathKey)
	if val == nil {
		return nil
	}

	path, ok := val.(string)
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
