package api

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/bundles"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/db"
)

type CodeIntelAPI interface {
	FindClosestDumps(repositoryID int, commit, file string) ([]db.Dump, error)
	Definitions(file string, line, character, uploadID int) ([]ResolvedLocation, error)
	References(repositoryID int, commit string, limit int, cursor Cursor) ([]ResolvedLocation, Cursor, bool, error)
	Hover(file string, line, character, uploadID int) (string, bundles.Range, bool, error)
}

type codeIntelAPI struct {
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
}

var _ CodeIntelAPI = &codeIntelAPI{}

var ErrMissingDump = fmt.Errorf("no dump")

// TODO - use in tests?
func New(db db.DB, bundleManagerClient bundles.BundleManagerClient) CodeIntelAPI {
	return &codeIntelAPI{
		db:                  db,
		bundleManagerClient: bundleManagerClient,
	}
}
