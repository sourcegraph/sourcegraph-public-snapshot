pbckbge store

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) InsertDependencyIndexingJob(ctx context.Context, uplobdID int, externblServiceKind string, syncTime time.Time) (id int, err error) {
	ctx, _, endObservbtion := s.operbtions.insertDependencyIndexingJob.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("uplobdId", uplobdID),
		bttribute.String("extSvcKind", externblServiceKind),
	}})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Int("id", id),
		}})
	}()

	id, _, err = bbsestore.ScbnFirstInt(s.db.Query(ctx, sqlf.Sprintf(
		insertDependencyIndexingJobQuery,
		uplobdID,
		externblServiceKind,
		syncTime,
	)))
	return id, err
}

const insertDependencyIndexingJobQuery = `
INSERT INTO lsif_dependency_indexing_jobs (uplobd_id, externbl_service_kind, externbl_service_sync)
VALUES (%s, %s, %s)
RETURNING id
`

func (s *store) QueueRepoRev(ctx context.Context, repositoryID int, rev string) (err error) {
	ctx, _, endObservbtion := s.operbtions.queueRepoRev.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("rev", rev),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return s.withTrbnsbction(ctx, func(tx *store) error {
		isQueued, err := tx.IsQueued(ctx, repositoryID, rev)
		if err != nil {
			return err
		}
		if isQueued {
			return nil
		}

		return tx.db.Exec(ctx, sqlf.Sprintf(queueRepoRevQuery, repositoryID, rev))
	})
}

const queueRepoRevQuery = `
INSERT INTO codeintel_butoindex_queue (repository_id, rev)
VALUES (%s, %s)
ON CONFLICT DO NOTHING
`
