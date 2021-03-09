package graphql

import (
	"github.com/sourcegraph/src-cli/internal/api"
)

type commonBackend struct {
	client             api.Client
	useGzipCompression bool
}

func (b *commonBackend) newRequest(query string, vars map[string]interface{}) api.Request {
	if b.useGzipCompression {
		return b.client.NewGzippedRequest(query, vars)
	}
	return b.client.NewRequest(query, vars)
}
