package backend

import (
	"context"

	symbolsclient "github.com/sourcegraph/sourcegraph/pkg/symbols/client"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
)

// Symbols backend.
var Symbols = &symbols{}

type symbols struct{}

// ListTags returns symbols in a repository from ctags.
func (symbols) ListTags(ctx context.Context, args protocol.SearchArgs) ([]protocol.Symbol, error) {
	result, err := symbolsclient.DefaultClient.Search(ctx, args)
	if result == nil {
		return nil, err
	}
	return result.Symbols, err
}
