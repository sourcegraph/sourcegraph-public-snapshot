package resolvers

import (
	"context"
	"sync"
)

type userSymbolsResolver struct {
	once   sync.Once
	result []string
	err    error
}

func (r *CodeGraphPersonNodeResolver) symbols(ctx context.Context) ([]string, error) {
	// Get a list of all symbols (so we know which files/locations to blame).

	return []string{"mysymbol1", "mysymbol2", "mysymbol3"}, nil
}

func (r *CodeGraphPersonNodeResolver) Symbols(ctx context.Context) ([]string, error) {
	r.once.Do(func() { r.result, r.err = r.symbols(ctx) })
	return r.result, r.err
}
