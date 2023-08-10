package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RecentContributionSignalStore interface {
	AddCommit(ctx context.Context, commit Commit) error
	FindRecentAuthors(ctx context.Context, repoID api.RepoID, path string) ([]RecentContributorSummary, error)
	ClearSignals(ctx context.Context, repoID api.RepoID) error
	WithTransact(context.Context, func(store RecentContributionSignalStore) error) error
}

func RecentContributionSignalStoreWith(other basestore.ShareableStore) RecentContributionSignalStore {
	return &recentContributionSignalStore{Store: basestore.NewWithHandle(other.Handle())}
}

type Commit struct {
	RepoID       api.RepoID
	AuthorName   string
	AuthorEmail  string
	Timestamp    time.Time
	CommitSHA    string
	FilesChanged []string
}

type RecentContributorSummary struct {
	AuthorName        string
	AuthorEmail       string
	ContributionCount int
}

type recentContributionSignalStore struct {
	*basestore.Store
}

func (s *recentContributionSignalStore) WithTransact(ctx context.Context, f func(store RecentContributionSignalStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(RecentContributionSignalStoreWith(tx))
	})
}

func (s *recentContributionSignalStore) With(other basestore.ShareableStore) *recentContributionSignalStore {
	return &recentContributionSignalStore{Store: s.Store.With(other)}
}

const commitAuthorInsertFmtstr = `
	WITH already_exists (id) AS (
		SELECT id
		FROM commit_authors
		WHERE name = %s
		AND email = %s
	),
	need_to_insert (id) AS (
		INSERT INTO commit_authors (name, email)
		VALUES (%s, %s)
		ON CONFLICT (name, email) DO NOTHING
		RETURNING id
	)
	SELECT id FROM already_exists
	UNION ALL
	SELECT id FROM need_to_insert
`

// ensureAuthor makes sure the that commit author designated by name and email
// exists in the `commit_authors` table, and returns its ID.
func (s *recentContributionSignalStore) ensureAuthor(ctx context.Context, commit Commit) (int, error) {
	db := s.Store
	var authorID int
	if err := db.QueryRow(
		ctx,
		sqlf.Sprintf(
			commitAuthorInsertFmtstr,
			commit.AuthorName,
			commit.AuthorEmail,
			commit.AuthorName,
			commit.AuthorEmail,
		),
	).Scan(&authorID); err != nil {
		return 0, err
	}
	return authorID, nil
}

// ensureRepoPaths takes paths of files changed in the given commit
// and makes sure they all exist in the database (alongside with their ancestor paths)
// as per the schema.
//
// The operation makes a number of queries to the database that is comparable
// to the size of the given file tree. In other words, every directory mentioned
// in the `commit.FilesChanged` (including parents and ancestors) will be queried
// or inserted with a single query (no repetitions though).
// Optimizing this into fewer queries seems to make the implementation very hard to read.
//
// The result int slice is guaranteed to be in order corresponding to the order
// of `commit.FilesChanged`.
func (s *recentContributionSignalStore) ensureRepoPaths(ctx context.Context, commit Commit) ([]int, error) {
	return ensureRepoPaths(ctx, s.Store, commit.FilesChanged, commit.RepoID)
}

const insertRecentContributorSignalFmtstr = `
	INSERT INTO own_signal_recent_contribution (
		commit_author_id,
		changed_file_path_id,
		commit_timestamp,
		commit_id
	) VALUES (%s, %s, %s, %s)
`

const clearSignalsFmtstr = `
    WITH rps AS (
        SELECT id FROM repo_paths WHERE repo_id = %s
    )
    DELETE FROM %s
    WHERE changed_file_path_id IN (SELECT * FROM rps)
`

func (s *recentContributionSignalStore) ClearSignals(ctx context.Context, repoID api.RepoID) error {
	tables := []string{"own_signal_recent_contribution", "own_aggregate_recent_contribution"}

	for _, table := range tables {
		if err := s.Exec(ctx, sqlf.Sprintf(clearSignalsFmtstr, repoID, sqlf.Sprintf(table))); err != nil {
			return errors.Wrapf(err, "table: %s", table)
		}
	}
	return nil
}

// AddCommit inserts a recent contribution signal for each file changed by given commit.
//
// As per schema, `commit_id` is the git sha stored as bytea.
// This is used for the purpose of removing old recent contributor signals.
// The aggregate signals in `own_aggregate_recent_contribution` are updated atomically
// for each new signal appearing in `own_signal_recent_contribution` by using
// a trigger: `update_own_aggregate_recent_contribution`.
func (s *recentContributionSignalStore) AddCommit(ctx context.Context, commit Commit) (err error) {
	// Get or create commit author:
	authorID, err := s.ensureAuthor(ctx, commit)
	if err != nil {
		return errors.Wrap(err, "cannot insert commit author")
	}
	// Get or create necessary repo paths:
	pathIDs, err := s.ensureRepoPaths(ctx, commit)
	if err != nil {
		return errors.Wrap(err, "cannot insert repo paths")
	}
	// Insert individual signals into own_signal_recent_contribution:
	for _, pathID := range pathIDs {
		q := sqlf.Sprintf(insertRecentContributorSignalFmtstr,
			authorID,
			pathID,
			commit.Timestamp,
			dbutil.CommitBytea(commit.CommitSHA),
		)
		err = s.Exec(ctx, q)
		if err != nil {
			return err
		}
	}
	return nil
}

const findRecentContributorsFmtstr = `
	SELECT a.name, a.email, g.contributions_count
	FROM commit_authors AS a
	INNER JOIN own_aggregate_recent_contribution AS g
	ON a.id = g.commit_author_id
	INNER JOIN repo_paths AS p
	ON p.id = g.changed_file_path_id
	WHERE p.repo_id = %s
	AND p.absolute_path = %s
	ORDER BY 3 DESC
`

// FindRecentAuthors returns all recent authors for given `repoID` and `path`.
// Since the recent contributor signal aggregate is computed within `AddCommit`
// This just looks up `own_aggregate_recent_contribution` associated with given
// repo and path, and pulls all the related authors.
// Notes:
// - `path` has not forward slash at the beginning, example: "dir1/dir2/file.go", "file2.go".
// - Empty string `path` designates repo root (so all contributions for the whole repo).
// - TODO: Need to support limit & offset here.
func (s *recentContributionSignalStore) FindRecentAuthors(ctx context.Context, repoID api.RepoID, path string) ([]RecentContributorSummary, error) {
	q := sqlf.Sprintf(findRecentContributorsFmtstr, repoID, path)

	contributionsScanner := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (RecentContributorSummary, error) {
		var rcs RecentContributorSummary
		if err := scanner.Scan(&rcs.AuthorName, &rcs.AuthorEmail, &rcs.ContributionCount); err != nil {
			return RecentContributorSummary{}, err
		}
		return rcs, nil
	})

	contributions, err := contributionsScanner(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}
	return contributions, nil
}
