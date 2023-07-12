package batches

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	bbcs "github.com/sourcegraph/sourcegraph/internal/batches/sources/bitbucketcloud"
	bstore "github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type externalForkNameMigrator struct {
	store     *basestore.Store
	batchSize int
}

func NewExternalForkNameMigrator(store *basestore.Store, batchSize int) *externalForkNameMigrator {
	return &externalForkNameMigrator{
		store:     store,
		batchSize: batchSize,
	}
}

var _ oobmigration.Migrator = &externalForkNameMigrator{}

func (m *externalForkNameMigrator) ID() int                 { return 21 }
func (m *externalForkNameMigrator) Interval() time.Duration { return time.Second * 5 }

// Progress returns the percentage (ranged [0, 1]) of changesets published to a fork on
// Bitbucket Server or Bitbucket Cloud that have not had `external_fork_name` set on their
// DB record.
func (m *externalForkNameMigrator) Progress(ctx context.Context, _ bool) (float64, error) {
	progress, _, err := basestore.ScanFirstFloat(m.store.Query(ctx, sqlf.Sprintf(externalForkNameMigratorProgressQuery)))
	return progress, err
}

// This query compares the count of migrated changesets, which should have
// external_fork_name set, vs. the total count of changesets on a fork on Bitbucket Server
// or Cloud.
const externalForkNameMigratorProgressQuery = `
SELECT
	CASE total.count WHEN 0 THEN 1 ELSE
		CAST(migrated.count AS float) / CAST(total.count AS float)
	END
FROM
(SELECT COUNT(1) AS count FROM changesets
	WHERE external_fork_name IS NOT NULL
	AND external_fork_namespace IS NOT NULL
	AND external_deleted_at IS NULL
	AND external_service_type IN ('bitbucketServer', 'bitbucketCloud')) migrated,
(SELECT COUNT(1) AS count FROM changesets
	WHERE external_fork_namespace IS NOT NULL
	AND external_deleted_at IS NULL
	AND external_service_type IN ('bitbucketServer', 'bitbucketCloud')) total;`

func (m *externalForkNameMigrator) Up(ctx context.Context) (err error) {
	css, err := func() (css []*btypes.Changeset, err error) {
		rows, err := m.store.Query(ctx, sqlf.Sprintf(forkChangesetsSelectQuery, sqlf.Join(bstore.ChangesetColumns, ","), m.batchSize))
		if err != nil {
			return nil, err
		}

		defer func() { err = basestore.CloseRows(rows, err) }()

		for rows.Next() {
			var c btypes.Changeset
			if err = bstore.ScanChangeset(&c, rows); err != nil {
				return nil, err
			}
			css = append(css, &c)
		}

		return css, nil
	}()
	if err != nil {
		return err
	}

	getforkName := func(cs *btypes.Changeset) string {
		meta := cs.Metadata
		switch m := meta.(type) {
		// We only have the fork name available on the changeset metadata for Bitbucket
		// Server and Bitbucket Cloud. We live-backfill the fork name for changesets on
		// GitHub and GitLab the next time they are processed by the reconciler.
		case *bitbucketserver.PullRequest:
			return m.FromRef.Repository.Slug
		case *bbcs.AnnotatedPullRequest:
			return m.Source.Repo.Name
		default:
			return ""
		}
	}

	for _, cs := range css {
		forkName := getforkName(cs)
		if forkName == "" {
			continue
		}
		if err := m.store.Exec(ctx, sqlf.Sprintf(setForkNameUpdateQuery, forkName, cs.ID)); err != nil {
			return err
		}
	}

	return nil
}

const forkChangesetsSelectQuery = `
SELECT %s FROM changesets
	WHERE external_fork_namespace IS NOT NULL
	AND external_fork_name IS NULL
	AND external_deleted_at IS NULL
	AND external_service_type IN ('bitbucketServer', 'bitbucketCloud')
	ORDER BY id LIMIT %s FOR UPDATE;`

const setForkNameUpdateQuery = `
	UPDATE changesets SET external_fork_name = %s WHERE id = %s;`

func (m *externalForkNameMigrator) Down(ctx context.Context) (err error) {
	// Non-destructive
	return nil
}
