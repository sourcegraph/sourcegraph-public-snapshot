package dbstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// DependencySyncingJob is a subset of the lsif_dependency_syncing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type DependencySyncingJob struct {
	ID             int        `json:"id"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFailures    int        `json:"numFailures"`
	UploadID       int        `json:"uploadId"`
}

func (u DependencySyncingJob) RecordID() int {
	return u.ID
}

// scanDependencySyncingJobs scans a slice of dependency syncing jobs from the return value of
// `*Store.query`.
func scanDependencySyncingJobs(rows *sql.Rows, queryErr error) (_ []DependencySyncingJob, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var jobs []DependencySyncingJob
	for rows.Next() {
		var job DependencySyncingJob
		if err := rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			&job.UploadID,
		); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

var dependencySyncingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("j.id"),
	sqlf.Sprintf("j.state"),
	sqlf.Sprintf("j.failure_message"),
	sqlf.Sprintf("j.started_at"),
	sqlf.Sprintf("j.finished_at"),
	sqlf.Sprintf("j.process_after"),
	sqlf.Sprintf("j.num_resets"),
	sqlf.Sprintf("j.num_failures"),
	sqlf.Sprintf("j.upload_id"),
}

// scanFirstDependencySyncingingJob scans a slice of dependency indexing jobs from the return
// value of `*Store.query` and returns the first.
func scanFirstDependencySyncingingJob(rows *sql.Rows, err error) (DependencySyncingJob, bool, error) {
	jobs, err := scanDependencySyncingJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return DependencySyncingJob{}, false, err
	}
	return jobs[0], true, nil
}

// scanFirstDependencySyncingJobRecord scans a slice of dependency indexing jobs from the
// return value of `*Store.query` and returns the first.
func scanFirstDependencySyncingJobRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return scanFirstDependencySyncingingJob(rows, err)
}

// DependencyIndexingJob is a subset of the lsif_dependency_indexing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type DependencyIndexingJob struct {
	ID                  int        `json:"id"`
	State               string     `json:"state"`
	FailureMessage      *string    `json:"failureMessage"`
	StartedAt           *time.Time `json:"startedAt"`
	FinishedAt          *time.Time `json:"finishedAt"`
	ProcessAfter        *time.Time `json:"processAfter"`
	NumResets           int        `json:"numResets"`
	NumFailures         int        `json:"numFailures"`
	UploadID            int        `json:"uploadId"`
	ExternalServiceKind string     `json:"externalServiceKind"`
	ExternalServiceSync time.Time  `json:"externalServiceSync"`
}

func (u DependencyIndexingJob) RecordID() int {
	return u.ID
}

// scanDependencyIndexingJobs scans a slice of dependency indexing jobs from the return value of
// `*Store.query`.
func scanDependencyIndexingJobs(rows *sql.Rows, queryErr error) (_ []DependencyIndexingJob, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var jobs []DependencyIndexingJob
	for rows.Next() {
		var job DependencyIndexingJob
		if err := rows.Scan(
			&job.ID,
			&job.State,
			&job.FailureMessage,
			&job.StartedAt,
			&job.FinishedAt,
			&job.ProcessAfter,
			&job.NumResets,
			&job.NumFailures,
			&job.UploadID,
			&job.ExternalServiceKind,
			&job.ExternalServiceSync,
		); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

var dependencyIndexingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("j.id"),
	sqlf.Sprintf("j.state"),
	sqlf.Sprintf("j.failure_message"),
	sqlf.Sprintf("j.started_at"),
	sqlf.Sprintf("j.finished_at"),
	sqlf.Sprintf("j.process_after"),
	sqlf.Sprintf("j.num_resets"),
	sqlf.Sprintf("j.num_failures"),
	sqlf.Sprintf("j.upload_id"),
	sqlf.Sprintf("j.external_service_kind"),
	sqlf.Sprintf("j.external_service_sync"),
}

// scanFirstDependencyIndexingJob scans a slice of dependency indexing jobs from the return
// value of `*Store.query` and returns the first.
func scanFirstDependencyIndexingJob(rows *sql.Rows, err error) (DependencyIndexingJob, bool, error) {
	jobs, err := scanDependencyIndexingJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return DependencyIndexingJob{}, false, err
	}
	return jobs[0], true, nil
}

// scanFirstDependencyIndexingJobRecord scans a slice of dependency indexing jobs from the
// return value of `*Store.query` and returns the first.
func scanFirstDependencyIndexingJobRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return scanFirstDependencyIndexingJob(rows, err)
}

// InsertDependencySyncingJob inserts a new dependency syncing job and returns its identifier.
func (s *Store) InsertDependencySyncingJob(ctx context.Context, uploadID int) (id int, err error) {
	ctx, _, endObservation := s.operations.insertDependencySyncingJob.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(insertDependencySyncingJobQuery, uploadID)))
	return id, err
}

const insertDependencySyncingJobQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/dependency_index.go:InsertDependencySyncingJob
INSERT INTO lsif_dependency_syncing_jobs (upload_id) VALUES (%s)
RETURNING id
`

func (s *Store) InsertCloneableDependencyRepo(ctx context.Context, dependency precise.Package) (new bool, err error) {
	ctx, _, endObservation := s.operations.insertCloneableDependencyRepo.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Bool("new", new),
			log.Object("dependency", fmt.Sprint(dependency)),
		}})
	}()

	_, new, err = basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(insertCloneableDependencyRepoQuery, dependency.Scheme, dependency.Name, dependency.Version)))
	return
}

const insertCloneableDependencyRepoQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/dependency_index.go:InsertCloneableDependencyRepo
INSERT INTO lsif_dependency_repos (scheme, name, version)
VALUES (%s, %s, %s)
ON CONFLICT DO NOTHING
RETURNING 1
`

func (s *Store) InsertDependencyIndexingJob(ctx context.Context, uploadID int, externalServiceKind string, syncTime time.Time) (id int, err error) {
	ctx, _, endObservation := s.operations.insertDependencyIndexingJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadId", uploadID),
		log.String("extSvcKind", externalServiceKind),
	}})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("id", id),
		}})
	}()

	id, _, err = basestore.ScanFirstInt(s.Store.Query(ctx, sqlf.Sprintf(insertDependencyIndexingJobQuery, uploadID, externalServiceKind, syncTime)))
	return id, err
}

const insertDependencyIndexingJobQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/dependency_index.go:InsertDependencyIndexingJob
INSERT INTO lsif_dependency_indexing_jobs (upload_id, external_service_kind, external_service_sync)
VALUES (%s, %s, %s)
RETURNING id
`
