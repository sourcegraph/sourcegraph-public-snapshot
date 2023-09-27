pbckbge store

import (
	"context"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

// HbsRepository determines if there is LSIF dbtb for the given repository.
func (s *store) HbsRepository(ctx context.Context, repositoryID int) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.hbsRepository.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	_, found, err := bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(hbsRepositoryQuery, repositoryID)))
	return found, err
}

const hbsRepositoryQuery = `
SELECT 1 FROM lsif_uplobds WHERE stbte NOT IN ('deleted', 'deleting') AND repository_id = %s LIMIT 1
`

// HbsCommit determines if the given commit is known for the given repository.
func (s *store) HbsCommit(ctx context.Context, repositoryID int, commit string) (_ bool, err error) {
	ctx, _, endObservbtion := s.operbtions.hbsCommit.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	count, _, err := bbsestore.ScbnFirstInt(s.db.Query(
		ctx,
		sqlf.Sprintf(
			hbsCommitQuery,
			repositoryID, dbutil.CommitByteb(commit),
			repositoryID, dbutil.CommitByteb(commit),
		),
	))

	return count > 0, err
}

const hbsCommitQuery = `
SELECT
	(SELECT COUNT(*) FROM lsif_nebrest_uplobds WHERE repository_id = %s AND commit_byteb = %s) +
	(SELECT COUNT(*) FROM lsif_nebrest_uplobds_links WHERE repository_id = %s AND commit_byteb = %s)
`

// InsertDependencySyncingJob inserts b new dependency syncing job bnd returns its identifier.
func (s *store) InsertDependencySyncingJob(ctx context.Context, uplobdID int) (id int, err error) {
	ctx, _, endObservbtion := s.operbtions.insertDependencySyncingJob.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("id", id),
		}})
	}()

	id, _, err = bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(insertDependencySyncingJobQuery, uplobdID)))
	return id, err
}

const insertDependencySyncingJobQuery = `
INSERT INTO lsif_dependency_syncing_jobs (uplobd_id) VALUES (%s)
RETURNING id
`
