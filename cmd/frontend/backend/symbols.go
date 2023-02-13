package backend

import (
	"context"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	symbolsclient "github.com/sourcegraph/sourcegraph/internal/symbols"
)

// Symbols backend.
var Symbols = &symbols{}

type symbols struct{}

// ListTags returns symbols in a repository from ctags.
func (symbols) ListTags(ctx context.Context, metrics *grpc_prometheus.ClientMetrics, args search.SymbolsParameters) (result.Symbols, error) {
	symbols, err := symbolsclient.DefaultClient(metrics).Search(ctx, args)
	if err != nil {
		return nil, err
	}
	for i := range symbols {
		symbols[i].Line += 1 // callers expect 1-indexed lines
	}
	return symbols, nil
}
