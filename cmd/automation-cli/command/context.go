package command

import "context"

type contextKey int

const (
	ctxConfigPathKey contextKey = iota
	ctxConfigKey
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

func AttachConfig(ctx context.Context, config Config) context.Context {
	return context.WithValue(ctx, ctxConfigKey, config)
}

func GetConfigFromContext(ctx context.Context) *Config {
	val := ctx.Value(ctxConfigKey)
	if val == nil {
		return nil
	}

	config, ok := val.(Config)
	if !ok {
		return nil
	}

	return &config
}
