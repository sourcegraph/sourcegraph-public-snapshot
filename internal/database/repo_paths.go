package database

import (
	"context"
	"path"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const pathInsertFmtstr = `
	WITH already_exists (id) AS (
		SELECT id
		FROM repo_paths
		WHERE repo_id = %s
		AND absolute_path = %s
	),
	need_to_insert (id) AS (
		INSERT INTO repo_paths (repo_id, absolute_path, parent_id)
		SELECT %s, %s, %s
		WHERE NOT EXISTS (
			SELECT
			FROM repo_paths
			WHERE repo_id = %s
			AND absolute_path = %s
		)
		ON CONFLICT (repo_id, absolute_path) DO NOTHING
		RETURNING id
	)
	SELECT id FROM already_exists
	UNION ALL
	SELECT id FROM need_to_insert
`

// ensureRepoPaths takes paths and makes sure they all exist in the database
// (alongside with their ancestor paths) as per the schema.
//
// The operation makes a number of queries to the database that is comparable to
// the size of the given file tree. In other words, every directory mentioned in
// the `files` (including parents and ancestors) will be queried or inserted with
// a single query (no repetitions though). Optimizing this into fewer queries
// seems to make the implementation very hard to read.
//
// The result int slice is guaranteed to be in order corresponding to the order
// of `files`.
func ensureRepoPaths(ctx context.Context, db *basestore.Store, files []string, repoID api.RepoID) ([]int, error) {
	// Compute all the ancestor paths for all given files.
	var paths []string
	for _, file := range files {
		for p := file; p != "."; p = path.Dir(p) {
			paths = append(paths, p)
		}
	}
	// Add empty string which references the repo root directory.
	paths = append(paths, "")
	// Reverse paths so we start at the root.
	for i := 0; i < len(paths)/2; i++ {
		j := len(paths) - i - 1
		paths[i], paths[j] = paths[j], paths[i]
	}
	// Remove duplicates from paths, to avoid extra query, especially if many files
	// within the same directory structure are referenced.
	seen := make(map[string]bool)
	j := 0
	for i := 0; i < len(paths); i++ {
		if !seen[paths[i]] {
			seen[paths[i]] = true
			paths[j] = paths[i]
			j++
		}
	}
	paths = paths[:j]
	// Insert all directories one query each and note the IDs.
	ids := map[string]int{}
	for _, p := range paths {
		var parentID *int
		parent := path.Dir(p)
		if parent == "." {
			parent = ""
		}
		if id, ok := ids[parent]; p != "" && ok {
			parentID = &id
		} else if p != "" {
			return nil, errors.Newf("cannot find parent id of %q: this is a bug", p)
		}
		r := db.QueryRow(ctx, sqlf.Sprintf(pathInsertFmtstr, repoID, p, repoID, p, parentID, repoID, p))
		var id int
		if err := r.Scan(&id); err != nil {
			return nil, errors.Wrapf(err, "failed to insert or retrieve %q", p)
		}
		ids[p] = id
	}
	// Return the IDs of inserted files changed, in order of `files`.
	fIDs := make([]int, len(files))
	for i, f := range files {
		id, ok := ids[f]
		if !ok {
			return nil, errors.Newf("cannot find id of %q which should have been inserted, this is a bug", f)
		}
		fIDs[i] = id
	}
	return fIDs, nil
}
