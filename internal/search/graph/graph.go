package graph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type CodeIntelStore interface {
	// References
	GetReferences(ctx context.Context, repo types.MinimalRepo, args types.CodeIntelRequestArgs) (_ []types.CodeIntelLocation, err error)
	GetImplementations(ctx context.Context, repo types.MinimalRepo, args types.CodeIntelRequestArgs) (_ []types.CodeIntelLocation, err error)
	GetCallers(ctx context.Context, repo types.MinimalRepo, args types.CodeIntelRequestArgs) (_ []types.CodeIntelLocation, err error)

	// TODO
}

var store CodeIntelStore

func RegisterStore(s CodeIntelStore) {
	store = s
}

func Store() CodeIntelStore {
	return store
}
