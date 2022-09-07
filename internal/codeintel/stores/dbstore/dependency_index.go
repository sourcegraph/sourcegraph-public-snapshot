package dbstore

import (
	"context"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

func scanDependencySyncingJob(s dbutil.Scanner) (job DependencySyncingJob, err error) {
	return job, s.Scan(
		&job.ID,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.UploadID,
	)
}

var dependencySyncingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("lsif_dependency_syncing_jobs.id"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.state"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.failure_message"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.started_at"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.finished_at"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.process_after"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.num_resets"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.num_failures"),
	sqlf.Sprintf("lsif_dependency_syncing_jobs.upload_id"),
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

func scanDependencyIndexingJob(s dbutil.Scanner) (job DependencyIndexingJob, err error) {
	return job, s.Scan(
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
	)
}

var dependencyIndexingJobColumns = []*sqlf.Query{
	sqlf.Sprintf("lsif_dependency_indexing_jobs.id"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.state"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.failure_message"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.started_at"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.finished_at"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.process_after"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.num_resets"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.num_failures"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.upload_id"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.external_service_kind"),
	sqlf.Sprintf("lsif_dependency_indexing_jobs.external_service_sync"),
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
-- source: internal/codeintel/stores/dbstore/dependency_index.go:InsertDependencySyncingJob
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
-- source: internal/codeintel/stores/dbstore/dependency_index.go:InsertCloneableDependencyRepo
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
-- source: internal/codeintel/stores/dbstore/dependency_index.go:InsertDependencyIndexingJob
INSERT INTO lsif_dependency_indexing_jobs (upload_id, external_service_kind, external_service_sync)
VALUES (%s, %s, %s)
RETURNING id
`
