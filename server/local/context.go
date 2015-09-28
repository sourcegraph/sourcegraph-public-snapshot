package local

import "golang.org/x/net/context"

type contextKey int

const (
	configKey contextKey = iota
)

// NewContext creates a new child context for use by local services.
func NewContext(ctx context.Context, conf Config) context.Context {
	return context.WithValue(ctx, configKey, conf)
}

// stores returns the context's Config struct.
func config(ctx context.Context) Config {
	conf, ok := ctx.Value(configKey).(Config)
	if !ok {
		panic("no local service Config set in context")
	}
	return conf
}
