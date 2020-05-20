package api

import (
	"context"
	"errors"

	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
)

// CodeIntelAPI is the main interface into precise code intelligence data.
type CodeIntelAPI interface {
	// FindClosestDumps returns the set of dumps that can most accurately answer code intelligence
	// queries for the given file. These dump IDs should be subsequently passed to invocations of
	// Definitions, References, and Hover.
	FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) ([]db.Dump, error)

	// Definitions returns the list of source locations that define the symbol at the given position.
	// This may include remote definitions if the remote repository is also indexed.
	Definitions(ctx context.Context, file string, line, character, uploadID int) ([]ResolvedLocation, error)

	// References returns the list of source locations that reference the symbol at the given position.
	// This may include references from other dumps and repositories.
	References(ctx context.Context, repositoryID int, commit string, limit int, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error)

	// Hover returns the hover text and range for the symbol at the given position.
	Hover(ctx context.Context, file string, line, character, uploadID int) (string, bundles.Range, bool, error)
}

type codeIntelAPI struct {
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
	gitserverClient     gitserver.Client
}

var _ CodeIntelAPI = &codeIntelAPI{}

var ErrMissingDump = errors.New("no dump")

func New(db db.DB, bundleManagerClient bundles.BundleManagerClient, gitserverClient gitserver.Client) CodeIntelAPI {
	return &codeIntelAPI{
		db:                  db,
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
	}
}
