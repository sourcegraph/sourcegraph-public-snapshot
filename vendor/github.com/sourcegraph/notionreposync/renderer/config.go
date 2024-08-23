package renderer

import (
	"context"
)

// config represents the configuration of the renderer.
type config struct {
	ctx          context.Context
	links        LinkResolver
	debugHandler func(string)
}

// newConfig returns a new default Config, such that all values have a usable
// default. The provided context should be used for all internal operations.
func newConfig(ctx context.Context) config {
	return config{
		ctx:   ctx,
		links: noopLinkResolver{},
	}
}

// Option applies configuration of the renderer. It is intentionally unexported
// such that only this package can export Options.
type Option interface{ setConfig(*config) }

type optionFunc func(*config)

func (o optionFunc) setConfig(c *config) { o(c) }

// WithLinkResolver configures a LinkResolver to use. Otherwise, a default no-op
// one is used that uses links as-is.
func WithLinkResolver(links LinkResolver) Option {
	return optionFunc(func(c *config) {
		c.links = links
	})
}

func WithDebugHandler(fn func(line string)) Option {
	return optionFunc(func(c *config) {
		c.debugHandler = fn
	})
}
