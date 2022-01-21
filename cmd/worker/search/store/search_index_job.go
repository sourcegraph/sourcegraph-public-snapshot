package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/search/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var searchIndexJobColumns = []*sqlf.Query{
	sqlf.Sprintf("search_index_jobs.id"),
	sqlf.Sprintf("search_index_jobs.repo_id"),
	sqlf.Sprintf("search_index_jobs.revision"),
}

func (s *Store) GetSearchIndexJob(ctx context.Context, id int) (_ *types.SearchIndexJob, err error) {
	ctx, endObservation := s.operations.getSearchIndexJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", id),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.Store.Query(ctx, sqlf.Sprintf(getSearchIndexJobQueryFmtstr, sqlf.Join(searchIndexJobColumns, ", \n"), id))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	job, found, err := scanFirstSearchIndexJob(rows)
	if err != nil {
		return nil, err
	}

	if !found {
		return job, errors.New("not found")
	}

	return job, nil
}

const getSearchIndexJobQueryFmtstr = `
-- source: cmd/worker/search/store/search_index_job.go:GetSearchIndexJob
SELECT
	%s
FROM
	search_index_jobs
WHERE id = %s
`

func (s *Store) CreateSearchIndexJob(ctx context.Context, job *types.SearchIndexJob) (err error) {
	ctx, endObservation := s.operations.createSearchIndexJob.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	rows, err := s.Store.Query(ctx, sqlf.Sprintf(createSearchIndexJobQueryFmtstr, job.RepoID, job.Revision, sqlf.Join(searchIndexJobColumns, ", \n")))
	if err != nil {
		return err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := scanSearchIndexJob(rows, job); err != nil {
			return err
		}
	}

	return nil
}

const createSearchIndexJobQueryFmtstr = `
-- source: cmd/worker/search/store/search_index_job.go:CreateSearchIndexJob
INSERT INTO
	search_index_jobs
(repo_id, revision)
VALUES
(%s, %s)
RETURNING
%s
`

func scanFirstSearchIndexJob(rows *sql.Rows) (*types.SearchIndexJob, bool, error) {
	jobs, err := scanSearchIndexJobs(rows)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}
	return jobs[0], true, nil
}

func scanSearchIndexJobs(rows *sql.Rows) ([]*types.SearchIndexJob, error) {
	var jobs []*types.SearchIndexJob

	for rows.Next() {
		var job *types.SearchIndexJob
		if err := scanSearchIndexJob(rows, job); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func scanSearchIndexJob(sc dbutil.Scanner, j *types.SearchIndexJob) error {
	return sc.Scan(
		&j.ID,
		&j.RepoID,
		&j.Revision,
	)
}
