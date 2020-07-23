package store

import (
	"context"
	"database/sql"
	"sort"

	"github.com/keegancsmith/sqlf"
)

// scanCommits scans pairs of commits/parentCommits from the return value of `*store.query`.
func scanCommits(rows *sql.Rows, queryErr error) (_ map[string][]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = closeRows(rows, err) }()

	commits := map[string][]string{}
	for rows.Next() {
		var commit string
		var parentCommit *string
		if err := rows.Scan(&commit, &parentCommit); err != nil {
			return nil, err
		}

		if _, ok := commits[commit]; !ok {
			commits[commit] = nil
		}

		if parentCommit != nil {
			commits[commit] = append(commits[commit], *parentCommit)
		}
	}

	return commits, nil
}

// HasCommit determines if the given commit is known for the given repository.
func (s *store) HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error) {
	count, _, err := scanFirstInt(s.query(ctx, sqlf.Sprintf(`
		SELECT COUNT(*)
		FROM lsif_commits
		WHERE repository_id = %s and commit = %s
		LIMIT 1
	`, repositoryID, commit)))

	return count > 0, err
}

// UpdateCommits upserts commits/parent-commit relations for the given repository ID.
func (s *store) UpdateCommits(ctx context.Context, repositoryID int, commits map[string][]string) error {
	if len(commits) == 0 {
		return nil
	}

	var qs []*sqlf.Query
	for commit := range commits {
		qs = append(qs, sqlf.Sprintf("%s", commit))
	}

	knownCommits, err := scanCommits(s.query(
		ctx,
		sqlf.Sprintf(`
			SELECT "commit", parent_commit
			FROM lsif_commits
			WHERE repository_id = %s AND "commit" IN (%s)
		`, repositoryID, sqlf.Join(qs, ",")),
	))
	if err != nil {
		return err
	}

	unknownCommits := map[string][]string{}
	for commit, parentCommits := range commits {
		if knownParents, ok := knownCommits[commit]; ok {
			// Filter out any known parents. Only keep this commit in the map
			// if we have at least one new unknown parent, otherwise we'll end
			// up inserting the `(commit, NULL)` which will pollute the table.
			if d := diff(parentCommits, knownParents); len(d) > 0 {
				unknownCommits[commit] = d
			}
		} else {
			// New commit, all parents unknown
			unknownCommits[commit] = parentCommits
		}
	}

	if len(unknownCommits) == 0 {
		return nil
	}

	// Make the order in which we construct the values for insertion determinstic.
	// We want this to happen because many workers/api-servers can be inserting
	// commits for the same repository. Having them inserted in random order may
	// cause a deadlock to occur where two threads are writing the same tuples in
	// different orders: e.g. A writes t1 then t2, and B writes t2 then t1. If we
	// always write in the same order, then such a deadlock is impossible.
	var keys []string
	for commit, parentCommits := range unknownCommits {
		keys = append(keys, commit)
		sort.Strings(parentCommits)
	}
	sort.Strings(keys)

	var rows []*sqlf.Query
	for _, commit := range keys {
		for _, parent := range unknownCommits[commit] {
			rows = append(rows, sqlf.Sprintf("(%d, %s, %s)", repositoryID, commit, parent))
		}

		if len(unknownCommits[commit]) == 0 {
			// Insert a commit even if its parent is not known
			rows = append(rows, sqlf.Sprintf("(%d, %s, NULL)", repositoryID, commit))
		}
	}

	return s.queryForEffect(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_commits (repository_id, "commit", parent_commit)
		VALUES %s
		ON CONFLICT DO NOTHING
	`, sqlf.Join(rows, ",")))
}

// diff returns a slice containing the elements of left not present in right.
func diff(left, right []string) []string {
	rightSet := map[string]struct{}{}
	for _, v := range right {
		rightSet[v] = struct{}{}
	}

	var diff []string
	for _, v := range left {
		if _, ok := rightSet[v]; !ok {
			diff = append(diff, v)
		}
	}

	return diff
}
