package graph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CodeIntelStore is an abstraction over enterprise/internal/codeintel.Services for use
// in code graph search. It must not use any enterprise types, and an implementation
// should be registered during enterprise setup with RegisterStore.
type CodeIntelStore interface {
	GetDefinitions(context.Context, types.MinimalRepo, types.CodeIntelRequestArgs) ([]types.CodeIntelLocation, error)
	GetReferences(context.Context, types.MinimalRepo, types.CodeIntelRequestArgs) ([]types.CodeIntelLocation, error)
	GetImplementations(context.Context, types.MinimalRepo, types.CodeIntelRequestArgs) ([]types.CodeIntelLocation, error)
}

var store CodeIntelStore = UnimplementedCodeIntelStore{}

// RegisterStore sets the global CodeIntelStore implementation for search, and should only
// be called on initialization.
func RegisterStore(s CodeIntelStore) { store = s }

// Store retrieves the globally registered CodeIntelStore implementation for search, and
// must only be called after initialization.
//
// If no implementation is registered, UnimplementedCodeIntelStore is returned.
func Store() CodeIntelStore { return store }

// UnimplementedCodeIntelStore is the default graph.CodeIntelStore implementation, unless
// RegisterStore is called on the package. All methods return ErrCodeIntelStoreUnimplemented.
type UnimplementedCodeIntelStore struct{}

var ErrCodeIntelStoreUnimplemented = errors.New("code-intel graph store unimplemented")

func (UnimplementedCodeIntelStore) GetDefinitions(context.Context, types.MinimalRepo, types.CodeIntelRequestArgs) ([]types.CodeIntelLocation, error) {
	return nil, ErrCodeIntelStoreUnimplemented
}

func (UnimplementedCodeIntelStore) GetReferences(context.Context, types.MinimalRepo, types.CodeIntelRequestArgs) ([]types.CodeIntelLocation, error) {
	return nil, ErrCodeIntelStoreUnimplemented
}

func (UnimplementedCodeIntelStore) GetImplementations(context.Context, types.MinimalRepo, types.CodeIntelRequestArgs) ([]types.CodeIntelLocation, error) {
	return nil, ErrCodeIntelStoreUnimplemented
}
