package graph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// CodeIntelStore is an abstraction over enterprise/internal/codeintel.Services for use
// in code graph search. It must not use any enterprise types, and an implementation
// should be registered during enterprise setup with RegisterStore.
type CodeIntelStore interface {
	GetReferences(ctx context.Context, repo types.MinimalRepo, args types.CodeIntelRequestArgs) (_ []types.CodeIntelLocation, err error)
	GetImplementations(ctx context.Context, repo types.MinimalRepo, args types.CodeIntelRequestArgs) (_ []types.CodeIntelLocation, err error)
}

var store CodeIntelStore

// RegisterStore sets the global CodeIntelStore implementation for search, and should only
// be called on initialization.
func RegisterStore(s CodeIntelStore) { store = s }

// Store retrieves the globally registered CodeIntelStore implementation for search, and
// must only be called after initialization. It may return nil if no implementation is
// registered.
func Store() CodeIntelStore { return store }
