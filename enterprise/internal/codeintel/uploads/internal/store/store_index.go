package store

import (
	"context"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// ReindexUploads reindexes uploads matching the given filter criteria.
func (s *store) ReindexUploads(ctx context.Context, opts shared.ReindexUploadsOptions) (err error) {
	ctx, _, endObservation := s.operations.reindexUploads.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", opts.RepositoryID),
		log.String("states", strings.Join(opts.States, ",")),
		log.String("term", opts.Term),
		log.Bool("visibleAtTip", opts.VisibleAtTip),
	}})
	defer endObservation(1, observation.Args{})

	var conds []*sqlf.Query

	if opts.RepositoryID != 0 {
		conds = append(conds, sqlf.Sprintf("u.repository_id = %s", opts.RepositoryID))
	}
	if opts.Term != "" {
		conds = append(conds, makeSearchCondition(opts.Term))
	}
	if len(opts.States) > 0 {
		conds = append(conds, makeStateCondition(opts.States))
	}
	if opts.VisibleAtTip {
		conds = append(conds, sqlf.Sprintf("EXISTS ("+visibleAtTipSubselectQuery+")"))
	}

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.db))
	if err != nil {
		return err
	}
	conds = append(conds, authzConds)

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	unset, _ := tx.db.SetLocal(ctx, "codeintel.lsif_uploads_audit.reason", "direct reindex by filter criteria request")
	defer unset(ctx)

	err = tx.db.Exec(ctx, sqlf.Sprintf(reindexUploadsQuery, sqlf.Join(conds, " AND ")))
	if err != nil {
		return err
	}

	return nil
}

const reindexUploadsQuery = `
WITH
upload_candidates AS (
    SELECT u.id, u.associated_index_id
	FROM lsif_uploads u
	JOIN repo ON repo.id = u.repository_id
	WHERE %s
    ORDER BY u.id
    FOR UPDATE
),
update_uploads AS (
	UPDATE lsif_uploads u
	SET should_reindex = true
	WHERE u.id IN (SELECT id FROM upload_candidates)
),
index_candidates AS (
	SELECT u.id
	FROM lsif_indexes u
	WHERE u.id IN (SELECT associated_index_id FROM upload_candidates)
	ORDER BY u.id
	FOR UPDATE
)
UPDATE lsif_indexes u
SET should_reindex = true
WHERE u.id IN (SELECT id FROM index_candidates)
`

// ReindexUploadByID reindexes an upload by its identifier.
func (s *store) ReindexUploadByID(ctx context.Context, id int) (err error) {
	ctx, _, endObservation := s.operations.reindexUploadByID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("id", id),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.db.Done(err) }()

	return tx.db.Exec(ctx, sqlf.Sprintf(reindexUploadByIDQuery, id, id))
}

const reindexUploadByIDQuery = `
WITH
update_uploads AS (
	UPDATE lsif_uploads u
	SET should_reindex = true
	WHERE id = %s
)
UPDATE lsif_indexes u
SET should_reindex = true
WHERE id IN (SELECT associated_index_id FROM lsif_uploads WHERE id = %s)
`
