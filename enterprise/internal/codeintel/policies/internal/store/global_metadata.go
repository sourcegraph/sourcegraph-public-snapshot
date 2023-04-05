package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// RepoCount returns the total number of policy-selectable repos.
func (s *store) RepoCount(ctx context.Context) (_ int, err error) {
	count, _, err := basestore.ScanFirstInt(s.db.Query(ctx, sqlf.Sprintf(`SELECT SUM(total) FROM repo_statistics`)))
	if err != nil {
		return 0, err
	}

	return count, nil
}
