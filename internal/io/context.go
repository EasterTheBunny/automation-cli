package io

import "context"

type ctxKey int

const (
	environmentContextKey ctxKey = iota
)

func ContextWithEnvironment(ctx context.Context, env Environment) context.Context {
	return context.WithValue(ctx, environmentContextKey, env)
}

func EnvironmentFromContext(ctx context.Context) *Environment {
	val := ctx.Value(environmentContextKey)
	if val == nil {
		return nil
	}

	env, ok := val.(Environment)
	if !ok {
		return nil
	}

	return &env
}
