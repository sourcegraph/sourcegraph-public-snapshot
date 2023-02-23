package codeowners

import (
	"context"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

// Ownership associated a ResolvedOwner with a single Rule that matched a file path for ownership.
type Ownership struct {
	Owner ResolvedOwner
	Rule  *codeownerspb.Rule
}

// Graph serves ownership information as it is for a specific codebase
// at a specific version.
type Graph interface {
	// FindOwners returns ownership for given file path in context of a codebase
	// that this OwnershipGraph represents.
	FindOwners(ctx context.Context, filePath string) ([]Ownership, error)
}

type NilGraph struct{}

func (g NilGraph) FindOwners(context.Context, string) ([]Ownership, error) { return nil, nil }
